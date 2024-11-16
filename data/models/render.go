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
	First  []VariantBlock
	Second []VariantBlock
	Third  []VariantBlock
}

type ProductRender struct {
	Price     string
	VariantID int
	Inventory int
	VarImage  string
	FullName  string
	Blocks    AllVariants
}
