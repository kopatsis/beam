package routing

import (
	"beam/config"
	"beam/data"

	"github.com/gin-gonic/gin"
)

func New(fullService *data.MainService, mutexes *config.AllMutexes) *gin.Engine {
	router := gin.Default()
	return router
}
