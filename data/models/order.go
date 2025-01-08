package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Order struct {
	ID                     primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PrintfulID             string             `bson:"printful_id" json:"printful_id"`
	CustomerID             int                `bson:"customer_id" json:"customer_id"`
	DraftOrderID           string             `bson:"draft_order_id" json:"draft_order_id"`
	Status                 string             `bson:"status" json:"status"`
	Email                  string             `bson:"email" json:"email"`
	FirstName              string             `bson:"fname" json:"fname"`
	LastName               string             `bson:"lname" json:"lname"`
	DateCreated            time.Time          `bson:"date_created" json:"date_created"`
	DateCancelled          *time.Time         `bson:"date_cancelled,omitempty" json:"date_cancelled,omitempty"`
	DateProcessedPrintful  *time.Time         `bson:"date_processed_printful,omitempty" json:"date_processed_printful,omitempty"`
	DateShipped            *time.Time         `bson:"date_shipped,omitempty" json:"date_shipped,omitempty"`
	DateDelivered          *time.Time         `bson:"date_delivered,omitempty" json:"date_delivered,omitempty"`
	DateReturnInitiated    *time.Time         `bson:"date_return_initiated,omitempty" json:"date_return_initiated,omitempty"`
	DateReturnCompleted    *time.Time         `bson:"date_return_completed,omitempty" json:"date_return_completed,omitempty"`
	StripePaymentIntentID  string             `bson:"stripe_payment_intent_id" json:"stripe_payment_intent_id"`
	Subtotal               int                `bson:"subtotal" json:"subtotal"`
	OrderLevelDiscount     int                `bson:"order_level_discount" json:"order_level_discount"`
	PostDiscountTotal      int                `bson:"post_disc_total" json:"post_disc_total"`
	Shipping               int                `bson:"shipping" json:"shipping"`
	Tax                    int                `bson:"tax" json:"tax"`
	PostTaxTotal           int                `bson:"post_tax_total" json:"post_tax_total"`
	Tip                    int                `bson:"tip" json:"tip"`
	PreGiftCardTotal       int                `bson:"pgc_total" json:"pgc_total"`
	GiftCardSum            int                `bson:"gc_sum" json:"gc_sum"`
	Total                  int                `bson:"total" json:"total"`
	OrderDiscount          OrderDiscount      `bson:"non_stacking_discount_code" json:"non_stacking_discount_code"`
	ShippingContact        *Contact           `bson:"shipping_contact" json:"shipping_contact"`
	Lines                  []OrderLine        `bson:"lines" json:"lines"`
	Tags                   []string           `bson:"tags" json:"tags"`
	DeliveryNote           string             `bson:"delivery_note" json:"delivery_note"`
	ShippingIdentification string             `bson:"shipping_identification" json:"shipping_identification"`
	Guest                  bool               `bson:"guest" json:"guest"`
	GuestID                *string            `bson:"guest_id,omitempty" json:"guest_id,omitempty"`
	External               bool               `bson:"external" json:"external"`
	ExternalPlatform       string             `bson:"external_platform" json:"external_platform"`
	ExternalID             string             `bson:"external_id" json:"external_id"`
	ShippingCarrier        string             `bson:"shipping_carrier" json:"shipping_carrier"`
	ShippingService        string             `bson:"shipping_service" json:"shipping_service"`
	ShippingTrackingNumber string             `bson:"shipping_tracking_number" json:"shipping_tracking_number"`
	ShippingTrackingURL    string             `bson:"shipping_tracking_url" json:"shipping_tracking_url"`
	ActualRate             ShippingRate       `bson:"ship_current" json:"ship_current"`
	GiftSubject            string             `bson:"gift_sub" json:"gift_sub"`
	GiftMessage            string             `bson:"gift_mess" json:"gift_mess"`
	CATax                  bool               `bson:"ca_tax" json:"ca_tax"`
	CATaxRate              float64            `bson:"ca_tax_rate" json:"ca_tax_rate"`
	PaymentMethodID        string             `bson:"pm_id" json:"pm_id"`
}

