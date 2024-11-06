package models

type Discount struct {
	ID              int    `gorm:"primaryKey"`
	DiscountCode    string `gorm:"unique"`
	IsPercentageOff bool
	PercentageOff   float64
	IsDollarsOff    bool
	DollarsOff      int
	HasMinSubtotal  bool
	MinSubtotal     int
	Stacks          bool
	AppliesToAll    bool
}

type DiscountUser struct {
	DiscountID int `gorm:"primaryKey;index"`
	CustomerID int `gorm:"primaryKey;index"`
}
