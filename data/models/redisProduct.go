package models

import "time"

type ProductRedis struct {
	PK             int            `json:"pk"`
	Handle         string         `json:"h"`
	Store          string         `json:"sr"`
	Title          string         `json:"t"`
	Description    string         `json:"d"`
	Bullets        []string       `json:"b"`
	ImageURL       string         `json:"iu"`
	AltImageURLs   []string       `json:"aiu"`
	Status         string         `json:"s"`
	DateAdded      time.Time      `json:"da"`
	Tags           []string       `json:"tg"`
	Rating         float64        `json:"r"`
	RatingCt       int            `json:"a"`
	Var1Key        string         `json:"vk1"`
	Var2Key        string         `json:"vk2,omitempty"` // Optional
	Var3Key        string         `json:"vk3,omitempty"` // Optional
	SEOTitle       string         `json:"st"`
	SEODescription string         `json:"sd"`
	Variants       []VariantRedis `json:"v"`
	StandardPrice  int            `json:"sp"`
	VolumeDisc     bool           `json:"vd"`
}

type VariantRedis struct {
	PK              int                    `json:"pk"`
	ProductID       int                    `json:"pr"`
	Printful        []OriginalProductRedis `json:"pid"`
	SKU             string                 `json:"sku"`
	Var1Value       string                 `json:"vk1"`
	Var2Value       string                 `json:"vk2,omitempty"` // Optional
	Var3Value       string                 `json:"vk3,omitempty"` // Optional
	Price           int                    `json:"p"`
	CompareAtPrice  int                    `json:"c"`
	Quantity        int                    `json:"q"`
	VariantImageURL string                 `json:"vu"`
	Barcode         string                 `json:"bc"`
	AlwaysUp        bool                   `json:"a"`
}

type ProductInfo struct {
	ID         int       `json:"id"`
	Handle     string    `json:"h"`
	Title      string    `json:"t"`
	DateAdded  time.Time `json:"da"`
	ImageURL   string    `json:"m"`
	Sales      int       `json:"s"`
	Price      int       `json:"p"`
	Inventory  int       `json:"i"`
	AvgRate    float64   `json:"r"`
	RateCt     int       `json:"a"`
	Tags       []string  `json:"tg"`
	Var1Key    string    `json:"vk1"`
	Var2Key    string    `json:"vk2,omitempty"` // Optional
	Var3Key    string    `json:"vk3,omitempty"` // Optional
	Var1Values []string  `json:"vv1"`
	Var2Values []string  `json:"vv2"`
	Var3Values []string  `json:"vv3"`
	SKUs       []string  `json:"ss"`
}

type OriginalProductRedis struct {
	Quantity          int                 `json:"q"`
	ProductID         string              `json:"p"`
	VariantID         string              `json:"v"`
	ExternalProductID string              `json:"ep"`
	ExternalVariantID string              `json:"ev"`
	OriginalProductID string              `json:"op"`
	OriginalVariantID string              `json:"ov"`
	FullVariantName   string              `json:"f"`
	SKU               string              `json:"s"`
	RetailPrice       int                 `json:"rp"`
	Fulfillment       *SubLineFulfillment `json:"l"`
}

type LimitedVariantRedis struct {
	VariantID       int    `json:"v"`
	ProductID       int    `json:"p"`
	Handle          string `json:"h"`
	Title           string `json:"t"`
	VariantImageURL string `json:"i"`
	Var1Key         string `json:"k1"`
	Var2Key         string `json:"k2,omitempty"` // Optional
	Var3Key         string `json:"k3,omitempty"` // Optional
	Var1Value       string `json:"v1"`
	Var2Value       string `json:"v2,omitempty"` // Optional
	Var3Value       string `json:"v3,omitempty"` // Optional
	Price           int    `json:"c"`
}
