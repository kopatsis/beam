package draftorderhelp

import (
	"beam/data/models"
	"errors"
)

func AddTipToOrder(draftOrder *models.DraftOrder, cents int) error {

	if cents < 0 {
		return errors.New("negative tip not allowed")
	}

	oldTotal := draftOrder.Total

	newPGTotal := draftOrder.PostTaxTotal + cents

	if newPGTotal < draftOrder.GiftCardSum {
		LowerGiftCardSum(draftOrder, newPGTotal)
	}

	newTotal := newPGTotal - draftOrder.GiftCardSum

	if newTotal != oldTotal {
		if err := updateStripePaymentIntent(draftOrder.StripePaymentIntentID, newTotal); err != nil {
			return err
		}
	}

	draftOrder.Tip = cents
	draftOrder.PreGiftCardTotal = newPGTotal
	draftOrder.Total = newTotal

	return nil
}

func DeleteTipFromOrder(draftOrder *models.DraftOrder) error {
	oldTotal := draftOrder.Total

	newPGTotal := draftOrder.PostTaxTotal

	if newPGTotal < draftOrder.GiftCardSum {
		LowerGiftCardSum(draftOrder, newPGTotal)
	}

	newTotal := newPGTotal - draftOrder.GiftCardSum

	if newTotal != oldTotal {
		if err := updateStripePaymentIntent(draftOrder.StripePaymentIntentID, newTotal); err != nil {
			return err
		}
	}

	draftOrder.Tip = 0
	draftOrder.PreGiftCardTotal = newPGTotal
	draftOrder.Total = newTotal

	return nil
}
