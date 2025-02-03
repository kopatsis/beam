package sessionhelp

import (
	"beam/data/models"
	"fmt"

	"github.com/gin-gonic/gin"
)

func SessionContextData(c *gin.Context, cookie *models.SessionCookie) {
	fmt.Println("not done")
}
