package models

import "time"

type Discount struct {
	ID                int    `gorm:"primaryKey"`
	DiscountCode      string `gorm:"unique"`
	Created           time.Time
	IsPercentageOff   bool
	PercentageOff     float64
	IsDollarsOff      bool
	DollarsOff        int
	HasMinSubtotal    bool
	MinSubtotal       int
	Stacks            bool
	AppliesToAllUsers bool
}

type DiscountUser struct {
	DiscountID int `gorm:"primaryKey;index"`
	CustomerID int `gorm:"primaryKey;index"`
}

type GiftCard struct {
	ID            int    `gorm:"primaryKey"`
	UUIDCode      string `gorm:"unique;index"`
	Created       time.Time
	Status        string
	OriginalCents int
	LeftoverCents int
	ShortMessage  string
}
