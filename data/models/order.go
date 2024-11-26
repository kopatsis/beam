package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Order struct {
	ID                      primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PrintfulID              string             `bson:"printful_id" json:"printful_id"`
	CustomerID              int                `bson:"customer_id" json:"customer_id"`
	DraftOrderID            string             `bson:"draft_order_id" json:"draft_order_id"`
	Status                  string             `bson:"status" json:"status"`
	Email                   string             `bson:"email" json:"email"`
	DateCreated             time.Time          `bson:"date_created" json:"date_created"`
	DateCancelled           *time.Time         `bson:"date_cancelled,omitempty" json:"date_cancelled,omitempty"`
	DateProcessedPrintful   *time.Time         `bson:"date_processed_printful,omitempty" json:"date_processed_printful,omitempty"`
	DateShipped             *time.Time         `bson:"date_shipped,omitempty" json:"date_shipped,omitempty"`
	DateDelivered           *time.Time         `bson:"date_delivered,omitempty" json:"date_delivered,omitempty"`
	DateReturnInitiated     *time.Time         `bson:"date_return_initiated,omitempty" json:"date_return_initiated,omitempty"`
	DateReturnCompleted     *time.Time         `bson:"date_return_completed,omitempty" json:"date_return_completed,omitempty"`
	StripePaymentIntentID   string             `bson:"stripe_payment_intent_id" json:"stripe_payment_intent_id"`
	Subtotal                int                `bson:"subtotal" json:"subtotal"`
	Shipping                int                `bson:"shipping" json:"shipping"`
	OrderLevelDiscount      int                `bson:"order_level_discount" json:"order_level_discount"`
	Tax                     int                `bson:"tax" json:"tax"`
	Tip                     int                `bson:"tip" json:"tip"`
	Total                   int                `bson:"total" json:"total"`
	NonStackingDiscountCode string             `bson:"non_stacking_discount_code" json:"non_stacking_discount_code"`
	StackingDiscountCodes   []string           `bson:"stacking_discount_codes" json:"stacking_discount_codes"`
	BillingZip              string             `bson:"billing_contact" json:"billing_contact"`
	ShippingContact         OrderContact       `bson:"shipping_contact" json:"shipping_contact"`
	Lines                   []OrderLine        `bson:"lines" json:"lines"`
	Tags                    []string           `bson:"tags" json:"tags"`
	DeliveryNote            string             `bson:"delivery_note" json:"delivery_note"`
	ShippingIdentification  string             `bson:"shipping_identification" json:"shipping_identification"`
	Guest                   bool               `bson:"guest" json:"guest"`
	GuestID                 *string            `bson:"guest_id,omitempty" json:"guest_id,omitempty"`
	External                bool               `bson:"external" json:"external"`
	ExternalPlatform        string             `bson:"external_platform" json:"external_platform"`
	ExternalID              string             `bson:"external_id" json:"external_id"`
	ShippingCarrier         string             `bson:"shipping_carrier" json:"shipping_carrier"`
	ShippingService         string             `bson:"shipping_service" json:"shipping_service"`
	ShippingTrackingNumber  string             `bson:"shipping_tracking_number" json:"shipping_tracking_number"`
	ShippingTrackingURL     string             `bson:"shipping_tracking_url" json:"shipping_tracking_url"`
}

type DraftOrder struct {
	ID                      primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PrintfulID              string             `bson:"printful_id" json:"printful_id"`
	CustomerID              int                `bson:"customer_id" json:"customer_id"`
	Status                  string             `bson:"status" json:"status"`
	Email                   string             `bson:"email" json:"email"`
	DateCreated             time.Time          `bson:"date_created" json:"date_created"`
	DateAbandoned           *time.Time         `bson:"date_abandoned,omitempty" json:"date_abandoned,omitempty"`
	StripePaymentIntentID   string             `bson:"stripe_payment_intent_id" json:"stripe_payment_intent_id"`
	Subtotal                int                `bson:"subtotal" json:"subtotal"`
	Shipping                int                `bson:"shipping" json:"shipping"`
	OrderLevelDiscount      int                `bson:"order_level_discount" json:"order_level_discount"`
	Tax                     int                `bson:"tax" json:"tax"`
	Tip                     int                `bson:"tip" json:"tip"`
	Total                   int                `bson:"total" json:"total"`
	NonStackingDiscountCode string             `bson:"non_stacking_discount_code" json:"non_stacking_discount_code"`
	StackingDiscountCodes   []string           `bson:"stacking_discount_codes" json:"stacking_discount_codes"`
	BillingZip              string             `bson:"billing_contact" json:"billing_contact"`
	ShippingContact         OrderContact       `bson:"shipping_contact" json:"shipping_contact"`
	Lines                   []OrderLine        `bson:"lines" json:"lines"`
	Tags                    []string           `bson:"tags" json:"tags"`
	DeliveryNote            string             `bson:"delivery_note" json:"delivery_note"`
	Guest                   bool               `bson:"guest" json:"guest"`
	GuestID                 *string            `bson:"guest_id,omitempty" json:"guest_id,omitempty"`
}

type OrderLine struct {
	ImageURL          string `bson:"image_url" json:"image_url"`
	ProductTitle      string `bson:"product_title" json:"product_title"`
	Handle            string `bson:"handle" json:"handle"`
	Variant1Key       string `bson:"variant_1_key" json:"variant_1_key"`
	Variant1Value     string `bson:"variant_1_value" json:"variant_1_value"`
	Variant2Key       string `bson:"variant_2_key" json:"variant_2_key"`
	Variant2Value     string `bson:"variant_2_value" json:"variant_2_value"`
	Variant3Key       string `bson:"variant_3_key" json:"variant_3_key"`
	Variant3Value     string `bson:"variant_3_value" json:"variant_3_value"`
	ProductID         string `bson:"product_id" json:"product_id"`
	VariantID         string `bson:"variant_id" json:"variant_id"`
	Quantity          int    `bson:"quantity" json:"quantity"`
	UndiscountedPrice int    `bson:"undiscounted_price" json:"undiscounted_price"`
	LineLevelDiscount int    `bson:"line_level_discount" json:"line_level_discount"`
	EndPrice          int    `bson:"end_price" json:"end_price"`
	LineTotal         int    `bson:"line_total" json:"line_total"`
}

type OrderContact struct {
	FirstName      string `bson:"first_name" json:"first_name"`
	LastName       string `bson:"last_name" json:"last_name"`
	PhoneNumber    string `bson:"phone_number" json:"phone_number"`
	StreetAddress1 string `bson:"street_address_1" json:"street_address_1"`
	StreetAddress2 string `bson:"street_address_2" json:"street_address_2"`
	City           string `bson:"city" json:"city"`
	ProvinceState  string `bson:"province_state" json:"province_state"`
	ZipCode        string `bson:"zip_code" json:"zip_code"`
	Country        string `bson:"country" json:"country"`
}
