package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Notification struct {
	ID                  primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CustomerID          int                `bson:"customer_id" json:"customer_id"`
	CreatedAt           string             `bson:"created_at" json:"created_at"`
	CompletedAt         string             `bson:"completed_at" json:"completed_at"`
	RemovedAt           string             `bson:"removed_at" json:"removed_at"`
	EventClassification string             `bson:"event_classification" json:"event_classification"`
	EventDescription    string             `bson:"event_description" json:"event_description"`
	OrderID             *string            `bson:"order_id,omitempty" json:"order_id,omitempty"`
	ProductID           *string            `bson:"product_id,omitempty" json:"product_id,omitempty"`
	ListID              *string            `bson:"list_id,omitempty" json:"list_id,omitempty"`
	CartID              *string            `bson:"cart_id,omitempty" json:"cart_id,omitempty"`
	CollectionID        *string            `bson:"collection_id,omitempty" json:"collection_id,omitempty"`
	DiscountID          *string            `bson:"discount_id,omitempty" json:"discount_id,omitempty"`
	ReviewID            *string            `bson:"review_id,omitempty" json:"review_id,omitempty"`
	SpecialNote         string             `bson:"special_note" json:"special_note"`
	Tags                []string           `bson:"tags" json:"tags"`
	Status              string             `bson:"status" json:"status"` // Enum: "created", "completed", "removed"
	SpecificEmail       string             `bson:"specific_email" json:"specific_email"`
	SpecificPhone       string             `bson:"specific_phone" json:"specific_phone"`
	Email               bool               `bson:"email" json:"email"`
	Phone               bool               `bson:"phone" json:"phone"`
}

type Event struct {
	ID                  primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CustomerID          int                `bson:"customer_id" json:"customer_id"`
	GuestID             string             `bson:"guest_id" json:"guest_id"`
	SessionID           string             `bson:"session_id" json:"session_id"`
	Timestamp           time.Time          `bson:"timestamp" json:"timestamp"`
	EventClassification string             `bson:"event_classification" json:"event_classification"` // Super General, like "List"
	EventDescription    string             `bson:"event_description" json:"event_description"`       // More specific function, like "Delete Faves Line"
	EventDetails        string             `bson:"event_details" json:"event_details"`               // Exact, like "Successfully deleted faves line no error, optional"
	OrderID             *string            `bson:"order_id,omitempty" json:"order_id,omitempty"`
	DraftOrderID        *string            `bson:"draftorder_id,omitempty" json:"draftorder_id,omitempty"`
	ProductID           *string            `bson:"product_id,omitempty" json:"product_id,omitempty"`
	VariantID           *string            `bson:"variant_id,omitempty" json:"variant_id,omitempty"`
	SavesID             *string            `bson:"saves_id,omitempty" json:"saves_id,omitempty"`
	FavesID             *string            `bson:"faves_id,omitempty" json:"faves_id,omitempty"`
	LOListID            *string            `bson:"lolist_id,omitempty" json:"lolist_id,omitempty"`
	CartID              *string            `bson:"cart_id,omitempty" json:"cart_id,omitempty"`
	CartLineID          *string            `bson:"cart_line_id,omitempty" json:"cart_line_id,omitempty"`
	DiscountID          *string            `bson:"discount_id,omitempty" json:"discount_id,omitempty"`
	GiftCardID          *string            `bson:"giftcard_id,omitempty" json:"giftcard_id,omitempty"`
	SpecialNote         string             `bson:"special_note" json:"special_note"`
	Tags                []string           `bson:"tags" json:"tags"`
	AnyError            bool               `bson:"any_err" json:"any_err"`
	AllErrorsSt         []string           `bson:"errors" json:"errors"`
}

type EventNew struct {
	ID                  primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CustomerID          int                `bson:"customer_id" json:"customer_id"`
	GuestID             string             `bson:"guest_id" json:"guest_id"`
	Timestamp           time.Time          `bson:"timestamp" json:"timestamp"`
	EventClassification string             `bson:"event_classification" json:"event_classification"` // Super General, like "List"
	EventDescription    string             `bson:"event_description" json:"event_description"`       // More specific function, like "Delete Faves Line"
	EventDetails        string             `bson:"event_details" json:"event_details"`               // Exact, like "Successfully deleted faves line no error, optional"
	OrderID             string             `json:"order_id" bson:"order_id"`
	DraftOrderID        string             `json:"draft_order_id" bson:"draft_order_id"`
	ProductID           int                `json:"product_id" bson:"product_id"`
	ProductHandle       string             `json:"product_handle" bson:"product_handle"`
	VariantID           int                `json:"variant_id" bson:"variant_id"`
	SavesID             int                `json:"saves_id" bson:"saves_id"`
	FavesID             int                `json:"faves_id" bson:"faves_id"`
	LastOrderListID     int                `json:"last_order_list_id" bson:"last_order_list_id"`
	CartID              int                `json:"cart_id" bson:"cart_id"`
	CartLineID          int                `json:"cart_line_id" bson:"cart_line_id"`
	DiscountID          int                `json:"discount_id" bson:"discount_id"`
	DiscountCode        string             `json:"discount_code" bson:"discount_code"`
	GiftCardID          int                `json:"gift_card_id" bson:"gift_card_id"`
	GiftCardCode        string             `json:"gift_card_code" bson:"gift_card_code"`
	SessionID           string             `json:"session_id" bson:"session_id"`
	SpecialNote         string             `bson:"special_note" json:"special_note"`
	Tags                []string           `bson:"tags" json:"tags"`
	AnyError            bool               `bson:"any_err" json:"any_err"`
	AllErrorsSt         []string           `bson:"errors" json:"errors"`
}

