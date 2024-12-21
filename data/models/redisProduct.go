package models

import "time"

type ProductRedis struct {
	PK             int               `json:"pk"`
	Handle         string            `json:"h"`
	Store          string            `json:"sr"`
	Title          string            `json:"t"`
	Description    string            `json:"d"`
	Bullets        []string          `json:"b"`
	ImageURL       string            `json:"iu"`
	AltImageURLs   []string          `json:"aiu"`
	Status         string            `json:"s"`
	DateAdded      time.Time         `json:"da"`
	Tags           []string          `json:"tg"`
	Rating         float64           `json:"r"`
	Var1Key        string            `json:"vk1"`
	Var2Key        string            `json:"vk2,omitempty"` // Optional
	Var3Key        string            `json:"vk3,omitempty"` // Optional
	SEOTitle       string            `json:"st"`
	SEODescription string            `json:"sd"`
	Comparables    []ComparableRedis `json:"c"`
	Variants       []VariantRedis    `json:"v"`
	StandardPrice  int               `json:"sp"`
	VolumeDisc     bool              `json:"vd"`
}

type ComparableRedis struct {
	Handle string `json:"h"`
	Title  string `json:"t"`
	Image  string `json:"iu"`
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
}

type ProductInfo struct {
	ID         int       `json:"id"`
	Handle     string    `json:"h"`
	Title      string    `json:"t"`
	DateAdded  time.Time `json:"da"`
	Sales      int       `json:"s"`
	Price      int       `json:"p"`
	Inventory  int       `json:"i"`
	AvgRate    float64   `json:"ar"`
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
	Quantity          int    `json:"q"`
	ProductID         string `json:"p"`
	VariantID         string `json:"v"`
	ExternalProductID string `json:"ep"`
	ExternalVariantID string `json:"ev"`
	OriginalProductID string `json:"op"`
	OriginalVariantID string `json:"ov"`
	FullVariantName   string `json:"f"`
	RetailPrice       int    `json:"rp"`
}
