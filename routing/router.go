package routing

import (
	"beam/data"
	"net/http"

	"github.com/gin-gonic/gin"
)

func New(fullService *data.AllServices, client *http.Client) *gin.Engine {
	router := gin.Default()
	return router
}
