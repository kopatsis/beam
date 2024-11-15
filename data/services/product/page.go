package product

import (
	"beam/config"
	"beam/data/models"
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
