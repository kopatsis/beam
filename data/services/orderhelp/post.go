package orderhelp

import (
	"beam/background/apidata"
	"beam/background/emails"
	"beam/config"
	"beam/data/models"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

func CopyString(s *string) *string {
	if s == nil {
		return nil
	}
	copy := *s
	return &copy
}

func CopyContact(c *models.Contact) *models.Contact {
	if c == nil {
		return nil
	}
	copy := *c
	return &copy
}

func CreateOrderFromDraft(draft *models.DraftOrder) *models.Order {
	return &models.Order{
		PrintfulID:         draft.PrintfulID,
		CustomerID:         draft.CustomerID,
		DraftOrderID:       draft.ID.Hex(),
		Status:             "Draft",
		Email:              draft.Email,
		Name:               draft.Name,
		DateCreated:        time.Now(),
		Subtotal:           draft.Subtotal,
		OrderLevelDiscount: draft.OrderLevelDiscount,
		PostDiscountTotal:  draft.PostDiscountTotal,
		Shipping:           draft.Shipping,
		Tax:                draft.Tax,
		PostTaxTotal:       draft.PostTaxTotal,
		Tip:                draft.Tip,
		PreGiftCardTotal:   draft.PreGiftCardTotal,
		GiftCardSum:        draft.GiftCardSum,
		Total:              draft.Total,
		OrderDiscount:      draft.OrderDiscount,
		ShippingContact:    CopyContact(draft.ShippingContact),
		Lines:              draft.Lines,
		GiftCards:          draft.GiftCards,
		Tags:               draft.Tags,
		Guest:              draft.Guest,
		GuestID:            CopyString(draft.GuestID),
		GuestStripeID:      CopyString(draft.GuestStripeID),
		ActualRate:         draft.ActualRate,
		GiftSubject:        draft.GiftSubject,
		GiftMessage:        draft.GiftMessage,
		CATax:              draft.CATax,
		CATaxRate:          draft.CATaxRate,
	}
}

func CreatePrintfulOrder(order *models.Order, mutex *config.AllMutexes) (*apidata.Order, error) {
	ret := &apidata.Order{
		ExternalID: order.ID.Hex(),
		Shipping:   order.ActualRate.ID,
		Recipient: apidata.OrderRecipient{
			Name:        order.ShippingContact.FirstName,
			Address1:    order.ShippingContact.StreetAddress1,
			City:        order.ShippingContact.City,
			Zip:         order.ShippingContact.ZipCode,
			CountryName: order.ShippingContact.Country,
			CountryCode: order.ShippingContact.CountryCode,
			Email:       order.Email,
		},
		Items: []apidata.OrderItems{},
	}

	if order.ShippingContact.StateCode != "" && order.ShippingContact.ProvinceState != "" {
		ret.Recipient.StateName = order.ShippingContact.ProvinceState
		ret.Recipient.StateCode = order.ShippingContact.StateCode
	}

	if order.ShippingContact.LastName != nil {
		ret.Recipient.Name += " " + *CopyString(order.ShippingContact.LastName)
	}

	if order.ShippingContact.Company != nil {
		ret.Recipient.Company = *CopyString(order.ShippingContact.Company)
	}

	if order.ShippingContact.StreetAddress2 != nil {
		ret.Recipient.Address2 = *CopyString(order.ShippingContact.StreetAddress2)
	}

	if order.ShippingContact.PhoneNumber != nil {
		ret.Recipient.Phone = *CopyString(order.ShippingContact.PhoneNumber)
	}

	itemMap := map[string]models.OriginalProductRedis{}

	for _, line := range order.Lines {
		for _, o := range line.PrintfulID {
			if l, ok := itemMap[o.VariantID]; ok {
				l.Quantity += o.Quantity
				itemMap[o.VariantID] = l
			} else {
				itemMap[o.VariantID] = o
			}
		}
	}

	i := 0
	for vid, line := range itemMap {
		i++

		varID, err := strconv.Atoi(vid)
		if err != nil {
			return ret, err
		}

		syncVarID, err := strconv.Atoi(line.OriginalVariantID)
		if err != nil {
			return ret, err
		}

		ret.Items = append(ret.Items, apidata.OrderItems{
			ID:                i,
			VariantID:         varID,
			SyncVariantID:     syncVarID,
			ExternalVariantID: line.ExternalVariantID,
			ExternalProductID: line.ExternalProductID,
			Quantity:          line.Quantity,
			Sku:               line.SKU,
			Price:             fmt.Sprintf("%d.%02d", line.RetailPrice/100, line.RetailPrice%100),
			RetailPrice:       fmt.Sprintf("%d.%02d", line.RetailPrice/100, line.RetailPrice%100),
			Name:              line.FullVariantName,
		})
	}

	if order.GiftSubject != "" || order.GiftMessage != "" {
		ret.Gift = &apidata.OrderGift{
			Subject: order.GiftSubject,
			Message: order.GiftMessage,
		}
	}

	return ret, nil
}

func PostOrderToPrintful(order *models.Order, name string, mutexes *config.AllMutexes, tools *config.Tools) (*apidata.OrderResponse, error) {
	mutexes.Api.Mu.RLock()
	apiKey := mutexes.Api.KeyMap[name]
	mutexes.Api.Mu.RUnlock()

	bodyStruct, err := CreatePrintfulOrder(order, mutexes)
	if err != nil {
		return nil, err
	}

	bodyJSON, err := json.Marshal(bodyStruct)
	if err != nil {
		return nil, err
	}

	base := os.Getenv("PF_URL")
	req, err := http.NewRequest("POST", base+"/orders?confirm=true", bytes.NewBuffer(bodyJSON))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := tools.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error with response: http status: %d", resp.StatusCode)
	}

	var apiResponse apidata.OrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, err
	}

	return &apiResponse, nil
}

func ConfirmOrderPostResponse(resp *apidata.OrderResponse, order *models.Order) error {

	if resp.Code != 200 {
		return fmt.Errorf("response code not okay, should be 200, is : %d", resp.Code)
	}

	order.PrintfulID = strconv.Itoa(resp.Result.ID)

	return nil
}

func OrderEmailWithProfit(resp *apidata.OrderResponse, order *models.Order, tools *config.Tools, name string) error {

	cost, err := convertRateToCents(resp.Result.Costs.Total)
	if err != nil {
		return err
	}

	price := order.PreGiftCardTotal

	if order.CATax {
		price -= order.Tax
	}

	emails.OrderSuccessWithProfit(name, order.ID.Hex(), order.PrintfulID, tools, cost, price)

	return nil
}

func convertRateToCents(rate string) (int, error) {
	var rateInt int
	_, err := fmt.Sscanf(rate, "%f", &rateInt)
	if err != nil {
		return 0, fmt.Errorf("invalid rate format: %v", err)
	}
	return rateInt, nil
}
