package models

type TagMap struct {
	ToURL   map[string]string `json:"k"`
	FromURL map[string]string `json:"v"`
}

type StoreNames struct {
	ToDomain   map[string]string `json:"t"`
	FromDomain map[string]string `json:"f"`
}

type AllFilters struct {
	Items []FilterBlock `json:"i"`
}

type FilterBlock struct {
	Key    string   `json:"k"`
	Values []string `json:"i"`
}

type TotalFilters struct {
	All map[string]AllFilters `json:"a"`
}

type TotalTags struct {
	All map[string]TagMap `json:"a"`
}

type SideBar struct {
	Groups []SideBarGroup
}

type SideBarGroup struct {
	Key  string
	Rows []SideBarRow
}

type SideBarRow struct {
	Name     string
	Link     string
	Selected bool
}
