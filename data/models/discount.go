package models

import "time"

type Discount struct {
	ID               int    `gorm:"primaryKey"`
	DiscountCode     string `gorm:"unique;index"`
	Status           string // Active, Deactivated (Expired)
	Created          time.Time
	Deactivated      time.Time
	Expired          time.Time
	IsPercentageOff  bool
	PercentageOff    float64
	IsDollarsOff     bool // DNE Now
	DollarsOff       int  // DNE Now
	HasMaxUses       bool
	MaxUses          int
	Uses             int
	HasMinSubtotal   bool
	MinSubtotal      int
	Stacks           bool // DNE Now
	AppliesToAllAny  bool
	SingleCustomerID int
	ShortMessage     string
	HasUserList      bool
}

type DiscountUser struct {
	DiscountID int `gorm:"primaryKey;index"`
	CustomerID int `gorm:"primaryKey;index"`
	Uses       int
}

type GiftCard struct {
	ID            int    `gorm:"primaryKey"`
	IDCode        string `gorm:"unique;index"`
	Created       time.Time
	Activated     time.Time
	Spent         time.Time
	Expired       time.Time
	Status        string // Draft, Active, Spent (Expired)
	OriginalCents int
	LeftoverCents int
	ShortMessage  string
}
