package config

import (
	"beam/data/models"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
)

type StoreNamesWithMutex struct {
	Mu    sync.RWMutex
	Store models.StoreNames
}

type TotalFiltersWithMutex struct {
	Mu      sync.RWMutex
	Filters models.TotalFilters
}

type TotalTagsWithMutex struct {
	Mu   sync.RWMutex
	Tags models.TotalTags
}

type TaxMutex struct {
	Mu    sync.RWMutex
	CATax map[string]float64
}

type APIKeyMutex struct {
	Mu     sync.RWMutex
	KeyMap map[string]string
}

type IsoCodesMutex struct {
	Mu        sync.RWMutex
	Countries models.CountryCodes
	States    models.StateCodes
}

type CurrencyMutex struct {
	Mu   sync.RWMutex
	List models.CurrencyCodes
}

type SettingsMutex struct {
	Mu       sync.RWMutex
	Settings models.SpecialStoreSettings
}

type AllMutexes struct {
	Store    StoreNamesWithMutex
	Filters  TotalFiltersWithMutex
	Tags     TotalTagsWithMutex
	Tax      TaxMutex
	Api      APIKeyMutex
	Iso      IsoCodesMutex
	Currency CurrencyMutex
	Settings SettingsMutex
}

func unmarshalJSONFile(filePath string, v interface{}) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("unable to read file %s: %v", filePath, err)
	}
	err = json.Unmarshal(data, v)
	if err != nil {
		return fmt.Errorf("unable to unmarshal file %s: %v", filePath, err)
	}
	return nil
}

func LoadAllData() *AllMutexes {
	storeNamesFile := "static/ref/allstorenames.json"
	filtersFile := "static/ref/allfilters.json"
	tagsFile := "static/ref/alltags.json"
	taxFile := "static/ref/tax.json"
	countryFile := "static/ref/countryiso.json"
	stateFile := "static/ref/stateiso.json"
	currFile := "static/ref/currency.json"
	settingsFile := "static/ref/storesettings.json"

	var storeNames models.StoreNames
	var totalFilters models.TotalFilters
	var totalTags models.TotalTags
	var countries models.CountryCodes
	var states models.StateCodes
	var currency models.CurrencyCodes
	var tax map[string]float64
	var settings models.SpecialStoreSettings

	if err := unmarshalJSONFile(storeNamesFile, &storeNames); err != nil {
		log.Fatalf("Unable to load the store mutex vars: %v", err)
	}
	if err := unmarshalJSONFile(filtersFile, &totalFilters); err != nil {
		log.Fatalf("Unable to load the filter mutex vars: %v", err)
	}
	if err := unmarshalJSONFile(tagsFile, &totalTags); err != nil {
		log.Fatalf("Unable to load the tags mutex vars: %v", err)
	}
	if err := unmarshalJSONFile(taxFile, &tax); err != nil {
		log.Fatalf("Unable to load the tags mutex vars: %v", err)
	}
	if err := unmarshalJSONFile(countryFile, &countries); err != nil {
		log.Fatalf("Unable to load the country mutex vars: %v", err)
	}
	if err := unmarshalJSONFile(stateFile, &states); err != nil {
		log.Fatalf("Unable to load the state mutex vars: %v", err)
	}
	if err := unmarshalJSONFile(currFile, &currency); err != nil {
		log.Fatalf("Unable to load the currency mutex vars: %v", err)
	}
	if err := unmarshalJSONFile(settingsFile, &settings); err != nil {
		log.Fatalf("Unable to load the store special settings mutex vars: %v", err)
	}

	keyMap := map[string]string{}

	for key := range storeNames.ToDomain {
		envKey := key + "_PF_API_KEY"
		apiKey := os.Getenv(envKey)
		if apiKey != "" {
			keyMap[key] = apiKey
		} else {
			log.Fatalf("No key for: %s\n", key)
		}
	}

	return &AllMutexes{
		Store:    StoreNamesWithMutex{Store: storeNames},
		Filters:  TotalFiltersWithMutex{Filters: totalFilters},
		Tags:     TotalTagsWithMutex{Tags: totalTags},
		Tax:      TaxMutex{CATax: tax},
		Api:      APIKeyMutex{KeyMap: keyMap},
		Iso:      IsoCodesMutex{Countries: countries, States: states},
		Currency: CurrencyMutex{List: currency},
		Settings: SettingsMutex{Settings: settings},
	}
}
