package products

import (
	"beam/data"
	"beam/data/models"
	"fmt"

	"github.com/gin-gonic/gin"
)

func ServeProducts(fullService *data.AllServices, name string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var query, page, sort string
		otherParams := map[string][]string{}

		for key, values := range c.Request.URL.Query() {
			switch key {
			case "qy":
				if len(values) > 0 {
					query = values[0]
				}
			case "pg":
				if len(values) > 0 {
					page = values[0]
				}
			case "st":
				if len(values) > 0 {
					sort = values[0]
				}
			default:
				otherParams[key] = values
			}
		}
		if len(query) > 128 {
			query = query[0:127]
		}

		products, err := fullService.Map[name].Product.GetAllProductInfo(name)
		if err != nil {
			fmt.Print(err.Error())
		}

		endURL := map[string][]string{}
		forFilter := models.AllFilters{}
		if len(otherParams) > 0 {
			products, endURL, forFilter = FilterByTags(otherParams, products, fullService.Mutex, name)
		}

		realsort := ""
		if query != "" {
			products, err = FuzzySearch(query, products)
			if err != nil {
				fmt.Print(err.Error())
			}
		} else {
			realsort, products = SortProducts(sort, products)
		}

		var left, pg, right int
		products, left, pg, right = PageProducts(page, products)
		fmt.Print(left, right, products, forFilter)

		baseURL := CreateBasisURL(query, realsort, pg, endURL)
		fmt.Print(baseURL)

	}
}
