package models

import "net/url"

// Full collection equivalent info
type CollectionRender struct {
	Products []ProductInfo
	URL      url.Values
	SideBar  SideBar
	TopWords TopWords
	Paging   Paging
}

// For sidebar filtering
type SideBar struct {
	Groups []SideBarGroup
}

type SideBarGroup struct {
	Key  string
	Rows []SideBarRow
}

type SideBarRow struct {
	Name     string
	Link     url.Values
	Selected bool
}

// For top of page wording
type TopWords struct {
	Query      string
	Collection string
	Lines      []string
}

type Paging struct {
	Page      int
	PageLeft  int
	PageRight int
	LeftURL   url.Values
	RightURL  url.Values
}

// For each variant block in a row
type VariantBlock struct {
	Name      string
	VariantID int
	Selected  bool
	Stocked   bool
}

type AllVariants struct {
	First     []VariantBlock
	FirstKey  string
	Second    []VariantBlock
	SecondKey string
	Third     []VariantBlock
	ThirdKey  string
}

type ProductRender struct {
	Price       string
	CompareAt   string
	VariantID   int
	Inventory   int
	VarImage    string
	FullName    string
	HasVariants bool
	Blocks      AllVariants
}

// Cart
type CartRender struct {
	CartError   string
	LineError   string
	Empty       bool
	SumQuantity int
	Subtotal    int
	Cart        Cart
	CartLines   []CartLineRender
}

type CartLineRender struct {
	ActualLine    CartLine
	QuantityMaxed bool
	Subtotal      int
}

// Payment
type PaymentMethodStripe struct {
	ID       string
	CardType string
	Last4    string
	ExpMonth int64
	ExpYear  int64
}

// Gift Card
type GiftCardRender struct {
	GiftCard GiftCard
	Expired  bool
}

// List Renders:
type FavesLineRender struct {
	Found     bool
	FavesLine FavesLine
	Variant   LimitedVariantRedis
}

type SavesLineRender struct {
	Found     bool
	SavesLine SavesList
	Variant   LimitedVariantRedis
}

type LastOrderLineRender struct {
	Found   bool
	LOLine  LastOrdersList
	Variant LimitedVariantRedis
}

type FavesListRender struct {
	NoData bool
	Data   []*FavesLineRender
	Prev   bool
	Next   bool
}

type SavesListRender struct {
	NoData bool
	Data   []*SavesLineRender
	Prev   bool
	Next   bool
}

type LastOrderListRender struct {
	NoData bool
	Data   []*LastOrderLineRender
	Prev   bool
	Next   bool
}
