package product

import (
	"beam/config"
	"beam/data/models"
	"errors"
	"net/url"
	"slices"
)

func FilterByTags(tags map[string][]string, products []models.ProductInfo, mutex *config.AllMutexes, name string) ([]models.ProductInfo, map[string][]string, models.AllFilters) {

	realTags := [][]string{}
	endURL := map[string][]string{}
	forFilter := models.AllFilters{}

	mutex.Tags.Mu.RLock()

	for tag, vals := range tags {
		realTag, ok := mutex.Tags.Tags.All[name].FromURL[tag]
		if !ok {
			continue
		}
		row := []string{}
		endURL[tag] = []string{}
		forFilter.Items = append(forFilter.Items, models.FilterBlock{Key: realTag, Values: []string{}})
		for _, val := range vals {
			realVal, ok := mutex.Tags.Tags.All[name].FromURL[val]
			if !ok {
				continue
			}
			row = append(row, realTag+"__"+realVal)
			endURL[tag] = append(endURL[tag], val)
			forFilter.Items[len(forFilter.Items)-1].Values = append(forFilter.Items[len(forFilter.Items)-1].Values, realVal)
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

	return filteredProducts, endURL, forFilter
}

func CreateFilterBar(mutex *config.AllMutexes, baseURL url.Values, name string, endParams map[string][]string) (models.SideBar, error) {
	ret := models.SideBar{}

	mutex.Filters.Mu.RLock()
	mutex.Tags.Mu.RLock()

	for _, block := range mutex.Filters.Filters.All[name].Items {
		ret.Groups = append(ret.Groups, models.SideBarGroup{
			Key:  block.Key,
			Rows: []models.SideBarRow{},
		})

		for _, actual := range block.Values {

			encKey, ok := mutex.Tags.Tags.All[name].FromURL[block.Key]
			if !ok {
				return ret, errors.New("unable to locate encoded key for named Filter Tag Key: " + block.Key)
			}

			encVal, ok := mutex.Tags.Tags.All[name].FromURL[actual]
			if !ok {
				return ret, errors.New("unable to locate encoded value for named Filter Tag Value: " + actual)
			}

			exists := false
			if list, ok := endParams[encKey]; ok {
				if slices.Contains(list, encVal) {
					exists = true
				}
			}

			version := copyValues(baseURL)
			if exists {
				removeValue(version, encKey, encVal)
			} else {
				version.Add(encKey, encVal)
			}

			ret.Groups[len(ret.Groups)-1].Rows = append(ret.Groups[len(ret.Groups)-1].Rows, models.SideBarRow{
				Name:     actual,
				Link:     version,
				Selected: exists,
			})
		}
	}

	mutex.Tags.Mu.RUnlock()
	mutex.Filters.Mu.RUnlock()

	return ret, nil
}

func copyValues(original url.Values) url.Values {
	copied := url.Values{}
	for key, values := range original {
		copiedValues := make([]string, len(values))
		copy(copiedValues, values)
		copied[key] = copiedValues
	}
	return copied
}

func removeValue(params url.Values, key, valueToRemove string) {
	values := params[key]
	updatedValues := []string{}
	for _, v := range values {
		if v != valueToRemove {
			updatedValues = append(updatedValues, v)
		}
	}
	if len(updatedValues) > 0 {
		params[key] = updatedValues
	} else {
		params.Del(key)
	}
}
