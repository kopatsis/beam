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
	"github.com/google/uuid"
)

func CookieMiddleware(fullService *data.AllServices, tools *config.Tools) gin.HandlerFunc {
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

		clientCookie, sessionCookie, affiliateCookie, deviceCookie, twofaCookie, signInCookie := GetClientCookie(c), GetSessionCookie(c), GetAffiliateCookie(c), GetDeviceCookie(c), GetTwoFACookie(c), GetSignInCodeCookie(c)

		service, ok := fullService.Map[store]
		if !ok {
			log.Printf("Store unable to be found in service map: %s\n", store)
			c.Redirect(http.StatusFound, config.BASIS_PAGE)
			return
		}

		service.Session.DeviceMiddleware(deviceCookie)
		if deviceCookie == nil {
			deviceCookie = &models.DeviceCookie{DeviceID: "DV:" + uuid.NewString()}
		}

		service.Customer.FullMiddleware(clientCookie, deviceCookie, store)
		if clientCookie == nil {
			clientCookie = &models.ClientCookie{}
		}

		service.Session.SessionMiddleware(sessionCookie, clientCookie.CustomerID, clientCookie.GuestID, store, c, tools)
		if sessionCookie == nil {
			sessionCookie = &models.SessionCookie{}
		}

		service.Session.AffiliateMiddleware(affiliateCookie, sessionCookie.SessionID, store, c)
		if affiliateCookie == nil {
			affiliateCookie = &models.AffiliateSession{}
		}

		cartID, err := service.Cart.CartMiddleware(clientCookie.GetCart(), clientCookie.CustomerID, clientCookie.GuestID)
		if err != nil {
			log.Printf("Unable to correctly set cart id, error: %v; store: %s; customer ID: %d, guest ID: %s; old cart ID: %d\n", err, store, clientCookie.CustomerID, clientCookie.GuestID, clientCookie.GetCart())
		} else {
			clientCookie.SetCart(cartID)
		}

		service.Customer.TwoFAMiddleware(clientCookie, twofaCookie)
		if twofaCookie == nil {
			twofaCookie = &models.TwoFactorCookie{}
		}

		service.Customer.SignInCodeMiddleware(clientCookie, signInCookie)
		if signInCookie == nil {
			signInCookie = &models.SignInCodeCookie{}
		}

		SetClientCookie(c, *clientCookie)
		SetSessionCookie(c, *sessionCookie)
		SetAffiliateCookie(c, *affiliateCookie)
		SetDeviceCookie(c, *deviceCookie)
		SetTwoFACookie(c, *twofaCookie)

		c.Next()

		SetCheckCookie(c)
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
		Expires:  time.Now().AddDate(100, 0, 0),
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

func SetDeviceCookie(c *gin.Context, device models.DeviceCookie) {
	data, _ := json.Marshal(device)
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "device",
		Value:    string(data),
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   0,
	})
}

func GetDeviceCookie(c *gin.Context) *models.DeviceCookie {
	cookie, err := c.Cookie("device")
	if err != nil {
		return nil
	}
	var device models.DeviceCookie
	if err := json.Unmarshal([]byte(cookie), &device); err != nil {
		return nil
	}
	return &device
}

func SetTwoFACookie(c *gin.Context, twofa models.TwoFactorCookie) {
	if twofa.TwoFactorCode == "" {
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     "twofa",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			MaxAge:   -1,
		})
		return
	}

	data, _ := json.Marshal(twofa)
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "twofa",
		Value:    string(data),
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   0,
	})
}

func GetTwoFACookie(c *gin.Context) *models.TwoFactorCookie {
	cookie, err := c.Cookie("twofa")
	if err != nil {
		return nil
	}
	var twofa models.TwoFactorCookie
	if err := json.Unmarshal([]byte(cookie), &twofa); err != nil {
		return nil
	}
	return &twofa
}

func SetResetCookie(c *gin.Context, reset models.ResetEmailCookie) {
	if reset.Param == "" {
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     "reset",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			MaxAge:   -1,
		})
		return
	}

	data, _ := json.Marshal(reset)
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "reset",
		Value:    string(data),
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   0,
	})
}

func GetResetCookie(c *gin.Context) *models.ResetEmailCookie {
	cookie, err := c.Cookie("reset")
	if err != nil {
		return nil
	}
	var reset models.ResetEmailCookie
	if err := json.Unmarshal([]byte(cookie), &reset); err != nil {
		return nil
	}
	return &reset
}

func SetSignInCodeCookie(c *gin.Context, si models.SignInCodeCookie) {
	if si.Param == "" {
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     "signin6",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			MaxAge:   -1,
		})
		return
	}

	data, _ := json.Marshal(si)
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "signin6",
		Value:    string(data),
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   0,
	})
}

func GetSignInCodeCookie(c *gin.Context) *models.SignInCodeCookie {
	cookie, err := c.Cookie("signin6")
	if err != nil {
		return nil
	}
	var si models.SignInCodeCookie
	if err := json.Unmarshal([]byte(cookie), &si); err != nil {
		return nil
	}
	return &si
}

func SetCheckCookie(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "setcheck",
		Value:    "checkset",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   0,
	})
}

func CheckCookiesAllowed(c *gin.Context) bool {
	cookie, err := c.Cookie("setcheck")
	return err != nil && cookie == "checkset"
}
