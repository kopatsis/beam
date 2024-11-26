package models

import (
	"github.com/lib/pq"
)

type Customer struct {
	ID                       int    `gorm:"primaryKey"`
	FirebaseUID              string `gorm:"unique"`
	StripeID                 string `gorm:"unique"`
	DefaultName              string
	Email                    string `gorm:"unique"`
	EmailSubbed              bool
	DefaultShippingContactID int            `gorm:"index"`
	Tags                     pq.StringArray `gorm:"type:text[]"`
	PhoneNumber              *string
	Status                   string
	OrderCount               int
	AutodiscountPctOff       float64
	Notes                    string
}

type Contact struct {
	ID             int `gorm:"primaryKey"`
	CustomerID     int `gorm:"index"`
	FirstName      string
	LastName       *string
	PhoneNumber    *string
	StreetAddress1 string
	StreetAddress2 *string
	City           string
	ProvinceState  string
	ZipCode        string
	Country        string
}
