package products

import (
	"beam/config"
	"beam/data/models"
	"slices"
)

func FilterByTags(tags map[string][]string, products []models.ProductInfo, mutex *config.AllMutexes, name string) ([]models.ProductInfo, map[string][]string) {

	realTags := [][]string{}
	endURL := map[string][]string{}

	mutex.Tags.Mu.RLock()

	for tag, vals := range tags {
		realTag, ok := mutex.Tags.Tags.All[name].FromURL[tag]
		if !ok {
			continue
		}
		row := []string{}
		endURL[tag] = []string{}
		for _, val := range vals {
			realVal, ok := mutex.Tags.Tags.All[name].FromURL[val]
			if !ok {
				continue
			}
			row = append(row, realTag+"__"+realVal)
			endURL[tag] = append(endURL[tag], val)
		}
		realTags = append(realTags, row)
	}

	mutex.Tags.Mu.RUnlock()

	outer := make([]bool, len(products))
	for i := range outer {
		outer[i] = true
	}
	inner := make([]bool, len(products))

	for _, tagArr := range realTags {
		for _, tag := range tagArr {
			for i, p := range products {
				if slices.Contains(p.Tags, tag) {
					inner[i] = true
				}
			}
		}

		for i, allowed := range inner {
			if !allowed {
				outer[i] = false
			}
		}

		inner = make([]bool, len(products))
	}

	var filteredProducts []models.ProductInfo

	for i, product := range products {
		if outer[i] {
			filteredProducts = append(filteredProducts, product)
		}
	}

	return filteredProducts, endURL
}
