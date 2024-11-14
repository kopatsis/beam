package products

import "beam/data/models"

func FilterByTags(tags map[string][]string, products []models.ProductInfo) ([]models.ProductInfo, error) {
	return products, nil
}
