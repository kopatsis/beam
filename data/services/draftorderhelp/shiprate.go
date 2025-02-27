package draftorderhelp

import (
	"beam/config"
	"beam/data/models"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"slices"
	"strconv"
	"time"
)

func applyRateLimitsShip(storeName, ip string, tools *config.Tools) error {

	if unmaxed, err := config.RateLimit(tools.Redis, storeName, "SHP", ip, config.IPREQS, time.Minute); err != nil {
		return err
	} else if !unmaxed {
		return errors.New("maximum usage reached: by IP")
	}

	if unmaxed, err := config.RateLimit(tools.Redis, storeName, "SHP", "", config.SHIPREQS, time.Minute); err != nil {
		return err
	} else if !unmaxed {
		return errors.New("maximum usage reached: by store")
	}

	return nil
}

func UpdateShippingRates(draft *models.DraftOrder, newContact *models.Contact, mutexes *config.AllMutexes, name, ip string, tools *config.Tools) error {
	address := newContact.StreetAddress1 + ", " + newContact.City + ", " + newContact.ProvinceState + ", " + newContact.ZipCode + ", " + newContact.Country

	freeship := slices.Contains(draft.Tags, "FREESHIP_O")

	currentRateID := draft.ActualRate.ID

	if rates, exists := draft.AllShippingRates[address]; exists && len(rates) > 1 {
		if time.Since(rates[0].Timestamp) < time.Hour {
			draft.CurrentShipping = rates
			return UpdateActualShippingRate(draft, currentRateID)
		}
	}

	newRates, err := getApiShipRates(draft, newContact, mutexes, name, ip, freeship, tools)
	if err != nil {
		return err
	}
	draft.AllShippingRates[address] = newRates
	draft.CurrentShipping = newRates

	return UpdateActualShippingRate(draft, currentRateID)
}

func UpdateActualShippingRate(order *models.DraftOrder, shipID string) error {
	var selectedRate *models.ShippingRate

	if shipID != "" {
		for _, rate := range order.CurrentShipping {
			if rate.ID == shipID {
				selectedRate = &rate
				break
			}
		}
	}

	if selectedRate == nil {
		if len(order.CurrentShipping) == 0 {
			return errors.New("no ship rate to find")
		}
		selectedRate = &order.CurrentShipping[0]
	}

	if time.Since(selectedRate.Timestamp) > time.Hour {
		return errors.New("shipping rate has expired")
	}

	order.ActualRate = *selectedRate

	rateInt, err := convertRateToCents(selectedRate.Rate)
	if err != nil {
		return err
	}

	order.Shipping = rateInt

	checkDays := selectedRate.MinDeliveryDays
	if checkDays <= 0 {
		checkDays = 14
	} else {
		checkDays += 3
	}

	order.CheckDeliveryDate = time.Now().AddDate(0, 0, checkDays)

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

	if err := applyRateLimitsShip(name, ip, tools); err != nil {
		return []models.ShippingRate{}, err
	}

	resp, err := tools.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error with response: http status: %d", resp.StatusCode)
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
			p, ok := products[l.ProductID]
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
