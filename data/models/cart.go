package models

import (
	"time"
)

type Cart struct {
	ID             int    `gorm:"primaryKey"`
	CustomerID     int    `gorm:"index"`
	GuestID        string `gorm:"index"`
	DateCreated    time.Time
	DateModified   time.Time
	LastRetrieved  time.Time
	Status         string
	EverCheckedOut bool
}

type CartLine struct {
	ID              int `gorm:"primaryKey"`
	CartID          int `gorm:"index"`
	VariantID       int `gorm:"index"`
	ProductID       int `gorm:"index"`
	Quantity        int
	NonDiscPrice    int
	Price           int
	IsGiftCard      bool
	GiftCardCode    string
	GiftCardMessage string
}
