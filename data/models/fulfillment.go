package models

import "time"

type SubLineFulfillment struct {
	LineItemID          int
	SubLineQuantity     int
	OrderFulfillmentIDs []string
	Status              string // Unfulfilled, Partial, Fulfilled
}

type OrderFulfillment struct {
	ID             string
	Status         string    // Active, Inactive
	PrintfulID     int       `json:"id"`
	Carrier        string    `json:"carrier"`
	Service        string    `json:"service"`
	TrackingNumber int       `json:"tracking_number"`
	TrackingURL    string    `json:"tracking_url"`
	Created        time.Time `json:"created"`
	ShipDate       time.Time `json:"ship_date"`
	ShippedAt      time.Time `json:"shipped_at"`
	Reshipment     bool      `json:"reshipment"`
}
