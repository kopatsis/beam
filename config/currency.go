package config

import (
	"beam/data/models"
	"fmt"
	"log"
)

func CollectionCurrency(c *models.ClientCookie, t *Tools, render *models.CollectionRender) {
	rate, currency, otherCurrency := getCurrency(c, t)

	for _, p := range render.Products {
		render.PricedProducts = append(render.PricedProducts, models.ProductWithPriceRender{
			Product: p,
			PriceRender: models.PriceRender{
				DollarPrice:    fmt.Sprintf("$%.2f", float64(p.Price)/100),
				IsOtherPrice:   otherCurrency,
				OtherPrice:     fmt.Sprintf("%.2f", (float64(p.Price)*rate)/100),
				OtherPriceCode: currency,
			},
		})
	}
	render.Products = []models.ProductInfo{}
}

func ProductCurrency(c *models.ClientCookie, t *Tools, render *models.ProductRender) {
	rate, currency, otherCurrency := getCurrency(c, t)

	render.PriceRender = models.PriceRender{
		DollarPrice:    fmt.Sprintf("$%.2f", float64(render.Price)/100),
		IsOtherPrice:   otherCurrency,
		OtherPrice:     fmt.Sprintf("%.2f", (float64(render.Price)*rate)/100),
		OtherPriceCode: currency,
	}

	render.CompareAtRender = models.PriceRender{
		DollarPrice:    fmt.Sprintf("$%.2f", float64(render.CompareAt)/100),
		IsOtherPrice:   otherCurrency,
		OtherPrice:     fmt.Sprintf("%.2f", (float64(render.CompareAt)*rate)/100),
		OtherPriceCode: currency,
	}
}

func CartCurrency(c *models.ClientCookie, t *Tools, render *models.CartRender) {

	rate, currency, otherCurrency := getCurrency(c, t)

	for i, cl := range render.CartLines {
		cl.PriceRender = models.PriceRender{
			DollarPrice:    fmt.Sprintf("$%.2f", float64(cl.Subtotal)/100),
			IsOtherPrice:   otherCurrency,
			OtherPrice:     fmt.Sprintf("%.2f", (float64(cl.Subtotal)*rate)/100),
			OtherPriceCode: currency,
		}
		render.CartLines[i] = cl
	}

	render.PriceRender = models.PriceRender{
		DollarPrice:    fmt.Sprintf("$%.2f", float64(render.Subtotal)/100),
		IsOtherPrice:   otherCurrency,
		OtherPrice:     fmt.Sprintf("%.2f", (float64(render.Subtotal)*rate)/100),
		OtherPriceCode: currency,
	}

}

func DraftOrderCurrency(c *models.ClientCookie, t *Tools, draft *models.DraftOrder) models.DraftOrderRender {

	rate, currency, otherCurrency := getCurrency(c, t)

	return models.DraftOrderRender{
		DraftOrder: draft,
		TotalPriceRender: models.PriceRender{
			DollarPrice:    fmt.Sprintf("$%.2f", float64(draft.Total)/100),
			IsOtherPrice:   otherCurrency,
			OtherPrice:     fmt.Sprintf("%.2f", (float64(draft.Total)*rate)/100),
			OtherPriceCode: currency,
		},
	}

}

// Returns the flot64 USD -> rate, name of the other rate, and whether or not to continue
// Will return false for continue if currency is USD and/or failed to get rates
func getCurrency(c *models.ClientCookie, t *Tools) (float64, string, bool) {
	if c == nil || t == nil || !c.OtherCurrency {
		return 1, "USD", false
	}

	rates, err := t.GetRates()
	if err != nil {
		log.Printf("Unable to retrieve rates for currency conversion, err: %v; code: %s\n", err, c.Currency)
		return 1, "USD", false
	}

	if rate, ok := rates[c.Currency]; ok {
		return rate, c.Currency, true
	}

	return 1, "USD", false
}
