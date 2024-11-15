package product

import "beam/data/models"

func CreateTopWords(forTop models.AllFilters, query string) models.TopWords {
	ret := models.TopWords{}

	if query != "" {
		ret.Query = `Matching Search: "` + query + `"`
	}

	return ret
}
