package customer

import (
	"beam/config"
	"beam/data/models"
	"errors"
	"fmt"
)

func VerifyContact(contact *models.Contact, mutex *config.AllMutexes) error {

	if contact.FirstName == "" {
		return errors.New("first name cannot be blank")
	} else if contact.StreetAddress1 == "" {
		return errors.New("street address 1 cannot be blank")
	} else if contact.City == "" {
		return errors.New("city cannot be blank")
	} else if contact.ZipCode == "" {
		return errors.New("zip code cannot be blank")
	} else if contact.Country == "" {
		return errors.New("country cannot be blank")
	}

	countryCode := ""

	mutex.Iso.Mu.RLock()
	found := false
	for _, bl := range mutex.Iso.Countries.List {
		if bl.Name == contact.Country {
			countryCode = bl.Code
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("unknown country: %s encountered (should not be possible) on customer id: %d in convert to printful", contact.Country, contact.CustomerID)
	}
	contact.CountryCode = countryCode

	stateCode := ""
	found = false
	if countryCode == "US" {
		for _, bl := range mutex.Iso.States.US {
			if bl.Name == contact.ProvinceState {
				stateCode = bl.Code
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("unknown state: %s encountered for country code: %s (should not be possible) on customer id: %d in convert to printful", contact.ProvinceState, countryCode, contact.CustomerID)
		}
		contact.StateCode = stateCode

	} else if countryCode == "MX" {
		for _, bl := range mutex.Iso.States.MX {
			if bl.Name == contact.ProvinceState {
				stateCode = bl.Code
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("unknown state: %s encountered for country code: %s (should not be possible) on customer id: %d in convert to printful", contact.ProvinceState, countryCode, contact.CustomerID)
		}
		contact.StateCode = stateCode

	} else if countryCode == "AU" {
		for _, bl := range mutex.Iso.States.AU {
			if bl.Name == contact.ProvinceState {
				stateCode = bl.Code
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("unknown state: %s encountered for country code: %s (should not be possible) on customer id: %d in convert to printful", contact.ProvinceState, countryCode, contact.CustomerID)
		}
		contact.StateCode = stateCode

	} else if countryCode == "CA" {
		for _, bl := range mutex.Iso.States.CA {
			if bl.Name == contact.ProvinceState {
				stateCode = bl.Code
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("unknown state: %s encountered for country code: %s (should not be possible) on customer id: %d in convert to printful", contact.ProvinceState, countryCode, contact.CustomerID)
		}
		contact.StateCode = stateCode

	}

	mutex.Iso.Mu.RUnlock()

	return nil
}
