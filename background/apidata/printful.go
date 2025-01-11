package apidata

import "beam/data/models"

type PackageShippedPF struct {
	Type    string `json:"type"`
	Created int    `json:"created"`
	Retries int    `json:"retries"`
	Store   int    `json:"store"`
	Data    struct {
		Shipment struct {
			ID             int    `json:"id"`
			Carrier        string `json:"carrier"`
			Service        string `json:"service"`
			TrackingNumber int    `json:"tracking_number"`
			TrackingURL    string `json:"tracking_url"`
			Created        int    `json:"created"`
			ShipDate       string `json:"ship_date"`
			ShippedAt      int    `json:"shipped_at"`
			Reshipment     bool   `json:"reshipment"`
			Items          []struct {
				ItemID   int `json:"item_id"`
				Quantity int `json:"quantity"`
				Picked   int `json:"picked"`
				Printed  int `json:"printed"`
			} `json:"items"`
		} `json:"shipment"`
		Order struct {
			ID                  int    `json:"id"`
			ExternalID          string `json:"external_id"`
			Store               int    `json:"store"`
			Status              string `json:"status"`
			Shipping            string `json:"shipping"`
			ShippingServiceName string `json:"shipping_service_name"`
			Created             int    `json:"created"`
			Updated             int    `json:"updated"`
			Recipient           struct {
				Name        string `json:"name"`
				Company     string `json:"company"`
				Address1    string `json:"address1"`
				Address2    string `json:"address2"`
				City        string `json:"city"`
				StateCode   string `json:"state_code"`
				StateName   string `json:"state_name"`
				CountryCode string `json:"country_code"`
				CountryName string `json:"country_name"`
				Zip         string `json:"zip"`
				Phone       string `json:"phone"`
				Email       string `json:"email"`
				TaxNumber   string `json:"tax_number"`
			} `json:"recipient"`
			Items []struct {
				ID                        int    `json:"id"`
				ExternalID                string `json:"external_id"`
				VariantID                 int    `json:"variant_id"`
				SyncVariantID             int    `json:"sync_variant_id"`
				ExternalVariantID         string `json:"external_variant_id"`
				WarehouseProductVariantID int    `json:"warehouse_product_variant_id"`
				ProductTemplateID         int    `json:"product_template_id"`
				Quantity                  int    `json:"quantity"`
				Price                     string `json:"price"`
				RetailPrice               string `json:"retail_price"`
				Name                      string `json:"name"`
				Product                   struct {
					VariantID int    `json:"variant_id"`
					ProductID int    `json:"product_id"`
					Image     string `json:"image"`
					Name      string `json:"name"`
				} `json:"product"`
				Files []struct {
					Type    string `json:"type"`
					ID      int    `json:"id"`
					URL     string `json:"url"`
					Options []struct {
						ID    string `json:"id"`
						Value string `json:"value"`
					} `json:"options"`
					Hash         string `json:"hash"`
					Filename     string `json:"filename"`
					MimeType     string `json:"mime_type"`
					Size         int    `json:"size"`
					Width        int    `json:"width"`
					Height       int    `json:"height"`
					Dpi          int    `json:"dpi"`
					Status       string `json:"status"`
					Created      int    `json:"created"`
					ThumbnailURL string `json:"thumbnail_url"`
					PreviewURL   string `json:"preview_url"`
					Visible      bool   `json:"visible"`
					IsTemporary  bool   `json:"is_temporary"`
				} `json:"files"`
				Options []struct {
					ID    string `json:"id"`
					Value string `json:"value"`
				} `json:"options"`
				Sku          any  `json:"sku"`
				Discontinued bool `json:"discontinued"`
				OutOfStock   bool `json:"out_of_stock"`
			} `json:"items"`
			BrandingItems []struct {
				ID                        int    `json:"id"`
				ExternalID                string `json:"external_id"`
				VariantID                 int    `json:"variant_id"`
				SyncVariantID             int    `json:"sync_variant_id"`
				ExternalVariantID         string `json:"external_variant_id"`
				WarehouseProductVariantID int    `json:"warehouse_product_variant_id"`
				ProductTemplateID         int    `json:"product_template_id"`
				Quantity                  int    `json:"quantity"`
				Price                     string `json:"price"`
				RetailPrice               string `json:"retail_price"`
				Name                      string `json:"name"`
				Product                   struct {
					VariantID int    `json:"variant_id"`
					ProductID int    `json:"product_id"`
					Image     string `json:"image"`
					Name      string `json:"name"`
				} `json:"product"`
				Files []struct {
					Type    string `json:"type"`
					ID      int    `json:"id"`
					URL     string `json:"url"`
					Options []struct {
						ID    string `json:"id"`
						Value string `json:"value"`
					} `json:"options"`
					Hash         string `json:"hash"`
					Filename     string `json:"filename"`
					MimeType     string `json:"mime_type"`
					Size         int    `json:"size"`
					Width        int    `json:"width"`
					Height       int    `json:"height"`
					Dpi          int    `json:"dpi"`
					Status       string `json:"status"`
					Created      int    `json:"created"`
					ThumbnailURL string `json:"thumbnail_url"`
					PreviewURL   string `json:"preview_url"`
					Visible      bool   `json:"visible"`
					IsTemporary  bool   `json:"is_temporary"`
				} `json:"files"`
				Options []struct {
					ID    string `json:"id"`
					Value string `json:"value"`
				} `json:"options"`
				Sku          any  `json:"sku"`
				Discontinued bool `json:"discontinued"`
				OutOfStock   bool `json:"out_of_stock"`
			} `json:"branding_items"`
			IncompleteItems []struct {
				Name               string `json:"name"`
				Quantity           int    `json:"quantity"`
				SyncVariantID      int    `json:"sync_variant_id"`
				ExternalVariantID  string `json:"external_variant_id"`
				ExternalLineItemID string `json:"external_line_item_id"`
			} `json:"incomplete_items"`
			Costs struct {
				Currency          string `json:"currency"`
				Subtotal          string `json:"subtotal"`
				Discount          string `json:"discount"`
				Shipping          string `json:"shipping"`
				Digitization      string `json:"digitization"`
				AdditionalFee     string `json:"additional_fee"`
				FulfillmentFee    string `json:"fulfillment_fee"`
				RetailDeliveryFee string `json:"retail_delivery_fee"`
				Tax               string `json:"tax"`
				Vat               string `json:"vat"`
				Total             string `json:"total"`
			} `json:"costs"`
			RetailCosts struct {
				Currency string `json:"currency"`
				Subtotal string `json:"subtotal"`
				Discount string `json:"discount"`
				Shipping string `json:"shipping"`
				Tax      string `json:"tax"`
				Vat      string `json:"vat"`
				Total    string `json:"total"`
			} `json:"retail_costs"`
			PricingBreakdown []struct {
				CustomerPays   string `json:"customer_pays"`
				PrintfulPrice  string `json:"printful_price"`
				Profit         string `json:"profit"`
				CurrencySymbol string `json:"currency_symbol"`
			} `json:"pricing_breakdown"`
			Shipments []struct {
				ID             int    `json:"id"`
				Carrier        string `json:"carrier"`
				Service        string `json:"service"`
				TrackingNumber int    `json:"tracking_number"`
				TrackingURL    string `json:"tracking_url"`
				Created        int    `json:"created"`
				ShipDate       string `json:"ship_date"`
				ShippedAt      int    `json:"shipped_at"`
				Reshipment     bool   `json:"reshipment"`
				Items          []struct {
					ItemID   int `json:"item_id"`
					Quantity int `json:"quantity"`
					Picked   int `json:"picked"`
					Printed  int `json:"printed"`
				} `json:"items"`
			} `json:"shipments"`
			Gift struct {
				Subject string `json:"subject"`
				Message string `json:"message"`
			} `json:"gift"`
			PackingSlip struct {
				Email         string `json:"email"`
				Phone         string `json:"phone"`
				Message       string `json:"message"`
				LogoURL       string `json:"logo_url"`
				StoreName     string `json:"store_name"`
				CustomOrderID string `json:"custom_order_id"`
			} `json:"packing_slip"`
		} `json:"order"`
	} `json:"data"`
}

