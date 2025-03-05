package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func EvaluateWelcomePopup(c *gin.Context) bool {
	cookie, err := c.Cookie("welcomepop")
	return err != nil && cookie == "sent"
}

func WelcomePopupSent(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "welcomepop",
		Value:    "sent",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   0,
	})
}
