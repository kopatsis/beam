package models

import (
	"time"

	"github.com/lib/pq"
)

// Product represents the structure for the PRODUCT table.
type Product struct {
	PK             int            `gorm:"primaryKey"`
	Store          string         `gorm:"type:varchar(16)"`
	Handle         string         `gorm:"type:varchar(512)"`
	Title          string         `gorm:"type:varchar(512)"`
	Description    string         `gorm:"type:text"`
	Bullets        pq.StringArray `gorm:"type:text[]"`
	ImageURL       string         `gorm:"type:varchar(1024)"`
	AltImageURLs   pq.StringArray `gorm:"type:text[]"`
	Status         string         `gorm:"type:varchar(512)"`
	DateAdded      time.Time      `gorm:"type:timestamp"`
	Tags           pq.StringArray `gorm:"type:text[]"`
	Rating         float64        `gorm:"type:float8"`
	Redirect       *string        `gorm:"type:varchar(512)"`
	Variant1Key    string         `gorm:"type:varchar(256)"`
	Variant2Key    *string        `gorm:"type:varchar(256)"`
	Variant3Key    *string        `gorm:"type:varchar(256)"`
	SEOTitle       string         `gorm:"type:varchar(512)"`
	SEODescription string         `gorm:"type:text"`
	StandardPrice  int            `gorm:"type:int"`
	VolumeDisc     bool
}

// Comparable represents the structure for the COMPARABLE table.
type Comparable struct {
	PKFKProductID1 int `gorm:"primaryKey"`
	PKFKProductID2 int `gorm:"primaryKey"`
}

// PreComprable
type PreComprableSQL struct {
	Handle1 string
	Handle2 string
}

// Variant represents the structure for the VARIANT table.
type Variant struct {
	PK              int     `gorm:"primaryKey"`
	ProductID       int     `gorm:"index"`
	SKU             string  `gorm:"type:varchar(256)"`
	Variant1Value   string  `gorm:"type:varchar(256)"`
	Variant2Value   *string `gorm:"type:varchar(256)"`
	Variant3Value   *string `gorm:"type:varchar(256)"`
	Price           int     `gorm:"type:int"`
	CompareAtPrice  int     `gorm:"type:int"`
	Quantity        int     `gorm:"type:int"`
	VariantImageURL string  `gorm:"type:varchar(1024)"`
	VariantBarcode  string  `gorm:"type:varchar(256)"`
}

type OriginalProduct struct {
	PK                int `gorm:"primaryKey"`
	ActualVariantID   int `gorm:"index"`
	Quantity          int
	ProductID         string
	ExternalProductID string
	FullVariantName   string
	VariantID         string
	ExternalVariantID string
	RetailPrice       int
	OriginalProductID string
	OriginalVariantID string
}
