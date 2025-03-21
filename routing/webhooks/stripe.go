package webhooks

import (
	"beam/config"
	"beam/data"
	"beam/data/services/orderhelp"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/webhook"
)

func PaymentSuccess(c *gin.Context, fullService *data.AllServices, tools *config.Tools) {
	const maxBodyBytes = int64(65536)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodyBytes)
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	sigHeader := c.GetHeader("Stripe-Signature")
	event, err := webhook.ConstructEvent(payload, sigHeader, os.Getenv("STRIPE_SECRET"))
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if event.Type != "payment_intent.succeeded" {
		c.Status(http.StatusOK)
		return
	}

	var intent stripe.PaymentIntent
	err = json.Unmarshal(event.Data.Raw, &intent)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	paymentIntentID := intent.ID

	orderInfo, err := orderhelp.IntentToOrderGet(tools.Redis, paymentIntentID)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	service, ok := fullService.Map[orderInfo.Store]
	if !ok {
		log.Printf("Store unable to be found in service map: %s\n", orderInfo.Store)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	go func() {
		service.Order.CompleteOrder(orderInfo.Store, orderInfo.OrderID, service.Customer, service.DraftOrder, service.Discount, service.List, service.Product, service.Order, service.Session, fullService.Mutex, tools)
	}()

	c.Status(http.StatusOK)

}

func PaymentFailure(c *gin.Context, fullService *data.AllServices, tools *config.Tools) {
	const maxBodyBytes = int64(65536)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodyBytes)
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	sigHeader := c.GetHeader("Stripe-Signature")
	event, err := webhook.ConstructEvent(payload, sigHeader, os.Getenv("STRIPE_SECRET"))
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if event.Type != "payment_intent.payment_failed" {
		c.Status(http.StatusOK)
		return
	}

	var intent stripe.PaymentIntent
	err = json.Unmarshal(event.Data.Raw, &intent)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	paymentIntentID := intent.ID

	orderInfo, err := orderhelp.IntentToOrderGet(tools.Redis, paymentIntentID)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	service, ok := fullService.Map[orderInfo.Store]
	if !ok {
		log.Printf("Store unable to be found in service map: %s\n", orderInfo.Store)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	go func() {
		service.Order.OrderPaymentFailure(orderInfo.Store, orderInfo.OrderID, fullService.Mutex, tools)
	}()

	c.Status(http.StatusOK)

}
