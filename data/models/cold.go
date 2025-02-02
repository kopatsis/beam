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
	ID            string `gorm:"primaryKey"`
	CustomerID    int    `gorm:"index"`
	GuestID       string `gorm:"index"`
	CreatedAt     time.Time
	Referrer      string
	IPAddress     string
	InitialRoute  string
	IsAffiliate   bool
	AffiliateCode string
	AffiliateID   int
	SpecialStatus string
	City          string
	Country       string
	Browser       string
	OS            string
	Platform      string
	Mobile        bool
	Bot           bool
	FromQR        bool
}

type SessionLine struct {
	ID        int    `gorm:"primaryKey"`
	SessionID string `gorm:"index"`
	Route     string
	Accessed  time.Time
	AnyError  bool
	ErrorSt   string
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
	CreatedAt   time.Time
}

type AffiliateSale struct {
	ID          int    `gorm:"primaryKey"`
	AffiliateID int    `gorm:"index"`
	Code        string `gorm:"index"`
	SessionID   string `gorm:"index"`
	OrderID     string `gorm:"index"`
	CreatedAt   time.Time
}
