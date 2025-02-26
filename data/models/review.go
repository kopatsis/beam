package models

import (
	"time"

	"github.com/lib/pq"
)

type Review struct {
	PK          int       `gorm:"primaryKey"`
	CustomerID  int       `gorm:"index"`
	ProductID   int       `gorm:"index"`
	CreatedAt   time.Time `gorm:"index"`
	Status      string    `gorm:"index"`
	DisplayName string
	LastEdited  time.Time
	Stars       int
	JustStar    bool
	Subject     string
	Body        string
	ImageURLs   pq.StringArray
	Helpful     int
	Unhelpful   int
}

type ReviewFeedback struct {
	ReviewID   int `gorm:"primaryKey"`
	CustomerID int `gorm:"primaryKey"`
	Assigned   time.Time
	Helpful    bool
}