type FromCostEstimate struct {
	Code   int `json:"code"`
	Result struct {
		Costs       models.OrderEstimateCost `json:"costs"`
		RetailCosts struct {
			Currency string `json:"currency"`
			Subtotal int    `json:"subtotal"`
			Discount int    `json:"discount"`
			Shipping int    `json:"shipping"`
			Tax      int    `json:"tax"`
			Vat      int    `json:"vat"`
			Total    int    `json:"total"`
		} `json:"retail_costs"`
	} `json:"result"`
}

type ToCostEstimate struct {
	Shipping  string    `json:"shipping,omitempty"`
	Recipient Recipient `json:"recipient,omitempty"`
	Items     []Items   `json:"items,omitempty"`
	Currency  string    `json:"currency,omitempty"`
	Locale    string    `json:"locale,omitempty"`
}

type Recipient struct {
	Address1    string `json:"address1,omitempty"`
	City        string `json:"city,omitempty"`
	CountryCode string `json:"country_code,omitempty"`
	StateCode   string `json:"state_code,omitempty"`
	Zip         string `json:"zip,omitempty"`
}
type Product struct {
	VariantID int `json:"variant_id,omitempty"`
	ProductID int `json:"product_id,omitempty"`
}
type Items struct {
	ID                int     `json:"id,omitempty"`
	ExternalID        string  `json:"external_id,omitempty"`
	VariantID         int     `json:"variant_id,omitempty"`
	SyncVariantID     int64   `json:"sync_variant_id,omitempty"`
	ExternalVariantID string  `json:"external_variant_id,omitempty"`
	Quantity          int     `json:"quantity,omitempty"`
	Product           Product `json:"product,omitempty"`
}

