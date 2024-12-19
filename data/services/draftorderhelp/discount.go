package draftorderhelp

import (
	"beam/data/models"
	"math"
)

func ApplyDiscountToOrder(disc *models.Discount, ul []*models.DiscountUser, draftOrder *models.DraftOrder) error {
	oldDiscPct := draftOrder.OrderDiscount.PercentageOff

	draftOrder.OrderDiscount = models.OrderDiscount{
		DiscountCode:     disc.DiscountCode,
		ShortMessage:     disc.ShortMessage,
		IsPercentageOff:  disc.IsPercentageOff,
		PercentageOff:    disc.PercentageOff,
		HasMinSubtotal:   disc.HasMinSubtotal,
		AppliesToAllAny:  disc.AppliesToAllAny,
		SingleCustomerID: disc.SingleCustomerID,
		HasUserList:      disc.HasUserList,
	}

	if disc.HasUserList {
		u := []int{}
		for _, l := range ul {
			u = append(u, l.CustomerID)
		}
		draftOrder.OrderDiscount.CustomerList = u
	}

	if disc.IsPercentageOff {
		return applyPercentOffToDraft(draftOrder, disc.PercentageOff, oldDiscPct)
	}

	return nil
}

func RemoveDiscountFromOrder(draftOrder *models.DraftOrder) error {
	oldDiscPct := draftOrder.OrderDiscount.PercentageOff

	draftOrder.OrderDiscount = models.OrderDiscount{}

	return applyPercentOffToDraft(draftOrder, 0, oldDiscPct)
}

func applyPercentOffToDraft(draftOrder *models.DraftOrder, percentageOff, oldDiscPct float64) error {
	discOff := int(math.Round(percentageOff * float64(draftOrder.Subtotal)))
	newPostDiscountTotal := draftOrder.Subtotal - discOff

	newTax := int(math.Round((1 - (percentageOff - oldDiscPct)) * float64(draftOrder.Tax)))

	newPostTaxTotal := newPostDiscountTotal + newTax + draftOrder.Shipping
	newPreGiftCardTotal := draftOrder.PostTaxTotal + draftOrder.Tip

	err := EnsureGiftCardSum(draftOrder, 0, newPreGiftCardTotal, false)
	if err != nil {
		return err
	}

	draftOrder.OrderLevelDiscount = newPostDiscountTotal
	draftOrder.PostDiscountTotal = newPostDiscountTotal
	draftOrder.PostTaxTotal = newPostTaxTotal
	draftOrder.PreGiftCardTotal = newPreGiftCardTotal
	draftOrder.Tax = newTax

	return nil
}
