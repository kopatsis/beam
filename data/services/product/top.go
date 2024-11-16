package product

import (
	"beam/data/models"
	"strings"
)

func CreateTopWords(forTop models.AllFilters, query string) models.TopWords {
	ret := models.TopWords{}

	if query != "" {
		ret.Query = `Matching Search: "` + query + `"`
	}

	forTop.Sort()

	for _, block := range forTop.Items {
		if block.Key == "Collection" {
			ret.Collection = "Collection: " + JoinEnglish(block.Values)
		} else {
			ret.Lines = append(ret.Lines, block.Key+": "+JoinEnglish(block.Values))
		}
	}

	return ret
}

func JoinEnglish(items []string) string {
	n := len(items)
	if n == 0 {
		return ""
	}
	if n == 1 {
		return items[0]
	}
	if n == 2 {
		return items[0] + " or " + items[1]
	}
	return strings.Join(items[:n-1], ", ") + ", or " + items[n-1]
}
