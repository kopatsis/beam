package sessionhelp

import (
	"beam/config"
	"beam/data/models"
	"crypto/sha256"
	"encoding/hex"

	"github.com/gin-gonic/gin"
	"github.com/mssola/user_agent"
	"github.com/ua-parser/uap-go/uaparser"
)

func CreateSessionDetails(c *gin.Context, tools *config.Tools, session *models.Session) {

	if session == nil {
		return
	}

	var city string
	var country string

	ipStr := c.ClientIP()

	if ipStr == "" || ipStr == "::1" {
		ipStr = c.Request.Header.Get("X-Forwarded-For")
	}

	city, country = config.GetLocation(ipStr, tools)

	parser := uaparser.NewFromSaved()

	ua := c.Request.UserAgent()
	client := parser.Parse(ua)

	browser := client.UserAgent.Family
	os := client.Os.Family
	platform := client.Device.Family

	uaM := user_agent.New(ua)
	isMobile := uaM.Mobile()
	isBot := uaM.Bot()

	savedIP := hex.EncodeToString(sha256.New().Sum([]byte(ipStr)))

	session.Referrer = c.Request.Referer()
	session.IPAddress = savedIP
	session.InitialRoute = c.Request.URL.Path
	session.FullURL = c.Request.URL.String()
	session.City = city
	session.Country = country
	session.Browser = browser
	session.OS = os
	session.Platform = platform
	session.Mobile = isMobile
	session.Bot = isBot
}
