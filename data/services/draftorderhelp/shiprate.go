package draftorderhelp

import (
	"beam/background/emails"
	"beam/config"
	"beam/data/models"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"strconv"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

var (
	storeLimiters = make(map[string]*rate.Limiter)
	ipLimiters    = make(map[string]*rate.Limiter)
	stMu          sync.Mutex
	ipMu          sync.Mutex
)

func getStoreLimiter(storeName string) *rate.Limiter {

	stMu.Lock()
	defer stMu.Unlock()

	limiter, exists := storeLimiters[storeName]
	if !exists {
		limiter = rate.NewLimiter(rate.Every(time.Duration(int64(config.SHIPINTERVAL)/int64(config.SHIPREQS))), config.SHIPREQS)
		storeLimiters[storeName] = limiter
	}

	return limiter
}

func getIPLimiter(ip string) *rate.Limiter {
	ipMu.Lock()
	defer ipMu.Unlock()

	limiter, exists := ipLimiters[ip]
	if !exists {
		limiter = rate.NewLimiter(rate.Every(time.Duration(int64(config.SHIPINTERVAL)/int64(config.IPREQS))), config.IPREQS)
		ipLimiters[ip] = limiter
	}

	return limiter
}

func applyStoreRateLimit(storeName string, tools *config.Tools) error {
	limiter := getStoreLimiter(storeName)
	ctx, cancel := context.WithTimeout(context.Background(), 9*time.Second)
	defer cancel()

	startTime := time.Now()

	err := limiter.Wait(ctx)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Printf("Warning: Store %s wait time exceeded 3 seconds: %v", storeName, 9*time.Second)
			go emails.AlertEmailRateDanger(storeName, 9*time.Second, tools, true)
			return fmt.Errorf("rate limit exceeded for %s, timeout after 10 seconds", storeName)
		}
		return fmt.Errorf("failed to wait for rate limit: %w", err)
	}

	waitDuration := time.Since(startTime)

	if waitDuration > 6*time.Second {
		go emails.AlertEmailRateDanger(storeName, waitDuration, tools, false)
	}

	if waitDuration > 3*time.Second {
		log.Printf("Warning: Store %s wait time exceeded 3 seconds: %v", storeName, waitDuration)
	}

	return nil
}

func applyIpRateLimit(ip string, tools *config.Tools) error {
	limiter := getIPLimiter(ip)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	startTime := time.Now()

	err := limiter.Wait(ctx)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Printf("Warning: IP %s wait time exceeded 3 seconds: %v", ip, 10*time.Second)
			go emails.AlertIPRateDanger(ip, 9*time.Second, tools, true)
			return fmt.Errorf("rate limit exceeded for %s, timeout after 10 seconds", ip)
		}
		return fmt.Errorf("failed to wait for rate limit: %w", err)
	}

	waitDuration := time.Since(startTime)

	if waitDuration > 6*time.Second {
		go emails.AlertIPRateDanger(ip, waitDuration, tools, false)
	}

	if waitDuration > 3*time.Second {
		log.Printf("Warning: IP %s wait time exceeded 3 seconds: %v", ip, waitDuration)
	}

	return nil
}

func applyRateLimitsConcurrently(storeName, ip string, tools *config.Tools) error {
	var storeErr, ipErr error
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		storeErr = applyStoreRateLimit(storeName, tools)
	}()

	go func() {
		defer wg.Done()
		ipErr = applyIpRateLimit(ip, tools)
	}()

	wg.Wait()

	if ipErr != nil {
		return errors.New("did not complete based on ip limiting")
	} else if storeErr != nil {
		return errors.New("did not complete based on store limiting")
	}
	return nil
}

func UpdateShippingRates(draft *models.DraftOrder, newContact *models.Contact, mutexes *config.AllMutexes, name, ip string, tools *config.Tools) error {
	address := newContact.StreetAddress1 + ", " + newContact.City + ", " + newContact.ProvinceState + ", " + newContact.ZipCode + ", " + newContact.Country

	freeship := slices.Contains(draft.Tags, "FREESHIP_O")

	if rates, exists := draft.AllShippingRates[address]; exists && len(rates) > 1 {
		if time.Since(rates[0].Timestamp) < time.Hour {
			draft.CurrentShipping = rates
			return UpdateActualShippingRate(draft, rates[0].ID)
		}
	}

	newRates, err := getApiShipRates(draft, newContact, mutexes, name, ip, freeship, tools)
	if err != nil {
		return err
	}
	draft.AllShippingRates[address] = newRates
	draft.CurrentShipping = newRates

	return UpdateActualShippingRate(draft, newRates[0].ID)
}

