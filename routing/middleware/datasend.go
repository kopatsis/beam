package middleware

import (
	"beam/background/logging"
	"beam/config"
	"beam/data"
	"beam/data/models"
	"beam/data/services"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func FormatDataForFunctions(c *gin.Context, fullService *data.AllServices) *services.DataPassIn {
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
		SessionLineID: "SL-" + uuid.NewString(),
		AffiliateID:   affiliateCookie.ID,
		AffiliateCode: affiliateCookie.ActualCode,
		IPAddress:     ipStr,
		TimeStarted:   time.Now(),
		Logs:          []models.EventFinal{},
		LogsMutex:     sync.Mutex{},
	}

	if serv, ok := fullService.Map[clientCookie.Store]; ok {
		ret.Logger = serv.Event
	}

	return &ret
}

func FormatDataWebhooks(c *gin.Context, fullService *data.AllServices, store string) *services.DataPassIn {
	ipStr := c.ClientIP()

	if ipStr == "" || ipStr == "::1" {
		ipStr = c.Request.Header.Get("X-Forwarded-For")
	}

	ret := services.DataPassIn{
		SessionLineID: "SL-" + uuid.NewString(),
		IPAddress:     ipStr,
		TimeStarted:   time.Now(),
		Logs:          []models.EventFinal{},
		LogsMutex:     sync.Mutex{},
	}

	if serv, ok := fullService.Map[store]; ok {
		ret.Logger = serv.Event
	}

	return &ret
}

func PostLogs(dpi *services.DataPassIn, tools *config.Tools) {
	payload, err := dpi.MarshalLogs()
	if err != nil {
		logging.AsyncCriticalError(tools, "", "Unable to marshal to payload logs from dpi", err)
		return
	}

	logging.LogsToLoggly(tools, payload)
}
