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
		go emails.AlertEmailRateDanger(ip, waitDuration, tools, false)
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

func UpdateShippingRates(draft *models.DraftOrder, newContact models.OrderContact, mutexes *config.AllMutexes, name, ip string, tools *config.Tools) error {
	address := newContact.StreetAddress1 + ", " + newContact.City + ", " + newContact.ProvinceState + ", " + newContact.ZipCode + ", " + newContact.Country
	now := time.Now()

	if rates, exists := draft.AllShippingRates[address]; exists && len(rates) > 1 {
		if time.Since(rates[0].Timestamp) < time.Hour {
			draft.CurrentShipping = rates
			draft.ActualRate = rates[0]
			return nil
		}
	}

	newRates, err := getApiShipRates(draft, newContact, mutexes, name, ip, tools)
	if err != nil {
		return err
	}

	for i := range newRates {
		newRates[i].Timestamp = now
	}

	draft.AllShippingRates[address] = newRates
	if len(newRates) > 0 {
		draft.CurrentShipping = newRates
		draft.ActualRate = newRates[0]
	}

	return nil
}

func getApiShipRates(draft *models.DraftOrder, newContact models.OrderContact, mutexes *config.AllMutexes, name, ip string, tools *config.Tools) ([]models.ShippingRate, error) {

	if err := applyRateLimitsConcurrently(name, ip, tools); err != nil {
		return []models.ShippingRate{}, err
	}

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
		"items":    createItemsArray(draft.Lines, mutexes, name),
		"currency": "USD",
		"locale":   "en_US",
	}
	reqBodyBytes, _ := json.Marshal(reqBody)

	base := os.Getenv("PF_URL")
	req, _ := http.NewRequest("POST", base+"/shipping/rates", bytes.NewBuffer(reqBodyBytes))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

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

	return apiResponse.Result, nil
}

func createItemsArray(orderLines []models.OrderLine, mutexes *config.AllMutexes, name string) []map[string]any {
	variantMap := make(map[string]int)

	for _, line := range orderLines {
		for variantID, count := range line.PrintfulID {
			variantMap[variantID] += count * line.Quantity
		}
	}

	mutexes.External.Mu.RLock()
	defer mutexes.External.Mu.RUnlock()
	items := make([]map[string]any, 0, len(variantMap))

	for variantID, quantity := range variantMap {
		externalKey := name + "::" + variantID
		if externalVariantID, exists := mutexes.External.IDMap[externalKey]; exists {
			items = append(items, map[string]any{
				"variant_id":          variantID,
				"external_variant_id": externalVariantID,
				"quantity":            quantity,
			})
		}
	}

	return items
}
