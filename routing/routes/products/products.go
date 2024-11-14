package products

import (
	"beam/data"
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

		if query != "" {
			products, err = FuzzySearch(query, products)
		}

		if len(otherParams) > 0 {
			products, err = FilterByTags(otherParams, products)
		}

		fmt.Println(query, page, sort)
		fmt.Print(products, err)

	}
}
