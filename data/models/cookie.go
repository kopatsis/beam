package models

import "time"

type ServerCookie struct {
	Store            string    `json:"s"`
	CustomerID       int       `json:"c"`
	LastForcedLogout time.Time `json:"l"`
	Archived         bool      `json:"a"`
	Incomplete       bool      `json:"i"`
}

// Client
type ClientCookie struct {
	Store         string    `json:"s"`
	CustomerID    int       `json:"c"`
	CustomerSet   time.Time `json:"t"`
	GuestID       string    `json:"g"`
	CustomerCart  int       `json:"a"`
	GuestCart     int       `json:"r"`
	OtherCurrency bool      `json:"o"`
	Currency      string    `json:"u"`
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

// For 2FA between steps
type TwoFactorCookie struct {
	TwoFactorCode string    `json:"t"`
	CustomerID    int       `json:"c"`
	Set           time.Time `json:"s"`
}

// Permanent as opposed to session
type DeviceCookie struct {
	DeviceID string `json:"d"`
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
