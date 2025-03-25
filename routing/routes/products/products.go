package products

import (
	"beam/data"
	"beam/routing/middleware"
	"fmt"

	"github.com/gin-gonic/gin"
)

func ServeProducts(fullService *data.AllServices, name string) gin.HandlerFunc {
	return func(c *gin.Context) {

		dpi := middleware.FormatDataForFunctions(c, fullService)

		allInfo, err := fullService.Map[name].Product.GetAllProductInfo(dpi, c.Request.URL.Query(), fullService.Mutex, name)
		fmt.Print(allInfo, err)

	}
}
