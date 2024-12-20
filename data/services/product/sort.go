package product

import (
	"beam/data/models"
	"sort"
)

func SortProducts(st string, products []models.ProductInfo) (string, []models.ProductInfo) {
	switch st {
	case "dd":
		sort.SliceStable(products, func(i, j int) bool {
			return products[i].DateAdded.After(products[j].DateAdded)
		})
	case "da":
		sort.SliceStable(products, func(i, j int) bool {
			return products[i].DateAdded.Before(products[j].DateAdded)
		})
	case "pd":
		sort.SliceStable(products, func(i, j int) bool {
			return products[i].Price > products[j].Price
		})
	case "pa":
		sort.SliceStable(products, func(i, j int) bool {
			return products[i].Price < products[j].Price
		})
	case "sd":
		sort.SliceStable(products, func(i, j int) bool {
			return products[i].Sales > products[j].Sales
		})
	case "sa":
		sort.SliceStable(products, func(i, j int) bool {
			return products[i].Sales < products[j].Sales
		})
	case "td":
		sort.SliceStable(products, func(i, j int) bool {
			return products[i].Title > products[j].Title
		})
	case "ta":
		sort.SliceStable(products, func(i, j int) bool {
			return products[i].Title < products[j].Title
		})
	case "rd":
		sort.SliceStable(products, func(i, j int) bool {
			return products[i].AvgRate > products[j].AvgRate
		})
	case "ra":
		sort.SliceStable(products, func(i, j int) bool {
			return products[i].AvgRate < products[j].AvgRate
		})
	default:
		st = ""
		sort.SliceStable(products, func(i, j int) bool {
			return products[i].DateAdded.After(products[j].DateAdded)
		})
	}

	return st, products
}
