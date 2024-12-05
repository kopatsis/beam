package draftorderhelp

import (
	"beam/config"
	"beam/data/models"
	"time"
)

func UpdateShippingRates(draft *models.DraftOrder, newContact models.OrderContact, mutexes *config.AllMutexes, name string) {
	address := newContact.StreetAddress1 + ", " + newContact.City + ", " + newContact.ProvinceState + ", " + newContact.ZipCode + ", " + newContact.Country
	now := time.Now()

	if rates, exists := draft.AllShippingRates[address]; exists && len(rates) > 1 {
		if time.Since(rates[0].Timestamp) < time.Hour {
			draft.CurrentShipping = rates
			draft.ActualRate = rates[0]
			return
		}
	}

	newRates := getApiShipRates(draft, newContact, mutexes, name)
	for i := range newRates {
		newRates[i].Timestamp = now
	}

	draft.AllShippingRates[address] = newRates
	if len(newRates) > 0 {
		draft.CurrentShipping = newRates
		draft.ActualRate = newRates[0]
	}
}

func getApiShipRates(draft *models.DraftOrder, newContact models.OrderContact, mutexes *config.AllMutexes, name string) []models.ShippingRate {
	return []models.ShippingRate{}
}
