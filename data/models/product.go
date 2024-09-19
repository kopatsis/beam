package models

import "time"

type Product struct {
	ID          uint    `gorm:"primaryKey"`
	Name        string  `gorm:"size:255"`
	Description string  `gorm:"type:text"`
	Price       float64 `gorm:"type:decimal(10,2)"`
	Stock       int
	Category    string `gorm:"size:255"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