func UpdateActualShippingRate(order *models.DraftOrder, shipID string) error {
	var selectedRate *models.ShippingRate

	for _, rate := range order.CurrentShipping {
		if rate.ID == shipID {
			selectedRate = &rate
			break
		}
	}

	if selectedRate == nil {
		return errors.New("shipping rate not found")
	}

	if time.Since(selectedRate.Timestamp) > time.Hour {
		return errors.New("shipping rate has expired")
	}

	order.ActualRate = *selectedRate

	rateInt, err := convertRateToCents(selectedRate.Rate)
	if err != nil {
		return err
	}

	if rateInt != order.Shipping {
		order.Shipping = rateInt
		order.Total += (rateInt - order.Shipping)

		if err := EnsureGiftCardSum(order, 0, order.Total, false); err != nil {
			return err
		}

		return updateStripePaymentIntent(order.StripePaymentIntentID, order.Total)
	}

	return nil
}

func convertRateToCents(rate string) (int, error) {
	var rateInt int
	_, err := fmt.Sscanf(rate, "%f", &rateInt)
	if err != nil {
		return 0, fmt.Errorf("invalid rate format: %v", err)
	}
	return rateInt, nil
}

func getApiShipRates(draft *models.DraftOrder, newContact *models.Contact, mutexes *config.AllMutexes, name, ip string, freeship bool, tools *config.Tools) ([]models.ShippingRate, error) {

	mutexes.Api.Mu.RLock()
	apiKey := mutexes.Api.KeyMap[name]
	mutexes.Api.Mu.RUnlock()

	reqBody := map[string]interface{}{
		"recipient": map[string]interface{}{
			"address1":     newContact.StreetAddress1,
			"city":         newContact.City,
			"country_code": newContact.Country,
			"state_code":   newContact.ProvinceState,
			"zip":          newContact.ZipCode,
		},
		"items":    createItemsArray(draft.Lines),
		"currency": "USD",
		"locale":   "en_US",
	}
	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	base := os.Getenv("PF_URL")
	req, err := http.NewRequest("POST", base+"/shipping/rates", bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	if err := applyRateLimitsConcurrently(name, ip, tools); err != nil {
		return []models.ShippingRate{}, err
	}

	resp, err := tools.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	var apiResponse struct {
		Code   int                   `json:"code"`
		Result []models.ShippingRate `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, err
	}

	if len(apiResponse.Result) == 0 {
		return nil, errors.New("no shipping options available")
	}

	now := time.Now()
	newRates := apiResponse.Result
	for i := range newRates {
		newRates[i].Timestamp = now
		cents, err := convertRateToCents(newRates[i].Rate)
		if err != nil {
			return nil, err
		}
		newRates[i].CentsRate = cents
	}

	if freeship {
		cheapest := newRates[0].CentsRate
		for _, r := range newRates {
			if r.CentsRate < cheapest {
				cheapest = r.CentsRate
			}
		}
		for i := range newRates {
			newRates[i].CentsRate = newRates[i].CentsRate - cheapest
		}
	}

	return newRates, nil
}

func createItemsArray(orderLines []models.OrderLine) []map[string]any {
	variantMap := map[string]int{}
	pfOGMap := map[string]models.OriginalProductRedis{}

	for _, line := range orderLines {
		for _, pf := range line.PrintfulID {
			variantMap[pf.VariantID] += pf.Quantity * line.Quantity
			copy := pf
			pfOGMap[pf.VariantID] = copy
		}
	}

	items := make([]map[string]any, 0, len(variantMap))

	for variantID, quantity := range variantMap {
		if pfProd, exists := pfOGMap[variantID]; exists {
			items = append(items, map[string]any{
				"variant_id":          variantID,
				"external_variant_id": pfProd.ExternalVariantID,
				"quantity":            quantity,
			})
		}
	}

	return items
}

func EvaluateFreeShip(draftOrder *models.DraftOrder, customer *models.Customer, products map[int]*models.ProductRedis) bool {
	freeShipSubtotal, err := strconv.Atoi(os.Getenv("FREESHIP_SUBTOTAL"))
	if err != nil {
		return false
	}
	if (customer != nil && slices.Contains(customer.Tags, "FREESHIP")) || draftOrder.Subtotal >= freeShipSubtotal {
		allAllowed := true
		for _, l := range draftOrder.Lines {
			pid, err := strconv.Atoi(l.ProductID)
			if err != nil {
				allAllowed = false
				break
			}
			p, ok := products[pid]
			if !ok {
				allAllowed = false
				break
			}
			if slices.Contains(p.Tags, "NOFREESHIP") {
				allAllowed = false
				break
			}
		}
		if allAllowed {
			draftOrder.Tags = append(draftOrder.Tags, "FREESHIP_O")
		}
		return allAllowed
	}
	return false
}
