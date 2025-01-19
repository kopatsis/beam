package config

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/sendgrid/sendgrid-go"
	"github.com/stripe/stripe-go/v81"
)

type Tools struct {
	SendGrid *sendgrid.Client
	Client   *http.Client
	Redis    *redis.Client
}

func NewTools(client *redis.Client) *Tools {
	t := &Tools{
		Client: &http.Client{},
	}
	if err := t.initializeSendGrid(); err != nil {
		log.Fatalf("Error initializing SendGrid: %v", err)
	}
	if err := t.initializeStripe(); err != nil {
		log.Fatalf("Error initializing Stripe: %v", err)
	}
	return t
}

func (t *Tools) initializeSendGrid() error {
	apiKey := os.Getenv("SENDGRID_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("SENDGRID_API_KEY is not set")
	}
	t.SendGrid = sendgrid.NewSendClient(apiKey)
	return nil
}

func (t *Tools) initializeStripe() error {
	stripeKey := os.Getenv("STRIPE_SECRET_KEY")
	if stripeKey == "" {
		return fmt.Errorf("STRIPE_SECRET_KEY is not set")
	}
	stripe.Key = stripeKey
	return nil
}
