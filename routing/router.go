package routing

import (
	"beam/data"

	"github.com/gin-gonic/gin"
)

func New(fullService *data.AllServices) *gin.Engine {
	router := gin.Default()
	return router
}
