package draftorderhelp

import (
	"beam/config"
	"beam/data/models"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func GetTaxRate(client *http.Client, contact models.OrderContact) (float64, error) {
	if client == nil {
		return 0, errors.New("http client is required")
	}

	baseURL := "https://services.maps.cdtfa.ca.gov/api/taxrate/GetRateByAddress"
	params := url.Values{}
	params.Add("address", strings.TrimSpace(contact.StreetAddress1))
	params.Add("city", strings.TrimSpace(contact.City))
	params.Add("zip", strings.TrimSpace(contact.ZipCode))
	apiURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	resp, err := client.Get(apiURL)
	if err != nil {
		return 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var taxRateResponse models.TaxRateResponse
	if err := json.NewDecoder(resp.Body).Decode(&taxRateResponse); err != nil {
		return 0, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(taxRateResponse.Errors) > 0 {
		return 0, fmt.Errorf("api error: %s", taxRateResponse.Errors[0].Message)
	}

	if len(taxRateResponse.TaxRateInfo) > 0 {
		return taxRateResponse.TaxRateInfo[0].Rate, nil
	}

	return 0, errors.New("no tax rate information found")
}

func GetRateWithFallback(client *http.Client, contact models.OrderContact, taxData *config.TaxMutex) (float64, error) {
	if client == nil || taxData == nil {
		return 0, errors.New("client, contact, and taxData are required")
	}

	taxData.Mu.RLock()
	zipBackup, hasZipBackup := taxData.CATax[contact.ZipCode]
	taxData.Mu.RUnlock()

	isCalifornia := strings.EqualFold(contact.ProvinceState, "ca") || strings.EqualFold(contact.ProvinceState, "california")
	isUS := strings.EqualFold(contact.Country, "US")

	if !(isCalifornia && isUS || hasZipBackup) {
		return 0, nil
	}

	rate, err := GetTaxRate(client, contact)
	if err != nil {
		if hasZipBackup {
			return zipBackup, nil
		}
		return 0.1025, err
	}

	return rate, nil
}

func ModifyTaxRate(draft *models.DraftOrder, tools *config.Tools, mutex *config.AllMutexes, name string, ip string) error {
	if draft.ShippingContact.StreetAddress1 == "" || draft.ShippingContact.City == "" || draft.ShippingContact.ZipCode == "" {
		return errors.New("contact is required")
	}

	isCalifornia := strings.EqualFold(draft.ShippingContact.ProvinceState, "ca") || strings.EqualFold(draft.ShippingContact.ProvinceState, "california")
	isUS := strings.EqualFold(draft.ShippingContact.Country, "US")

	if isCalifornia && isUS {
		draft.CATax = true
		rate, err := GetRateWithFallback(tools.Client, draft.ShippingContact, &mutex.Tax)
		if err != nil {
			return err
		}
		draft.CATaxRate = rate
	} else {
		err := DraftOrderEstimateUpdate(draft, draft.ShippingContact, mutex, name, ip, draft.OrderEstimate.ShipRate, tools)
		if err != nil {
			return err
		}
	}

	return nil
}
