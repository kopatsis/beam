package models

import (
	"time"
)

type List struct {
	ID          int `gorm:"primaryKey"`
	CustomerID  int `gorm:"index"`
	Title       string
	DateStarted time.Time
	ItemCount   int
}

type ListLine struct {
	ID            int `gorm:"primaryKey"`
	LineID        int `gorm:"index"`
	ProductID     int `gorm:"index"`
	VariantID     int `gorm:"index"`
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
	DateAdded     time.Time
}

type FavesLine struct {
	ID         int `gorm:"primaryKey"`
	CustomerID int `gorm:"index"`
	VariantID  int `gorm:"index"`
	ProductID  int `gorm:"index"`
	AddDate    time.Time
}

type SavesList struct {
	ID         int `gorm:"primaryKey"`
	CustomerID int `gorm:"index"`
	VariantID  int `gorm:"index"`
	ProductID  int `gorm:"index"`
	AddDate    time.Time
}

type LastOrdersList struct {
	ID          int `gorm:"primaryKey"`
	CustomerID  int `gorm:"index"`
	VariantID   int `gorm:"index"`
	ProductID   int `gorm:"index"`
	LastOrder   time.Time
	LastOrderID string
}
