package middleware

import (
	"beam/data"
	"beam/data/models"
	"beam/data/services"

	"github.com/gin-gonic/gin"
)

func FormatDataForFunctions(c *gin.Context, fullService *data.AllServices) services.DataPassIn {
	clientCookie, sessionCookie, affiliateCookie, deviceCookie := GetClientCookie(c), GetSessionCookie(c), GetAffiliateCookie(c), GetDeviceCookie(c)
	if clientCookie == nil {
		clientCookie = &models.ClientCookie{}
	}
	if sessionCookie == nil {
		sessionCookie = &models.SessionCookie{}
	}
	if affiliateCookie == nil {
		affiliateCookie = &models.AffiliateSession{}
	}
	if deviceCookie == nil {
		deviceCookie = &models.DeviceCookie{}
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
		DeviceID:      deviceCookie.DeviceID,
	}

	if serv, ok := fullService.Map[clientCookie.Store]; ok {
		ret.Logger = serv.Event
	}

	return ret
}
