package models

import "time"

type Discount struct {
	ID                int    `gorm:"primaryKey"`
	DiscountCode      string `gorm:"unique"`
	Status            string
	Created           time.Time
	IsPercentageOff   bool
	PercentageOff     float64
	IsDollarsOff      bool
	DollarsOff        int
	OneTime           bool
	Uses              int
	HasMinSubtotal    bool
	MinSubtotal       int
	Stacks            bool
	AppliesToAllUsers bool
	SingleCustomerID  int
}

type DiscountUser struct {
	DiscountID int `gorm:"primaryKey;index"`
	CustomerID int `gorm:"primaryKey;index"`
}

type GiftCard struct {
	ID            int    `gorm:"primaryKey"`
	IDCode        string `gorm:"unique;index"`
	Created       time.Time
	Activated     time.Time
	Spent         time.Time
	Expired       time.Time
	Status        string // Draft, Active, Staged, Spent, Expired
	OriginalCents int
	LeftoverCents int
	ShortMessage  string
}
