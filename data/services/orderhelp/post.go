package orderhelp

import (
	"beam/data/models"
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
		FirstName:          draft.FirstName,
		LastName:           draft.LastName,
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
		Tags:               draft.Tags,
		DeliveryNote:       draft.DeliveryNote,
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
