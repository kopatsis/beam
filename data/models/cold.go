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

type Session struct {
	ID            int    `gorm:"primaryKey"`
	CustomerID    int    `gorm:"index"`
	GuestID       string `gorm:"index"`
	CreatedAt     time.Time
	Referrer      string
	IPAddress     string
	InitialRoute  string
	IsAffiliate   bool
	AffiliateCode string
	SpecialStatus string
}

type SessionLine struct {
	ID        int    `gorm:"primaryKey"`
	SessionID string `gorm:"index"`
	Route     string
	Accessed  time.Time
	AnyError  bool
	ErrorSt   string
}
