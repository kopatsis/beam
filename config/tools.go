package config

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/sendgrid/sendgrid-go"
)

type Tools struct {
	SendGrid *sendgrid.Client
	Client   *http.Client
}

func NewTools() *Tools {
	t := &Tools{
		Client: &http.Client{},
	}
	if err := t.initializeSendGrid(); err != nil {
		log.Fatalf("Error initializing SendGrid: %v", err)
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
