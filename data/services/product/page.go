package product

import (
	"beam/config"
	"beam/data/models"
	"net/url"
	"strconv"
)

func PageProducts(page string, products []models.ProductInfo) ([]models.ProductInfo, int, int, int) {
	pNum := 1
	val, err := strconv.Atoi(page)
	if err == nil {
		if ((val-1)*config.PAGELEN) < len(products) && val > 0 {
			pNum = val
		}
	}
	left := len(products)/config.PAGELEN - pNum + 1

	return products[(pNum-1)*config.PAGELEN : min(len(products), pNum*config.PAGELEN)], pNum - 1, pNum, left
}

func PageRender(page, pageLeft, pageRight int, baseURL url.Values) models.Paging {
	leftURL, rightURL := url.Values{}, url.Values{}
	if pageLeft != 0 {
		leftURL = copyValues(baseURL)
		leftURL.Set("pg", strconv.Itoa(page-1))
	}
	if pageRight != 0 {
		rightURL = copyValues(baseURL)
		rightURL.Set("pg", strconv.Itoa(page+1))
	}
	return models.Paging{
		Page:      page,
		PageLeft:  pageLeft,
		PageRight: pageRight,
		LeftURL:   leftURL,
		RightURL:  rightURL,
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
