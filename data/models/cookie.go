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
	Store       string    `json:"s"`
	CustomerID  int       `json:"c"`
	CustomerSet time.Time `json:"t"`
	GuestID     string    `json:"g"`
}
