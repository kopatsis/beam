package models

import "time"

type TaxRateResponse struct {
	TaxRateInfo []struct {
		Rate float64 `json:"rate"`
	} `json:"taxRateInfo"`
	Errors []struct {
		Field   string `json:"field"`
		Message string `json:"message"`
	} `json:"errors"`
}

type ConversionStorage struct {
	Set     time.Time          `json:"s"`
	Expires time.Time          `json:"e"`
	Rates   map[string]float64 `json:"r"`
}

type ConversionResponse struct {
	Success   bool               `json:"success"`
	Terms     string             `json:"terms"`
	Privacy   string             `json:"privacy"`
	Timestamp int                `json:"timestamp"`
	Date      time.Time          `json:"date"`
	Base      string             `json:"base"`
	Rates     map[string]float64 `json:"rates"`
}
