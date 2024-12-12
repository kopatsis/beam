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
