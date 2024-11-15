package products

import (
	"beam/data"
	"fmt"

	"github.com/gin-gonic/gin"
)

func ServeProducts(fullService *data.AllServices, name string) gin.HandlerFunc {
	return func(c *gin.Context) {

		allInfo, err := fullService.Map[name].Product.GetAllProductInfo(c.Request.URL.Query(), fullService.Mutex, name)
		fmt.Print(allInfo, err)

	}
}
