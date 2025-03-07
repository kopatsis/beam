package middleware

import (
	"beam/data"
	"beam/data/models"
	"beam/data/services"

	"github.com/gin-gonic/gin"
)

func FormatDataForFunctions(c *gin.Context, fullService *data.AllServices) services.DataPassIn {
	clientCookie, sessionCookie, affiliateCookie := GetClientCookie(c), GetSessionCookie(c), GetAffiliateCookie(c)
	if clientCookie == nil {
		clientCookie = &models.ClientCookie{}
	}
	if sessionCookie == nil {
		sessionCookie = &models.SessionCookie{}
	}
	if affiliateCookie == nil {
		affiliateCookie = &models.AffiliateSession{}
	}

	ipStr := c.ClientIP()

	if ipStr == "" || ipStr == "::1" {
		ipStr = c.Request.Header.Get("X-Forwarded-For")
	}

	ret := services.DataPassIn{
		Store:         clientCookie.Store,
		CustomerID:    clientCookie.CustomerID,
		GuestID:       clientCookie.GuestID,
		IsLoggedIn:    clientCookie.CustomerID > 0,
		CartID:        clientCookie.GetCart(),
		SessionID:     sessionCookie.SessionID,
		AffiliateID:   affiliateCookie.ID,
		AffiliateCode: affiliateCookie.ActualCode,
		IPAddress:     ipStr,
	}

	if serv, ok := fullService.Map[clientCookie.Store]; ok {
		ret.Logger = serv.Event
	}

	return ret
}
