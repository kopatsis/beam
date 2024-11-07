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

type AllMutexes struct {
	Store   StoreNamesWithMutex
	Filters TotalFiltersWithMutex
	Tags    TotalTagsWithMutex
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

	var storeNames models.StoreNames
	var totalFilters models.TotalFilters
	var totalTags models.TotalTags

	if err := unmarshalJSONFile(storeNamesFile, &storeNames); err != nil {
		log.Fatalf("Unable to load the store mutex vars: %v", err)
	}
	if err := unmarshalJSONFile(filtersFile, &totalFilters); err != nil {
		log.Fatalf("Unable to load the filter mutex vars: %v", err)
	}
	if err := unmarshalJSONFile(tagsFile, &totalTags); err != nil {
		log.Fatalf("Unable to load the tags mutex vars: %v", err)
	}

	return &AllMutexes{
		Store:   StoreNamesWithMutex{Store: storeNames},
		Filters: TotalFiltersWithMutex{Filters: totalFilters},
		Tags:    TotalTagsWithMutex{Tags: totalTags},
	}
}