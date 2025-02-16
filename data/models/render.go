package models

import (
	"net/url"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PriceRender struct {
	DollarPrice    string
	IsOtherPrice   bool
	OtherPrice     string
	OtherPriceCode string
}

// Full collection equivalent info
type CollectionRender struct {
	Products       []ProductInfo
	PricedProducts []ProductWithPriceRender
	URL            url.Values
	SideBar        SideBar
	TopWords       TopWords
	Paging         Paging
}

type ProductWithPriceRender struct {
	Product     ProductInfo
	PriceRender PriceRender
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
	Price           int
	CompareAt       int
	VariantID       int
	Inventory       int
	VarImage        string
	FullName        string
	HasVariants     bool
	Blocks          AllVariants
	PriceRender     PriceRender
	CompareAtRender PriceRender
}

// Cart
type CartRender struct {
	CartError   string
	LineError   string
	Empty       bool
	SumQuantity int
	Subtotal    int
	PriceRender PriceRender
	Cart        Cart
	CartLines   []CartLineRender
}

type CartLineRender struct {
	ActualLine    CartLine
	Variant       LimitedVariantRedis
	QuantityMaxed bool
	Subtotal      int
	PriceRender   PriceRender
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
	Count  int
	NoData bool
	Data   []*FavesLineRender
	Prev   bool
	Next   bool
}

type SavesListRender struct {
	Count  int
	NoData bool
	Data   []*SavesLineRender
	Prev   bool
	Next   bool
}

type LastOrderListRender struct {
	Count  int
	NoData bool
	Data   []*LastOrderLineRender
	Prev   bool
	Next   bool
}

type CustomListLineRender struct {
	Found      bool
	CustomLine CustomListLine
	Variant    LimitedVariantRedis
}

type CustomListRender struct {
	Count      int
	NoData     bool
	CustomList *CustomList
	Data       []*CustomListLineRender
	Prev       bool
	Next       bool
}

type CustomListForVariant struct {
	HasVar     bool
	CustomList CustomList
}

type AllListsForVariant struct {
	VariantID     int
	FavesHasVar   bool
	CanAddAnother bool
	Customs       []CustomListForVariant
}

func (a *AllListsForVariant) Sort() {
	sort.Slice(a.Customs, func(i, j int) bool {
		return a.Customs[i].CustomList.LastUpdated.Before(a.Customs[j].CustomList.LastUpdated)
	})
}

type CustomListRenderBrief struct {
	Count      int
	CustomList CustomList
}

type AllCustomLists struct {
	Lists []CustomListRenderBrief
}

func (a *AllCustomLists) SortBy(field string, desc bool) {
	sort.Slice(a.Lists, func(i, j int) bool {
		var less bool

		switch field {
		case "created_at":
			less = a.Lists[i].CustomList.Created.Before(a.Lists[j].CustomList.Created)
		case "length":
			less = a.Lists[i].Count < a.Lists[j].Count
		default:
			less = a.Lists[i].CustomList.LastUpdated.Before(a.Lists[j].CustomList.LastUpdated)
		}

		if desc {
			return !less
		}
		return less
	})
}

type OrderSummary struct {
	ID          primitive.ObjectID `bson:"_id" json:"id"`
	CustomerID  int                `bson:"customer_id" json:"customer_id"`
	Status      string             `bson:"status" json:"status"`
	DateCreated time.Time          `bson:"date_created" json:"date_created"`
	Subtotal    int                `bson:"subtotal" json:"subtotal"`
	Total       int                `bson:"total" json:"total"`
	Tags        []string           `bson:"tags" json:"tags"`
}

type OrderRender struct {
	Orders     []*Order
	Previous   bool
	Next       bool
	Page       int
	SortColumn string
	Descending bool
}

type ReviewPageRender struct {
	AllReviews []*Review
	CustReview *Review
	Previous   bool
	Next       bool
	Page       int
	SortColumn string
	Descending bool
}

type ComparablesRender struct {
	Handle      string
	Title       string
	ImageURL    string
	Price       int
	PriceRender PriceRender
	Inventory   int
	AvgRate     float64
	RateCt      int
}

type DraftOrderRender struct {
	DraftOrder       *DraftOrder
	TotalPriceRender PriceRender
}
