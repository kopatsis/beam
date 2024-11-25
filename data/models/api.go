package models

type TaxRateResponse struct {
	TaxRateInfo []struct {
		Rate float64 `json:"rate"`
	} `json:"taxRateInfo"`
	Errors []struct {
		Field   string `json:"field"`
		Message string `json:"message"`
	} `json:"errors"`
}
