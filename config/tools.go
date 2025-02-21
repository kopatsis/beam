package config

import (
	"beam/data/models"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	emailverifier "github.com/AfterShip/email-verifier"
	"github.com/go-redis/redis/v8"
	"github.com/oschwald/geoip2-golang"
	"github.com/sendgrid/sendgrid-go"
	"github.com/stripe/stripe-go/v81"
)

type Tools struct {
	SendGrid      *sendgrid.Client
	Client        *http.Client
	Redis         *redis.Client
	Geo           *geoip2.Reader
	EmailVerifier *emailverifier.Verifier
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
	t.initializeGeo()
	t.initializeEmailVerifier()
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

func (t *Tools) initializeGeo() {
	ipDB, err := geoip2.Open("static/geo/GeoLite2-City.mmdb")
	if err != nil {
		log.Printf("Unable to create geolite mmdb, err: %v\n", err)
		t.Geo = nil
	}
	t.Geo = ipDB
}

func (t *Tools) initializeEmailVerifier() {
	t.EmailVerifier = emailverifier.NewVerifier().EnableSMTPCheck().EnableCatchAllCheck()
}

func (t *Tools) getConversionRates() (models.ConversionResponse, error) {
	var response models.ConversionResponse

	req, err := http.NewRequest("GET", CONV_RATES, nil)
	if err != nil {
		return response, err
	}

	resp, err := t.Client.Do(req)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return response, fmt.Errorf("failed to fetch conversion rates: %s", resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return response, err
	}

	return response, nil
}

func (t *Tools) saveRates(rates models.ConversionResponse) {
	stored := models.ConversionStorage{
		Set:     time.Now(),
		Expires: time.Now().Add(5 * time.Minute),
		Rates:   rates.Rates,
	}

	data, err := json.Marshal(stored)
	if err != nil {
		log.Printf("Error marshaling conversion rates: %v\n", err)

		if err := t.Redis.Set(context.Background(), "CNV::GETTING", "0", 0).Err(); err != nil {
			log.Printf("Error saving conversion rate retrieval in process as false: %v\n", err)
		}
	}

	if err := t.Redis.MSet(context.Background(), "CNV::RATES", data, "CNV::GETTING", "0").Err(); err != nil {
		log.Printf("Error saving conversion rates and retrieval in process as false: %v\n", err)
	}
}

func (t *Tools) GetRates() (map[string]float64, error) {
	results, err := t.Redis.MGet(context.Background(), "CNV::RATES", "CNV::GETTING").Result()
	if err == nil {
		var stored models.ConversionStorage
		var getting string

		if results[0] != nil {
			if err := json.Unmarshal([]byte(results[0].(string)), &stored); err == nil {
				if stored.Expires.After(time.Now()) {
					return stored.Rates, nil
				}
				if results[1] != nil {
					getting = results[1].(string)
					if getting == "1" {
						return stored.Rates, nil
					}
				}
			}
		}
	} else {
		log.Printf("Error retrieving conversion rates: %v\n", err)
	}

	if err := t.Redis.Set(context.Background(), "CNV::GETTING", "1", 0).Err(); err != nil {
		log.Printf("Error setting conversion rate retrieval in process as true: %v\n", err)
	}

	rates, err := t.getConversionRates()
	if err != nil {
		return nil, err
	}

	go func() {
		t.saveRates(rates)
	}()

	return rates.Rates, nil
}
