package apidata

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

type CodeEstimate struct {
	Code   int `json:"code"`
	Result struct {
		Costs struct {
			Currency       string `json:"currency"`
			Subtotal       int    `json:"subtotal"`
			Discount       int    `json:"discount"`
			Shipping       int    `json:"shipping"`
			Digitization   int    `json:"digitization"`
			AdditionalFee  int    `json:"additional_fee"`
			FulfillmentFee int    `json:"fulfillment_fee"`
			Tax            int    `json:"tax"`
			Vat            int    `json:"vat"`
			Total          int    `json:"total"`
		} `json:"costs"`
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
