package routing

import (
	"beam/config"
	"beam/data"

	"github.com/gin-gonic/gin"
)

func New(fullService *data.AllServices, tools *config.Tools) *gin.Engine {
	router := gin.Default()
	return router
}
