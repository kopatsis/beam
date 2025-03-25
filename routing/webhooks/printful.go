package webhooks

import (
	"beam/background/apidata"
	"beam/background/emails"
	"beam/config"
	"beam/data"
	"beam/routing/middleware"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func verifySignature(body []byte, signature, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expectedMAC := mac.Sum(nil)

	decodedSignature, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false
	}

	return hmac.Equal(expectedMAC, decodedSignature)
}

func HandlePrintfulWebhooks(c *gin.Context, fullService *data.AllServices, tools *config.Tools) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	signature := c.GetHeader("X-PF-WEBHOOK-SIGNATURE")
	secret := os.Getenv("PRINTFUL_WEBHOOK_SECRET")

	if !verifySignature(body, signature, secret) {
		c.Status(http.StatusUnauthorized)
		return
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid format for JSON"})
		return
	}

	store := c.Param("store")
	dpi := middleware.FormatDataWebhooks(c, fullService, store)

	if eventType, ok := payload["type"].(string); !ok || eventType != "package_shipped" {
		emails.HandleWebhook(tools, payload)
		c.Status(http.StatusOK)
		return
	}

	service, ok := fullService.Map[store]
	if !ok {
		c.Status(http.StatusBadRequest)
		return
	}

	var shippedData apidata.PackageShippedPF
	if err := c.ShouldBindJSON(&shippedData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON for package shipped"})
		return
	}

	if err := service.Order.ShipOrder(dpi, store, shippedData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}
