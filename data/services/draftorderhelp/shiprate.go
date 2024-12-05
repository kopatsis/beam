package draftorderhelp

import (
	"beam/background/emails"
	"beam/config"
	"beam/data/models"
	"bytes"
	"context"
	"encoding/json"
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
	mu            sync.Mutex
)

func getLimiter(storeName string) *rate.Limiter {

	mu.Lock()
	defer mu.Unlock()

	limiter, exists := storeLimiters[storeName]
	if !exists {
		limiter = rate.NewLimiter(rate.Every(time.Duration(int64(config.SHIPINTERVAL)/int64(config.SHIPREQS))), config.SHIPREQS)
		storeLimiters[storeName] = limiter
	}

	return limiter
}

func applyRateLimit(storeName string, tools *config.Tools) error {
	limiter := getLimiter(storeName)
	startTime := time.Now()

	err := limiter.Wait(context.Background())
	if err != nil {
		return fmt.Errorf("failed to wait for rate limit: %w", err)
	}

	waitDuration := time.Since(startTime)

	if waitDuration > 10*time.Second {
		return fmt.Errorf("rate limit exceeded for %s, wait time too long: %v", storeName, waitDuration)
	}

	if waitDuration > 6*time.Second {
		emails.AlertEmailRateDanger(storeName, waitDuration, tools)
	}

	if waitDuration > 3*time.Second {
		log.Printf("Warning: Store %s wait time exceeded 3 seconds: %v", storeName, waitDuration)
	}

	return nil
}

func UpdateShippingRates(draft *models.DraftOrder, newContact models.OrderContact, mutexes *config.AllMutexes, name string, tools *config.Tools) error {
	address := newContact.StreetAddress1 + ", " + newContact.City + ", " + newContact.ProvinceState + ", " + newContact.ZipCode + ", " + newContact.Country
	now := time.Now()

	if rates, exists := draft.AllShippingRates[address]; exists && len(rates) > 1 {
		if time.Since(rates[0].Timestamp) < time.Hour {
			draft.CurrentShipping = rates
			draft.ActualRate = rates[0]
			return nil
		}
	}

	newRates, err := getApiShipRates(draft, newContact, mutexes, name, tools)
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

func getApiShipRates(draft *models.DraftOrder, newContact models.OrderContact, mutexes *config.AllMutexes, name string, tools *config.Tools) ([]models.ShippingRate, error) {

	if err := applyRateLimit(name, tools); err != nil {
		return []models.ShippingRate{}, nil
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
