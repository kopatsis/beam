package draftorderhelp

import (
	"beam/config"
	"beam/data/models"
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"time"
)

func UpdateShippingRates(draft *models.DraftOrder, newContact models.OrderContact, mutexes *config.AllMutexes, name string, client *http.Client) {
	address := newContact.StreetAddress1 + ", " + newContact.City + ", " + newContact.ProvinceState + ", " + newContact.ZipCode + ", " + newContact.Country
	now := time.Now()

	if rates, exists := draft.AllShippingRates[address]; exists && len(rates) > 1 {
		if time.Since(rates[0].Timestamp) < time.Hour {
			draft.CurrentShipping = rates
			draft.ActualRate = rates[0]
			return
		}
	}

	newRates := getApiShipRates(draft, newContact, mutexes, name, client)
	for i := range newRates {
		newRates[i].Timestamp = now
	}

	draft.AllShippingRates[address] = newRates
	if len(newRates) > 0 {
		draft.CurrentShipping = newRates
		draft.ActualRate = newRates[0]
	}
}

func getApiShipRates(draft *models.DraftOrder, newContact models.OrderContact, mutexes *config.AllMutexes, name string, client *http.Client) []models.ShippingRate {
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

	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var apiResponse struct {
		Code   int                   `json:"code"`
		Result []models.ShippingRate `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil
	}

	return apiResponse.Result
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
