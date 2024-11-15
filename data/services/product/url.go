package product

import (
	"net/url"
	"strconv"
)

func CreateBasisURL(query string, sort string, page int, tags map[string][]string) url.Values {
	url := url.Values{}

	if query != "" {
		url.Add("qy", query)
	} else if sort != "" {
		url.Add("st", sort)
	}

	url.Add("pg", strconv.Itoa(page))

	for key, vals := range tags {
		for _, val := range vals {
			url.Add(key, val)
		}
	}

	return url
}
