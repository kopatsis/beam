package models

import (
	"time"
)

// Can sort by created asds, last updated asds, product count asds
type CustomList struct {
	ID          int `gorm:"primaryKey"`
	CustomerID  int `gorm:"index"`
	Title       string
	Created     time.Time
	LastUpdated time.Time
}

type CustomListLine struct {
	ID           int `gorm:"primaryKey"`
	CustomListID int `gorm:"index"`
	CustomerID   int `gorm:"index"`
	VariantID    int `gorm:"index"`
	ProductID    int `gorm:"index"`
	AddDate      time.Time
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
