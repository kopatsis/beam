package models

import "time"

type Review struct {
	PK         int       `gorm:"primaryKey"`
	UserID     int       `gorm:"index"`
	ProductID  int       `gorm:"index"`
	CreatedAt  time.Time `gorm:"index"`
	Status     string    `gorm:"index"`
	LastEdited time.Time
	Stars      int
	JustStar   bool
	Subject    string
	Body       string
}