type EventFinal struct {
	ID               string    `json:"id"`
	Level            string    `json:"level"`
	Store            string    `json:"store"`
	Timestamp        time.Time `json:"timestamp"`
	SessionID        string    `json:"session_id"`
	SessionLineID    string    `json:"session_line_id"`
	CustomerID       int       `json:"customer_id"`
	GuestID          string    `json:"guest_id"`
	ModelName        string    `json:"model_name"`
	FunctionName     string    `json:"function_name"`
	HasError         bool      `json:"has_error"`
	ErrorDescription string    `json:"error_description,omitempty"`
	ErrorValueSt     string    `json:"error_value_st,omitempty"`
	OptionalNote     string    `json:"optional_note,omitempty"`
	OrderID          string    `json:"order_id,omitempty"`
	DraftOrderID     string    `json:"draft_order_id,omitempty"`
	ProductID        int       `json:"product_id,omitempty"`
	ProductHandle    string    `json:"product_handle,omitempty"`
	VariantID        int       `json:"variant_id,omitempty"`
	SavesID          int       `json:"saves_id,omitempty"`
	FavesID          int       `json:"faves_id,omitempty"`
	LastOrderListID  int       `json:"last_order_list_id,omitempty"`
	CartID           int       `json:"cart_id,omitempty"`
	CartLineID       int       `json:"cart_line_id,omitempty"`
	DiscountID       int       `json:"discount_id,omitempty"`
	DiscountCode     string    `json:"discount_code,omitempty"`
	GiftCardID       int       `json:"gift_card_id,omitempty"`
	GiftCardCode     string    `json:"gift_card_code,omitempty"`
	ReviewID         int       `json:"review,omitempty"`
}

type EventPassInFinal struct {
	OrderID         string
	DraftOrderID    string
	ProductID       int
	ProductHandle   string
	VariantID       int
	SavesID         int
	FavesID         int
	LastOrderListID int
	CartID          int
	CartLineID      int
	DiscountID      int
	DiscountCode    string
	GiftCardID      int
	GiftCardCode    string
	ReviewID        int
	SessionID       string
}

type EventIDPassIn struct {
	CustomerID      int
	GuestID         string
	OrderID         string
	DraftOrderID    string
	ProductID       int
	ProductHandle   string
	VariantID       int
	SavesID         int
	FavesID         int
	LastOrderListID int
	CartID          int
	CartLineID      int
	DiscountID      int
	DiscountCode    string
	GiftCardID      int
	GiftCardCode    string
	SessionID       string
}

type Session struct {
	ID            string    `gorm:"primaryKey" json:"i"`
	CustomerID    int       `gorm:"index" json:"c"` // Upon creation
	GuestID       string    `gorm:"index" json:"g"`
	CreatedAt     time.Time `json:"ca"`
	Referrer      string    `json:"r"`
	IPAddress     string    `json:"ip"`
	InitialRoute  string    `json:"ir"`
	FullURL       string    `json:"fu"`
	SpecialStatus string    `json:"ss"`
	City          string    `json:"ct"`
	Country       string    `json:"co"`
	Browser       string    `json:"b"`
	OS            string    `json:"os"`
	Platform      string    `json:"p"`
	Mobile        bool      `json:"m"`
	Bot           bool      `json:"o"`
}

type SessionLine struct {
	ID         string    `gorm:"primaryKey" json:"i"`
	SessionID  string    `gorm:"index" json:"si"`
	CustomerID int       `gorm:"index" json:"ci"` // At the time of line
	BaseRoute  string    `json:"br"`
	FullRoute  string    `json:"fr"`
	FullURL    string    `json:"fu"`
	Accessed   time.Time `json:"a"`
	Ended      time.Time `json:"t"`
}

type Affiliate struct {
	ID        int    `gorm:"primaryKey"`
	Code      string `gorm:"uniqueIndex"`
	Name      string
	Email     string
	CreatedAt time.Time
	LastUsed  time.Time
	Valid     bool
}

type AffiliateLine struct {
	ID          int    `gorm:"primaryKey"`
	AffiliateID int    `gorm:"index"`
	Code        string `gorm:"index"`
	SessionID   string `gorm:"index"`
	Timestamp   time.Time
}

type AffiliateSale struct {
	ID          int    `gorm:"primaryKey"`
	AffiliateID int    `gorm:"index"`
	Code        string `gorm:"index"`
	SessionID   string `gorm:"index"`
	OrderID     string `gorm:"index"`
	Timestamp   time.Time
}

type InventoryAdjustment struct {
	ID              int    `gorm:"primaryKey"`
	ProductID       int    `gorm:"index"`
	VariantID       int    `gorm:"index"`
	PreviousInv     int    // Previous inventory
	EndInv          int    // End inventory
	FromOrder       bool   // ORDER: If originated from successful order
	OrderID         string // ORDER: Order ID
	InitialOrderDec int    // ORDER: (Negative) initial change from order aka qty ordered
	IsReversal      bool   // ORDER: Went through but was cancelled/reversed
	AlwaysUpAdj     bool   // ORDER: Setting to always have inventory occurred
	AlwaysUpInc     int    // ORDER: How much always up incremented by
	FromCommand     bool   // COMMAND: If originated from an Excel command
	CommandID       string // COMMAND: UUID generated for command
	CommandName     string // COMMAND: "+", "-", or "SET"
	CommandValue    int    // (COMMAND: Negative or positive) value for inventory adjustment
}