type Order struct {
	ExternalID  string            `json:"external_id"`
	Shipping    string            `json:"shipping"`
	Recipient   OrderRecipient    `json:"recipient"`
	Items       []OrderItems      `json:"items"`
	RetailCosts *OrderRetailCosts `json:"retail_costs,omitempty"`
	Gift        *OrderGift        `json:"gift,omitempty"`
	PackingSlip *OrderPackingSlip `json:"packing_slip,omitempty"`
}
type OrderRecipient struct {
	Name        string `json:"name"`
	Company     string `json:"company"`
	Address1    string `json:"address1"`
	Address2    string `json:"address2,omitempty"`
	City        string `json:"city"`
	StateCode   string `json:"state_code"`
	StateName   string `json:"state_name"`
	CountryCode string `json:"country_code"`
	CountryName string `json:"country_name"`
	Zip         string `json:"zip"`
	Phone       string `json:"phone,omitempty"`
	Email       string `json:"email"`
	TaxNumber   string `json:"tax_number,omitempty"`
}
type OrderProduct struct {
	VariantID int    `json:"variant_id"`
	ProductID int    `json:"product_id"`
	Image     string `json:"image"`
	Name      string `json:"name"`
}
type OrderOptions struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}
type OrderPosition struct {
	AreaWidth        int  `json:"area_width"`
	AreaHeight       int  `json:"area_height"`
	Width            int  `json:"width"`
	Height           int  `json:"height"`
	Top              int  `json:"top"`
	Left             int  `json:"left"`
	LimitToPrintArea bool `json:"limit_to_print_area"`
}
type OrderFiles struct {
	Type     string         `json:"type"`
	URL      string         `json:"url"`
	Options  []OrderOptions `json:"options"`
	Filename string         `json:"filename"`
	Visible  bool           `json:"visible"`
	Position OrderPosition  `json:"position"`
}
type OrderItems struct {
	ID                        int            `json:"id"`
	ExternalID                string         `json:"external_id"`
	VariantID                 int            `json:"variant_id"`
	SyncVariantID             int            `json:"sync_variant_id"`
	ExternalVariantID         string         `json:"external_variant_id"`
	WarehouseProductVariantID int            `json:"warehouse_product_variant_id"`
	ProductTemplateID         int            `json:"product_template_id"`
	ExternalProductID         string         `json:"external_product_id"`
	Quantity                  int            `json:"quantity"`
	Price                     string         `json:"price"`
	RetailPrice               string         `json:"retail_price"`
	Name                      string         `json:"name"`
	Product                   OrderProduct   `json:"product"`
	Files                     []OrderFiles   `json:"files"`
	Options                   []OrderOptions `json:"options"`
	Sku                       any            `json:"sku"`
	Discontinued              bool           `json:"discontinued"`
	OutOfStock                bool           `json:"out_of_stock"`
}
type OrderRetailCosts struct {
	Currency string `json:"currency"`
	Subtotal string `json:"subtotal"`
	Discount string `json:"discount"`
	Shipping string `json:"shipping"`
	Tax      string `json:"tax"`
}
type OrderGift struct {
	Subject string `json:"subject"`
	Message string `json:"message"`
}
type OrderPackingSlip struct {
	Email         string `json:"email"`
	Phone         string `json:"phone"`
	Message       string `json:"message"`
	LogoURL       string `json:"logo_url"`
	StoreName     string `json:"store_name"`
	CustomOrderID string `json:"custom_order_id"`
}

