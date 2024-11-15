package models

import "net/url"

// Full collection equivalent info
type CollectionRender struct {
	Products  []ProductInfo
	URL       url.Values
	SideBar   SideBar
	TopWords  TopWords
	Page      int
	PageLeft  int
	PageRight int
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
