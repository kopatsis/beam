package models

import "time"

type ProductRedis struct {
	PK             int               `json:"pk"`
	PrintfulID     map[string]int    `json:"pid"`
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
}

type ComparableRedis struct {
	Handle string `json:"h"`
	Title  string `json:"t"`
	Image  string `json:"iu"`
}

type VariantRedis struct {
	PK              int            `json:"pk"`
	ProductID       int            `json:"pr"`
	PrintfulID      map[string]int `json:"pid"`
	SKU             string         `json:"sku"`
	Var1Value       string         `json:"vk1"`
	Var2Value       string         `json:"vk2,omitempty"` // Optional
	Var3Value       string         `json:"vk3,omitempty"` // Optional
	Price           int            `json:"p"`
	Quantity        int            `json:"q"`
	VariantImageURL string         `json:"vu"`
	Barcode         string         `json:"bc"`
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
