package models

import "time"

type ServerCookie struct {
	Store      string    `json:"s"`
	FirebaseID string    `json:"f"`
	CustomerID int       `json:"c"`
	LastReset  time.Time `json:"l"`
	Archived   bool      `json:"a"`
}

type ClientCookie struct {
	Store        string    `json:"s"`
	CustomerID   int       `json:"c"`
	CustomerSet  time.Time `json:"t"`
	GuestID      string    `json:"g"`
	CustomerCart int       `json:"a"`
	GuestCart    int       `json:"r"`
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
