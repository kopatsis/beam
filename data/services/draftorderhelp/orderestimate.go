package draftorderhelp

import (
	"beam/background/apidata"
	"beam/background/emails"
	"beam/config"
	"beam/data/models"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

var (
	estimateLimiters = make(map[string]*rate.Limiter)
	ipEstLimiters    = make(map[string]*rate.Limiter)
	estMu            sync.Mutex
	ipeMu            sync.Mutex
)

func getEstimateLimiter(storeName string) *rate.Limiter {

	estMu.Lock()
	defer estMu.Unlock()

	limiter, exists := estimateLimiters[storeName]
	if !exists {
		limiter = rate.NewLimiter(rate.Every(time.Duration(int64(config.SHIPINTERVAL)/int64(config.ESTREQS))), config.ESTREQS)
		estimateLimiters[storeName] = limiter
	}

	return limiter
}

func getIPEstLimiter(ip string) *rate.Limiter {
	ipeMu.Lock()
	defer ipeMu.Unlock()

	limiter, exists := ipEstLimiters[ip]
	if !exists {
		limiter = rate.NewLimiter(rate.Every(time.Duration(int64(config.SHIPINTERVAL)/int64(config.IPEREQS))), config.IPEREQS)
		ipEstLimiters[ip] = limiter
	}

	return limiter
}

func applyEstimateRateLimit(storeName string, tools *config.Tools) error {
	limiter := getEstimateLimiter(storeName)
	ctx, cancel := context.WithTimeout(context.Background(), 9*time.Second)
	defer cancel()

	startTime := time.Now()

	err := limiter.Wait(ctx)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Printf("Warning: Store %s wait time exceeded 3 seconds: %v", storeName, 9*time.Second)
			go emails.AlertEmailEstRateDanger(storeName, 9*time.Second, tools, true)
			return fmt.Errorf("rate limit exceeded for %s, timeout after 10 seconds", storeName)
		}
		return fmt.Errorf("failed to wait for rate limit: %w", err)
	}

	waitDuration := time.Since(startTime)

	if waitDuration > 6*time.Second {
		go emails.AlertEmailEstRateDanger(storeName, waitDuration, tools, false)
	}

	if waitDuration > 3*time.Second {
		log.Printf("Warning: Store %s wait time exceeded 3 seconds: %v", storeName, waitDuration)
	}

	return nil
}

func applyIPEstRateLimit(ip string, tools *config.Tools) error {
	limiter := getIPEstLimiter(ip)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	startTime := time.Now()

	err := limiter.Wait(ctx)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Printf("Warning: IP %s wait time exceeded 3 seconds: %v", ip, 10*time.Second)
			go emails.AlertIPEstRateDanger(ip, 9*time.Second, tools, true)
			return fmt.Errorf("rate limit exceeded for %s, timeout after 10 seconds", ip)
		}
		return fmt.Errorf("failed to wait for rate limit: %w", err)
	}

	waitDuration := time.Since(startTime)

	if waitDuration > 6*time.Second {
		go emails.AlertIPEstRateDanger(ip, waitDuration, tools, false)
	}

	if waitDuration > 3*time.Second {
		log.Printf("Warning: IP %s wait time exceeded 3 seconds: %v", ip, waitDuration)
	}

	return nil
}

func applyEstRateLimitsConcurrently(storeName, ip string, tools *config.Tools) error {
	var storeErr, ipErr error
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		storeErr = applyEstimateRateLimit(storeName, tools)
	}()

	go func() {
		defer wg.Done()
		ipErr = applyIPEstRateLimit(ip, tools)
	}()

	wg.Wait()

	if ipErr != nil {
		return errors.New("did not complete based on ip limiting")
	} else if storeErr != nil {
		return errors.New("did not complete based on store limiting")
	}
	return nil
}

