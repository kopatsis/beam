package models

import "sort"

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

type CodeBlock struct {
	Name string `json:"Name"`
	Code string `json:"Code"`
}

type CountryCodes struct {
	List []CodeBlock `json:"List"`
}

type CurrencyCodes struct {
	List []CodeBlock `json:"List"`
}

type StateCodes struct {
	US []CodeBlock `json:"US"`
	CA []CodeBlock `json:"CA"`
	MX []CodeBlock `json:"MX"`
	AU []CodeBlock `json:"AU"`
}

var sizeOrder = map[string]int{
	"XXXS": 0,
	"XXS":  1,
	"XS":   2,
	"S":    3,
	"M":    4,
	"L":    5,
	"XL":   6,
	"XXL":  7,
	"2XL":  8,
	"XXXL": 9,
	"3XL":  10,
	"4XL":  11,
	"5XL":  12,
	"6XL":  13,
}

func SortSizes(sizes []string) {
	sort.Slice(sizes, func(i, j int) bool {
		sI, ok := sizeOrder[sizes[i]]
		if !ok {
			sI = 14
		}
		sJ, ok := sizeOrder[sizes[j]]
		if !ok {
			sJ = 14
		}
		return sI < sJ
	})
}

func (af *AllFilters) Sort() {
	sort.Slice(af.Items, func(i, j int) bool {
		order := map[string]int{"Collection": -2, "Product Type": -1}
		oI, okI := order[af.Items[i].Key]
		oJ, okJ := order[af.Items[j].Key]
		if okI && okJ {
			return oI < oJ
		}
		if okI {
			return true
		}
		if okJ {
			return false
		}
		return af.Items[i].Key < af.Items[j].Key
	})

	for i := range af.Items {
		if af.Items[i].Key == "Size" {
			SortSizes(af.Items[i].Values)
		} else {
			sort.Strings(af.Items[i].Values)
		}
	}
}