type OrderResponse struct {
	Code   int `json:"code"`
	Result struct {
		ID                  int    `json:"id"`
		ExternalID          string `json:"external_id"`
		Store               int    `json:"store"`
		Status              string `json:"status"`
		Shipping            string `json:"shipping"`
		ShippingServiceName string `json:"shipping_service_name"`
		Created             int    `json:"created"`
		Updated             int    `json:"updated"`
		Recipient           struct {
			Name        string `json:"name"`
			Company     string `json:"company"`
			Address1    string `json:"address1"`
			Address2    string `json:"address2"`
			City        string `json:"city"`
			StateCode   string `json:"state_code"`
			StateName   string `json:"state_name"`
			CountryCode string `json:"country_code"`
			CountryName string `json:"country_name"`
			Zip         string `json:"zip"`
			Phone       string `json:"phone"`
			Email       string `json:"email"`
			TaxNumber   string `json:"tax_number"`
		} `json:"recipient"`
		Items []struct {
			ID                        int    `json:"id"`
			ExternalID                string `json:"external_id"`
			VariantID                 int    `json:"variant_id"`
			SyncVariantID             int    `json:"sync_variant_id"`
			ExternalVariantID         string `json:"external_variant_id"`
			WarehouseProductVariantID int    `json:"warehouse_product_variant_id"`
			ProductTemplateID         int    `json:"product_template_id"`
			Quantity                  int    `json:"quantity"`
			Price                     string `json:"price"`
			RetailPrice               string `json:"retail_price"`
			Name                      string `json:"name"`
			Product                   struct {
				VariantID int    `json:"variant_id"`
				ProductID int    `json:"product_id"`
				Image     string `json:"image"`
				Name      string `json:"name"`
			} `json:"product"`
			Files []struct {
				Type    string `json:"type"`
				ID      int    `json:"id"`
				URL     string `json:"url"`
				Options []struct {
					ID    string `json:"id"`
					Value string `json:"value"`
				} `json:"options"`
				Hash         string `json:"hash"`
				Filename     string `json:"filename"`
				MimeType     string `json:"mime_type"`
				Size         int    `json:"size"`
				Width        int    `json:"width"`
				Height       int    `json:"height"`
				Dpi          int    `json:"dpi"`
				Status       string `json:"status"`
				Created      int    `json:"created"`
				ThumbnailURL string `json:"thumbnail_url"`
				PreviewURL   string `json:"preview_url"`
				Visible      bool   `json:"visible"`
				IsTemporary  bool   `json:"is_temporary"`
			} `json:"files"`
			Options []struct {
				ID    string `json:"id"`
				Value string `json:"value"`
			} `json:"options"`
			Sku          any  `json:"sku"`
			Discontinued bool `json:"discontinued"`
			OutOfStock   bool `json:"out_of_stock"`
		} `json:"items"`
		BrandingItems []struct {
			ID                        int    `json:"id"`
			ExternalID                string `json:"external_id"`
			VariantID                 int    `json:"variant_id"`
			SyncVariantID             int    `json:"sync_variant_id"`
			ExternalVariantID         string `json:"external_variant_id"`
			WarehouseProductVariantID int    `json:"warehouse_product_variant_id"`
			ProductTemplateID         int    `json:"product_template_id"`
			Quantity                  int    `json:"quantity"`
			Price                     string `json:"price"`
			RetailPrice               string `json:"retail_price"`
			Name                      string `json:"name"`
			Product                   struct {
				VariantID int    `json:"variant_id"`
				ProductID int    `json:"product_id"`
				Image     string `json:"image"`
				Name      string `json:"name"`
			} `json:"product"`
			Files []struct {
				Type    string `json:"type"`
				ID      int    `json:"id"`
				URL     string `json:"url"`
				Options []struct {
					ID    string `json:"id"`
					Value string `json:"value"`
				} `json:"options"`
				Hash         string `json:"hash"`
				Filename     string `json:"filename"`
				MimeType     string `json:"mime_type"`
				Size         int    `json:"size"`
				Width        int    `json:"width"`
				Height       int    `json:"height"`
				Dpi          int    `json:"dpi"`
				Status       string `json:"status"`
				Created      int    `json:"created"`
				ThumbnailURL string `json:"thumbnail_url"`
				PreviewURL   string `json:"preview_url"`
				Visible      bool   `json:"visible"`
				IsTemporary  bool   `json:"is_temporary"`
			} `json:"files"`
			Options []struct {
				ID    string `json:"id"`
				Value string `json:"value"`
			} `json:"options"`
			Sku          any  `json:"sku"`
			Discontinued bool `json:"discontinued"`
			OutOfStock   bool `json:"out_of_stock"`
		} `json:"branding_items"`
		IncompleteItems []struct {
			Name               string `json:"name"`
			Quantity           int    `json:"quantity"`
			SyncVariantID      int    `json:"sync_variant_id"`
			ExternalVariantID  string `json:"external_variant_id"`
			ExternalLineItemID string `json:"external_line_item_id"`
		} `json:"incomplete_items"`
		Costs struct {
			Currency          string `json:"currency"`
			Subtotal          string `json:"subtotal"`
			Discount          string `json:"discount"`
			Shipping          string `json:"shipping"`
			Digitization      string `json:"digitization"`
			AdditionalFee     string `json:"additional_fee"`
			FulfillmentFee    string `json:"fulfillment_fee"`
			RetailDeliveryFee string `json:"retail_delivery_fee"`
			Tax               string `json:"tax"`
			Vat               string `json:"vat"`
			Total             string `json:"total"`
		} `json:"costs"`
		RetailCosts struct {
			Currency string `json:"currency"`
			Subtotal string `json:"subtotal"`
			Discount string `json:"discount"`
			Shipping string `json:"shipping"`
			Tax      string `json:"tax"`
			Vat      string `json:"vat"`
			Total    string `json:"total"`
		} `json:"retail_costs"`
		PricingBreakdown []struct {
			CustomerPays   string `json:"customer_pays"`
			PrintfulPrice  string `json:"printful_price"`
			Profit         string `json:"profit"`
			CurrencySymbol string `json:"currency_symbol"`
		} `json:"pricing_breakdown"`
		Shipments []struct {
			ID             int    `json:"id"`
			Carrier        string `json:"carrier"`
			Service        string `json:"service"`
			TrackingNumber int    `json:"tracking_number"`
			TrackingURL    string `json:"tracking_url"`
			Created        int    `json:"created"`
			ShipDate       string `json:"ship_date"`
			ShippedAt      int    `json:"shipped_at"`
			Reshipment     bool   `json:"reshipment"`
			Items          []struct {
				ItemID   int `json:"item_id"`
				Quantity int `json:"quantity"`
				Picked   int `json:"picked"`
				Printed  int `json:"printed"`
			} `json:"items"`
		} `json:"shipments"`
		Gift struct {
			Subject string `json:"subject"`
			Message string `json:"message"`
		} `json:"gift"`
		PackingSlip struct {
			Email         string `json:"email"`
			Phone         string `json:"phone"`
			Message       string `json:"message"`
			LogoURL       string `json:"logo_url"`
			StoreName     string `json:"store_name"`
			CustomOrderID string `json:"custom_order_id"`
		} `json:"packing_slip"`
	} `json:"result"`
}
