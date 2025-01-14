package models

type Review struct {
	PK        int `gorm:"primaryKey"`
	UserID    int `gorm:"index"`
	ProductID int `gorm:"index"`
	Stars     int
	JustStar  bool
	Subject   string
	Body      string
}