func getEstApiShipRates(draft *models.DraftOrder, newContact *models.Contact, mutexes *config.AllMutexes, name, ip, rateName string, tools *config.Tools) (*models.OrderEstimateCost, error) {

	if rateName == "" {
		rateName = "STANDARD"
	}

	mutexes.Api.Mu.RLock()
	apiKey := mutexes.Api.KeyMap[name]
	mutexes.Api.Mu.RUnlock()

	lineItems, err := createEstItemsArray(draft.Lines)
	if err != nil {
		return nil, err
	}

	reqBody := apidata.ToCostEstimate{
		Shipping: rateName,
		Recipient: apidata.Recipient{
			Address1:    newContact.StreetAddress1,
			City:        newContact.City,
			CountryCode: newContact.Country,
			StateCode:   newContact.ProvinceState,
			Zip:         newContact.ZipCode,
		},
		Items:    lineItems,
		Currency: "USD",
		Locale:   "en_us",
	}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	base := os.Getenv("PF_URL")
	req, err := http.NewRequest("POST", base+"/orders/estimate-costs", bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	if err := applyEstRateLimitsConcurrently(name, ip, tools); err != nil {
		return nil, err
	}

	resp, err := tools.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	var apiResponse apidata.FromCostEstimate
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, err
	}

	ret := apiResponse.Result.Costs
	ret.Timestamp = time.Now()
	ret.ShipRate = rateName

	return &ret, nil
}

func createEstItemsArray(orderLines []models.OrderLine) ([]apidata.Items, error) {
	variantMap := map[string]int{}
	pfOGMap := map[string]models.OriginalProductRedis{}

	for _, line := range orderLines {
		for _, pf := range line.PrintfulID {
			variantMap[pf.VariantID] += pf.Quantity * line.Quantity
			copy := pf
			pfOGMap[pf.VariantID] = copy
		}
	}

	items := make([]apidata.Items, 0, len(variantMap))
	i := 0

	for variantID, quantity := range variantMap {
		i++
		if pfProd, exists := pfOGMap[variantID]; exists {
			syncVarID, err := strconv.Atoi(pfProd.VariantID)
			if err != nil {
				return nil, fmt.Errorf("Error converting variant ID to int: %v", err)
			}

			originalProdID, err := strconv.Atoi(pfProd.OriginalProductID)
			if err != nil {
				return nil, fmt.Errorf("Error converting og product ID to int: %v", err)
			}

			originalVarID, err := strconv.Atoi(pfProd.OriginalVariantID)
			if err != nil {
				return nil, fmt.Errorf("Error converting og variant ID to int: %v", err)
			}

			items = append(items, apidata.Items{
				ID:                i,
				ExternalID:        fmt.Sprintf("LINE_ITEM_%d", i),
				ExternalVariantID: pfProd.ExternalVariantID,
				VariantID:         originalVarID,
				SyncVariantID:     int64(syncVarID),
				Quantity:          quantity,
				Product: apidata.Product{
					ProductID: originalProdID,
					VariantID: originalVarID,
				},
			})
		}
	}

	return items, nil
}

func DraftOrderEstimateUpdate(draftOrder *models.DraftOrder, newContact *models.Contact, mutexes *config.AllMutexes, name, ip, shipRate string, tools *config.Tools) error {
	if shipRate == "" {
		shipRate = "STANDARD"
	}

	address := shipRate + " :: " + newContact.StreetAddress1 + ", " + newContact.City + ", " + newContact.ProvinceState + ", " + newContact.ZipCode + ", " + newContact.Country

	if est, ok := draftOrder.AllOrderEstimates[address]; ok && time.Since(est.Timestamp) <= 1*time.Hour {
		draftOrder.OrderEstimate = est
	} else {
		if newEst, err := getEstApiShipRates(draftOrder, newContact, mutexes, name, ip, shipRate, tools); err != nil {
			return err
		} else {
			newEst.AddressFormat = address
			draftOrder.AllOrderEstimates[address] = *newEst
			draftOrder.OrderEstimate = *newEst
		}
	}

	return CompareCostsOfDraftOrder(draftOrder, name, tools)
}

func CompareCostsOfDraftOrder(draftOrder *models.DraftOrder, name string, tools *config.Tools) error {
	if draftOrder.OrderEstimate.Total <= 0 {
		return errors.New("no pending order estiamte")
	}

	if draftOrder.PreGiftCardTotal <= 0 {
		return errors.New("no pending order price")
	}

	cost := int(math.Round(draftOrder.OrderEstimate.Total * 100))
	price := draftOrder.PreGiftCardTotal

	if draftOrder.CATax {
		price -= draftOrder.Tax
	}

	if cost >= price {
		go emails.AlertEstimateTooHigh(name, draftOrder.ID.Hex(), tools, true, cost, price)
	} else if cost*6/5 >= price {
		go emails.AlertEstimateTooHigh(name, draftOrder.ID.Hex(), tools, false, cost, price)
	}

	return nil
}
