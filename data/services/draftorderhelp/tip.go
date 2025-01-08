package draftorderhelp

import (
	"beam/data/models"
	"errors"
)

func AddTipToOrder(draftOrder *models.DraftOrder, cents int) error {

	if cents < 0 {
		return errors.New("negative tip not allowed")
	}

	newPGTotal := draftOrder.PostTaxTotal + cents
	draftOrder.Tip = cents

	return EnsureGiftCardSum(draftOrder, 0, newPGTotal, false)
}

func DeleteTipFromOrder(draftOrder *models.DraftOrder) error {

	newPGTotal := draftOrder.PostTaxTotal
	draftOrder.Tip = 0

	return EnsureGiftCardSum(draftOrder, 0, newPGTotal, false)
}
