package models

import (
	"time"

	"github.com/lib/pq"
)

type Customer struct {
	ID                       int    `gorm:"primaryKey"`
	FirebaseUID              string `gorm:"unique"`
	StripeID                 string `gorm:"unique"`
	DefaultName              string
	LastName                 string
	Email                    string `gorm:"unique"`
	PasswordHash             string
	EmailSubbed              bool
	EmailVerified            bool
	DefaultShippingContactID int            `gorm:"index"`
	Tags                     pq.StringArray `gorm:"type:text[]"`
	Created                  time.Time
	PhoneNumber              *string
	Status                   string // Active, Archived
	AutodiscountPctOff       float64
	Notes                    string
	LastReset                time.Time
	Uses2FA                  bool
}

type CustomerPost struct {
	DefaultName string  `json:"default_name" validate:"required"`
	Email       string  `json:"email" validate:"required,email"`
	EmailSubbed bool    `json:"email_subbed" validate:"required"`
	PhoneNumber *string `json:"phone_number,omitempty"`
}

type Contact struct {
	ID             int     `gorm:"primaryKey" json:"id" bson:"id"`
	CustomerID     int     `gorm:"index" json:"customer_id" bson:"customer_id"`
	FirstName      string  `json:"first_name" bson:"first_name"`
	LastName       *string `json:"last_name,omitempty" bson:"last_name,omitempty"`
	Company        *string `json:"company,omitempty" bson:"company,omitempty"`
	PhoneNumber    *string `json:"phone_number,omitempty" bson:"phone_number,omitempty"`
	StreetAddress1 string  `json:"street_address_1" bson:"street_address_1"`
	StreetAddress2 *string `json:"street_address_2,omitempty" bson:"street_address_2,omitempty"`
	City           string  `json:"city" bson:"city"`
	ProvinceState  string  `json:"province_state" bson:"province_state"`
	StateCode      string  `json:"state_code" bson:"state_code"`
	ZipCode        string  `json:"zip_code" bson:"zip_code"`
	Country        string  `json:"country" bson:"country"`
	CountryCode    string  `json:"country_code" bson:"country_code"`
}
