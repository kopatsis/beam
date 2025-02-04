package sessionhelp

import (
	"beam/config"
	"beam/data/models"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mssola/user_agent"
	"github.com/ua-parser/uap-go/uaparser"
)

func SessionContextData(c *gin.Context, cookie *models.SessionCookie) {
	fmt.Println("not done")
}

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

	if ipStr != "" && tools.Geo != nil {
		if commaIndex := strings.Index(ipStr, ","); commaIndex != -1 {
			ipStr = ipStr[:commaIndex]
		}

		ip := net.ParseIP(ipStr)
		if ip != nil {
			record, err := tools.Geo.City(ip)
			if err == nil && record != nil {
				city = record.City.Names["en"]
				country = record.Country.Names["en"]
			}
		}
	}

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
