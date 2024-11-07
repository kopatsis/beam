package models

import "time"

type User struct {
	UID       string `gorm:"primaryKey"`
	Name      string `gorm:"size:255"`
	Address   string `gorm:"size:500"`
	BirthDate time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
