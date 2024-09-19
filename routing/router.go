package routing

import (
	"beam/data"

	"github.com/gin-gonic/gin"
)

func New(fullService *data.MainService) *gin.Engine {
	router := gin.Default()
	return router
}
