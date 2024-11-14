package products

import (
	"beam/data/models"

	"github.com/schollz/closestmatch"

	"sort"
)

func FuzzySearch(query string, products []models.ProductInfo) ([]models.ProductInfo, error) {
	var cm *closestmatch.ClosestMatch
	allFields := []string{}

	for _, product := range products {
		allFields = append(allFields, product.Handle, product.Title)
		allFields = append(allFields, product.Tags...)
		allFields = append(allFields, product.SKUs...)
		allFields = append(allFields, product.Var1Values...)
		allFields = append(allFields, product.Var2Values...)
		allFields = append(allFields, product.Var3Values...)
	}

	cm = closestmatch.New(allFields, nil)

	productScores := []struct {
		Product       models.ProductInfo
		Score         int
		MatchedFields []string
	}{}

	for _, product := range products {
		score := 0
		matchedFields := []string{}
		for _, field := range []string{product.Handle, product.Title} {
			if cm.Closest(field) == query {
				score += 10
				matchedFields = append(matchedFields, field)
			}
		}
		for _, tag := range product.Tags {
			if cm.Closest(tag) == query {
				score += 5
				matchedFields = append(matchedFields, tag)
			}
		}
		for _, sku := range product.SKUs {
			if cm.Closest(sku) == query {
				score += 5
				matchedFields = append(matchedFields, sku)
			}
		}
		for _, value := range product.Var1Values {
			if cm.Closest(value) == query {
				score += 3
				matchedFields = append(matchedFields, value)
			}
		}
		for _, value := range product.Var2Values {
			if cm.Closest(value) == query {
				score += 3
				matchedFields = append(matchedFields, value)
			}
		}
		for _, value := range product.Var3Values {
			if cm.Closest(value) == query {
				score += 3
				matchedFields = append(matchedFields, value)
			}
		}

		productScores = append(productScores, struct {
			Product       models.ProductInfo
			Score         int
			MatchedFields []string
		}{Product: product, Score: score, MatchedFields: matchedFields})
	}

	sort.Slice(productScores, func(i, j int) bool {
		return productScores[i].Score > productScores[j].Score
	})

	var sortedProducts []models.ProductInfo
	for _, ps := range productScores {
		sortedProducts = append(sortedProducts, ps.Product)
	}

	return sortedProducts, nil
}
