package models

import (
	"time"
)

type Cart struct {
	ID             int    `gorm:"primaryKey"`
	CustomerID     int    `gorm:"index"`
	GuestID        string `gorm:"index"`
	DateStarted    time.Time
	Status         string
	EverCheckedOut bool
}

type CartLine struct {
	ID            int `gorm:"primaryKey"`
	CartID        int `gorm:"index"`
	ProductID     int `gorm:"index"`
	VariantID     int `gorm:"index"`
	ProductHandle string
	ImageURL      string
	ProductTitle  string
	Variant1Key   string
	Variant1Value string
	Variant2Key   *string
	Variant2Value *string
	Variant3Key   *string
	Variant3Value *string
	Quantity      int
	Price         int
}
