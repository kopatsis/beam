package middleware

import (
	"beam/config"
	"beam/data"
	"beam/data/models"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func CookieMiddleware(fullService *data.AllServices) gin.HandlerFunc {
	return func(c *gin.Context) {

		domain := strings.Split(c.Request.Host, ":")[0]

		fullService.Mutex.Store.Mu.RLock()
		store, ok := fullService.Mutex.Store.Store.FromDomain[domain]
		fullService.Mutex.Store.Mu.RUnlock()
		if !ok {
			log.Printf("Store unable to be found from domain: %s\n", store)
			c.Redirect(http.StatusFound, config.BASIS_PAGE)
			return
		}

		clientCookie, sessionCookie, affiliateCookie := GetClientCookie(c), GetSessionCookie(c), GetAffiliateCookie(c)

		service, ok := fullService.Map[store]
		if !ok {
			log.Printf("Store unable to be found in service map: %s\n", store)
			c.Redirect(http.StatusFound, config.BASIS_PAGE)
			return
		}

		service.Customer.FullMiddleware(clientCookie, store)

		service.Session.SessionMiddleware(sessionCookie, clientCookie.CustomerID, clientCookie.GuestID, store, c)

		service.Session.AffiliateMiddleware(affiliateCookie, sessionCookie.SessionID, store, c)

		cartID, err := service.Cart.CartMiddleware(clientCookie.GetCart(), clientCookie.CustomerID, clientCookie.GuestID)
		if err != nil {
			log.Printf("Unable to correctly set cart id, error: %v; store: %s; customer ID: %d, guest ID: %s; old cart ID: %d\n", err, store, clientCookie.CustomerID, clientCookie.GuestID, clientCookie.GetCart())
		} else {
			clientCookie.SetCart(cartID)
		}

		c.Next()
	}
}

func SetClientCookie(c *gin.Context, client models.ClientCookie) {
	data, _ := json.Marshal(client)
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "client",
		Value:    string(data),
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().AddDate(10, 0, 0),
	})
}

func GetClientCookie(c *gin.Context) *models.ClientCookie {
	cookie, err := c.Cookie("client")
	if err != nil {
		return nil
	}
	var client models.ClientCookie
	if err := json.Unmarshal([]byte(cookie), &client); err != nil {
		return nil
	}
	return &client
}

func SetSessionCookie(c *gin.Context, session models.SessionCookie) {
	data, _ := json.Marshal(session)
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "session",
		Value:    string(data),
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   0,
	})
}

func GetSessionCookie(c *gin.Context) *models.SessionCookie {
	cookie, err := c.Cookie("session")
	if err != nil {
		return nil
	}
	var session models.SessionCookie
	if err := json.Unmarshal([]byte(cookie), &session); err != nil {
		return nil
	}
	return &session
}

func SetAffiliateCookie(c *gin.Context, affiliate models.AffiliateSession) {
	data, _ := json.Marshal(affiliate)
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "affiliate",
		Value:    string(data),
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   0,
	})
}

func GetAffiliateCookie(c *gin.Context) *models.AffiliateSession {
	cookie, err := c.Cookie("affiliate")
	if err != nil {
		return nil
	}
	var affiliate models.AffiliateSession
	if err := json.Unmarshal([]byte(cookie), &affiliate); err != nil {
		return nil
	}
	return &affiliate
}
