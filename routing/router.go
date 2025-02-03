package routing

import (
	"beam/config"
	"beam/data"
	"beam/routing/middleware"

	"github.com/gin-gonic/gin"
)

func New(fullService *data.AllServices, tools *config.Tools) *gin.Engine {
	router := gin.Default()
	router.Use(middleware.CookieMiddleware(fullService))
	return router
}