type DraftOrder struct {
	ID                    primitive.ObjectID           `bson:"_id,omitempty" json:"id"`
	PrintfulID            string                       `bson:"printful_id" json:"printful_id"`
	CustomerID            int                          `bson:"customer_id" json:"customer_id"`
	Status                string                       `bson:"status" json:"status"` // Created (default), Modified, Expired, Abandoned, Attempted, Failed, Submitted
	Email                 string                       `bson:"email" json:"email"`
	FirstName             string                       `bson:"fname" json:"fname"`
	LastName              string                       `bson:"lname" json:"lname"`
	DateCreated           time.Time                    `bson:"date_created" json:"date_created"`
	DateAbandoned         *time.Time                   `bson:"date_abandoned,omitempty" json:"date_abandoned,omitempty"`
	StripePaymentIntentID string                       `bson:"stripe_payment_intent_id" json:"stripe_payment_intent_id"`
	StripeMethodID        string                       `bson:"stripe_method_id" json:"stripe_method_id"`
	Subtotal              int                          `bson:"subtotal" json:"subtotal"`
	OrderLevelDiscount    int                          `bson:"order_level_discount" json:"order_level_discount"`
	PostDiscountTotal     int                          `bson:"post_disc_total" json:"post_disc_total"`
	Shipping              int                          `bson:"shipping" json:"shipping"`
	Tax                   int                          `bson:"tax" json:"tax"`
	PostTaxTotal          int                          `bson:"post_tax_total" json:"post_tax_total"`
	Tip                   int                          `bson:"tip" json:"tip"`
	PreGiftCardTotal      int                          `bson:"pgc_total" json:"pgc_total"`
	GiftCardSum           int                          `bson:"gc_sum" json:"gc_sum"`
	Total                 int                          `bson:"total" json:"total"`
	OrderDiscount         OrderDiscount                `bson:"non_stacking_discount_code" json:"non_stacking_discount_code"`
	ShippingContact       *Contact                     `bson:"shipping_contact" json:"shipping_contact"`
	Lines                 []OrderLine                  `bson:"lines" json:"lines"`
	GiftCards             []OrderGiftCard              `bson:"gift_cards" json:"gift_cards"`
	Tags                  []string                     `bson:"tags" json:"tags"`
	DeliveryNote          string                       `bson:"delivery_note" json:"delivery_note"`
	Guest                 bool                         `bson:"guest" json:"guest"`
	GuestID               *string                      `bson:"guest_id,omitempty" json:"guest_id,omitempty"`
	GuestStripeID         *string                      `bson:"guest_stripe,omitempty" json:"guest_stripe,omitempty"`
	ActualRate            ShippingRate                 `bson:"ship_actual" json:"ship_actual"`
	CurrentShipping       []ShippingRate               `bson:"ship_current" json:"ship_current"`
	AllShippingRates      map[string][]ShippingRate    `bson:"ship_all" json:"ship_all"`
	AllOrderEstimates     map[string]OrderEstimateCost `bson:"order_est_all" json:"order_est_all"`
	OrderEstimate         OrderEstimateCost            `bson:"order_est" json:"order_est"`
	GiftSubject           string                       `bson:"gift_sub" json:"gift_sub"`
	GiftMessage           string                       `bson:"gift_mess" json:"gift_mess"`
	CATax                 bool                         `bson:"ca_tax" json:"ca_tax"`
	CATaxRate             float64                      `bson:"ca_tax_rate" json:"ca_tax_rate"`
	NewPaymentMethodID    string                       `bson:"new_pm_id" json:"new_pm_id"`
	ExistingPaymentMethod PaymentMethodStripe          `bson:"ex_pm" json:"ex_pm"`
	AllPaymentMethods     []PaymentMethodStripe        `bson:"all_pm" json:"all_pm"`
	ListedContacts        []*Contact                   `bson:"all_contacts" json:"all_contacts"`
}

type OrderGiftCard struct {
	GiftCardID      int    `bson:"gc_id" json:"gc_id"`
	Code            string `bson:"gc_code" json:"gc_code"`
	AmountAvailable int    `bson:"available" json:"available"`
	Charged         int    `bson:"charged" json:"charged"`
	Message         string `bson:"message" json:"message"`
	UseFullAmount   bool   `bson:"full_amount" json:"full_amount"`
}

