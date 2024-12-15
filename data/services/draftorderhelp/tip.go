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

	if err := EnsureGiftCardSum(draftOrder, 0, newPGTotal, false); err != nil {
		return err
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

	if err := EnsureGiftCardSum(draftOrder, 0, newPGTotal, false); err != nil {
		return err
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
