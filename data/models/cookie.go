package models

import "time"

type DataPassIn struct {
	Store         string
	CustomerID    int
	IsLoggedIn    bool
	GuestID       string
	CartID        string
	SessionID     string
	AffiliateID   int
	AffiliateCode string
	FirebaseID    string
}

type ServerCookie struct {
	Store      string    `json:"s"`
	FirebaseID string    `json:"f"`
	CustomerID int       `json:"c"`
	LastReset  time.Time `json:"l"`
	Archived   bool      `json:"a"`
}

// Client
type ClientCookie struct {
	Store        string    `json:"s"`
	CustomerID   int       `json:"c"`
	CustomerSet  time.Time `json:"t"`
	GuestID      string    `json:"g"`
	CustomerCart int       `json:"a"`
	GuestCart    int       `json:"r"`
}

// Session
type SessionCookie struct {
	Store      string    `json:"s"`
	CustomerID int       `json:"c"`
	GuestID    string    `json:"g"`
	Assigned   time.Time `json:"t"`
	SessionID  string    `json:"i"`
}

// Affiliate
type AffiliateSession struct {
	ID         int    `json:"i"`
	ActualCode string `json:"a"`
}

func (c *ClientCookie) GetCart() int {
	if c.CustomerID > 0 {
		return c.CustomerCart
	}
	return c.GuestCart
}

func (c *ClientCookie) SetCart(cartID int) {
	if c.CustomerID > 0 {
		c.CustomerCart = cartID
	} else {
		c.GuestCart = cartID
	}
}