type OrderDiscount struct {
	DiscountCode     string
	ShortMessage     string
	IsPercentageOff  bool
	PercentageOff    float64
	HasMinSubtotal   bool
	MinSubtotal      int
	AppliesToAllAny  bool
	SingleCustomerID int
	HasUserList      bool
	CustomerList     []int
}

type OrderLine struct {
	ImageURL          string                 `bson:"image_url" json:"image_url"`
	ProductTitle      string                 `bson:"product_title" json:"product_title"`
	Handle            string                 `bson:"handle" json:"handle"`
	PrintfulID        []OriginalProductRedis `bson:"pid" json:"pid"`
	Variant1Key       string                 `bson:"variant_1_key" json:"variant_1_key"`
	Variant1Value     string                 `bson:"variant_1_value" json:"variant_1_value"`
	Variant2Key       string                 `bson:"variant_2_key" json:"variant_2_key"`
	Variant2Value     string                 `bson:"variant_2_value" json:"variant_2_value"`
	Variant3Key       string                 `bson:"variant_3_key" json:"variant_3_key"`
	Variant3Value     string                 `bson:"variant_3_value" json:"variant_3_value"`
	ProductID         string                 `bson:"product_id" json:"product_id"`
	IsGiftCard        int                    `bson:"gift_card" json:"gift_card"`
	VariantID         string                 `bson:"variant_id" json:"variant_id"`
	Quantity          int                    `bson:"quantity" json:"quantity"`
	UndiscountedPrice int                    `bson:"undiscounted_price" json:"undiscounted_price"`
	Price             int                    `bson:"price" json:"price"`
	LineLevelDiscount int                    `bson:"line_level_discount" json:"line_level_discount"`
	EndPrice          int                    `bson:"end_price" json:"end_price"`
	LineTotal         int                    `bson:"line_total" json:"line_total"`
}

// type OrderContact struct {
// 	FirstName      string `bson:"first_name" json:"first_name"`
// 	LastName       string `bson:"last_name" json:"last_name"`
// 	CompanyName    string `bson:"comp_name" json:"comp_name"`
// 	PhoneNumber    string `bson:"phone_number" json:"phone_number"`
// 	StreetAddress1 string `bson:"street_address_1" json:"street_address_1"`
// 	StreetAddress2 string `bson:"street_address_2" json:"street_address_2"`
// 	City           string `bson:"city" json:"city"`
// 	ProvinceState  string `bson:"province_state" json:"province_state"`
// 	ZipCode        string `bson:"zip_code" json:"zip_code"`
// 	Country        string `bson:"country" json:"country"`
// }

type ShippingRate struct {
	ID              string    `json:"id" bson:"id"`
	Name            string    `json:"name" bson:"name"`
	Rate            string    `json:"rate" bson:"rate"`
	CentsRate       int       `json:"cents" bson:"cents"`
	Currency        string    `json:"currency" bson:"currency"`
	MinDeliveryDays int       `json:"minDeliveryDays" bson:"min_delivery_days"`
	MaxDeliveryDays int       `json:"maxDeliveryDays" bson:"max_delivery_days"`
	MinDeliveryDate string    `json:"minDeliveryDate" bson:"min_delivery_date"`
	MaxDeliveryDate string    `json:"maxDeliveryDate" bson:"max_delivery_date"`
	Timestamp       time.Time `json:"timestamp" bson:"timestamp"`
}

type OrderEstimateCost struct {
	Currency       string    `json:"currency" bson:"currency"`
	Subtotal       float64   `json:"subtotal" bson:"subtotal"`
	Discount       float64   `json:"discount" bson:"discount"`
	Shipping       float64   `json:"shipping" bson:"shipping"`
	Digitization   float64   `json:"digitization" bson:"digitization"`
	AdditionalFee  float64   `json:"additional_fee" bson:"additional_fee"`
	FulfillmentFee float64   `json:"fulfillment_fee" bson:"fulfillment_fee"`
	Tax            float64   `json:"tax" bson:"tax"`
	Vat            float64   `json:"vat" bson:"vat"`
	Total          float64   `json:"total" bson:"total"`
	Timestamp      time.Time `json:"timestamp" bson:"timestamp"`
	AddressFormat  string    `json:"address_form" bson:"address_form"`
	ShipRate       string    `json:"ship_rate" bson:"ship_rate"`
}
