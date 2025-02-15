package draftorderhelp

import (
	"beam/config"
	"beam/data/models"
	"errors"
)

func MoreGiftCardsAllowed(draftOrder *models.DraftOrder) bool {
	if draftOrder == nil {
		return false
	}

	return draftOrder.GiftCards[0] == nil || draftOrder.GiftCards[1] == nil || draftOrder.GiftCards[2] == nil
}

func AddGiftCardToOrder(giftCard *models.GiftCard, draftOrder *models.DraftOrder) error {
	if draftOrder.Total <= 0 {
		return errors.New("nothing to pay for with this card")
	}

	var i int
	if draftOrder.GiftCards[0] == nil {
		i = 0
	} else if draftOrder.GiftCards[1] == nil {
		i = 1
	} else if draftOrder.GiftCards[2] == nil {
		i = 2
	} else {
		return errors.New("all gift card slots filled")
	}

	draftOrder.GiftCards[i] = &models.OrderGiftCard{
		GiftCardID:      giftCard.ID,
		Code:            giftCard.IDCode,
		AmountAvailable: giftCard.LeftoverCents,
		Charged:         0,
		Message:         giftCard.ShortMessage,
		Pin:             giftCard.Pin,
	}

	return nil
}

func ApplyGiftCardToOrder(gcID, cents int, fullAmount bool, draftOrder *models.DraftOrder) error {
	if draftOrder.Total <= 0 {
		return errors.New("nothing to pay for with this card")
	}

	if cents < 0 {
		return errors.New("must be positive cents")
	}

	var gc *models.OrderGiftCard
	ind := -1
	for i, g := range draftOrder.GiftCards {
		if g.GiftCardID == gcID {
			gc = g
			ind = i
		}
	}

	if ind == -1 {
		return errors.New("no matching gift card with that id")
	}

	if cents > gc.AmountAvailable || fullAmount {
		cents = gc.AmountAvailable
	}

	delta := gc.Charged - cents
	if delta == 0 {
		return nil
	}

	draftOrder.GiftCards[ind].Charged = cents
	draftOrder.GiftCards[ind].UseFullAmount = fullAmount

	return EnsureGiftCardSum(draftOrder, draftOrder.GiftCardSum-delta, 0, true)
}

func RemoveGiftCardFromOrder(gcID int, draftOrder *models.DraftOrder) error {

	var gc *models.OrderGiftCard
	ind := -1
	for i, g := range draftOrder.GiftCards {
		if g.GiftCardID == gcID {
			gc = g
			ind = i
		}
	}

	if ind == -1 {
		return errors.New("no matching gift card with that id")
	}

	draftOrder.GiftCards[ind] = nil

	if ind == 0 {
		draftOrder.GiftCards[0] = draftOrder.GiftCards[1]
		draftOrder.GiftCards[1] = draftOrder.GiftCards[2]
		draftOrder.GiftCards[2] = nil
	} else if ind == 1 {
		draftOrder.GiftCards[1] = draftOrder.GiftCards[2]
		draftOrder.GiftCards[2] = nil
	}

	if gc.Charged != 0 {
		return EnsureGiftCardSum(draftOrder, draftOrder.GiftCardSum-gc.Charged, 0, true)
	}

	return nil
}

func EnsureGiftCardSum(draftOrder *models.DraftOrder, newGiftCardSum, newPreGiftCardTotal int, fromGiftCardChange bool) error {

	if newGiftCardSum < 0 && fromGiftCardChange {
		return errors.New("gift card sum must be positive or 0")
	} else if newPreGiftCardTotal <= 0 && !fromGiftCardChange {
		return errors.New("gift card sum must be positive")
	}

	if len(draftOrder.GiftCards) == 0 {
		if fromGiftCardChange && newGiftCardSum > 0 {
			return errors.New("no gift cards to work with")
		}
		draftOrder.PreGiftCardTotal = newPreGiftCardTotal
		draftOrder.GiftCardSum = 0
		draftOrder.Total = newPreGiftCardTotal
		return nil
	}

	oldTotal := draftOrder.Total

	newTotal, usedGiftCardSum, usedPreGiftCardTotal := 0, 0, 0
	if fromGiftCardChange {
		usedGiftCardSum = newGiftCardSum
		usedPreGiftCardTotal = draftOrder.PreGiftCardTotal
	} else {
		usedGiftCardSum = draftOrder.GiftCardSum
		usedPreGiftCardTotal = newPreGiftCardTotal
	}
	newTotal = usedPreGiftCardTotal - usedGiftCardSum

	minPreGCAllowed := config.MIN_ORDER_PRICE - draftOrder.GiftCardBuyTotal
	if minPreGCAllowed < 0 {
		minPreGCAllowed = 0
	}

	if newTotal < 0 {

		for i := len(draftOrder.GiftCards) - 1; i >= 0; i++ {
			if newTotal >= 0 {
				break
			}
			gc := draftOrder.GiftCards[i]
			if !gc.UseFullAmount {
				delta := -1 * newTotal
				if gc.Charged < delta {
					delta = gc.Charged
				}
				newTotal += delta
				draftOrder.GiftCards[i].Charged -= delta
			}
		}

		if newTotal < 0 {
			for i := len(draftOrder.GiftCards) - 1; i >= 0; i++ {
				if newTotal >= 0 {
					break
				}
				gc := draftOrder.GiftCards[i]
				delta := -1 * newTotal
				if gc.Charged < delta {
					delta = gc.Charged
				}
				newTotal += delta
				draftOrder.GiftCards[i].Charged -= delta
			}
		}

		if newTotal < 0 {
			return errors.New("unable to apply gift cards correctly under current system")
		}

	} else if newTotal < minPreGCAllowed || checkIfUnappliedMaxedGC(draftOrder) {
		for i, gc := range draftOrder.GiftCards {
			if newTotal == 0 {
				break
			}
			if gc.UseFullAmount && gc.Charged < gc.AmountAvailable {
				delta := newTotal
				if gc.AmountAvailable-gc.Charged < delta {
					delta = gc.AmountAvailable - gc.Charged
				}
				newTotal -= delta
				draftOrder.GiftCards[i].Charged += delta
			}
		}

		if newTotal < 0 {
			return errors.New("unable to apply gift cards correctly under current system")
		} else if newTotal < minPreGCAllowed && newTotal > 0 {
			newTotal, usedGiftCardSum, usedPreGiftCardTotal = minPriceFix(draftOrder, newTotal, usedGiftCardSum, usedPreGiftCardTotal)
			if newTotal < 0 || newTotal > 0 && newTotal < minPreGCAllowed {
				return errors.New("unable to apply gift cards correctly under current system")
			}
		}
	}

	draftOrder.PreGiftCardTotal = usedPreGiftCardTotal
	draftOrder.GiftCardSum = usedGiftCardSum
	draftOrder.PostGiftCardTotal = newTotal
	draftOrder.Total = newTotal + draftOrder.GiftCardBuyTotal

	if draftOrder.Total != oldTotal {
		return updateStripePaymentIntent(draftOrder.StripePaymentIntentID, draftOrder.Total)
	}
	return nil
}

func checkIfUnappliedMaxedGC(draftOrder *models.DraftOrder) bool {
	for _, gc := range draftOrder.GiftCards {
		if gc.UseFullAmount && gc.Charged < gc.AmountAvailable {
			return true
		}
	}
	return false
}

func minPriceFix(draftOrder *models.DraftOrder, newTotal, usedGiftCardSum, usedPreGiftCardTotal int) (int, int, int) {
	for i := len(draftOrder.GiftCards) - 1; i >= 0; i++ {
		if newTotal >= config.MIN_ORDER_PRICE {
			break
		}
		gc := draftOrder.GiftCards[i]
		if !gc.UseFullAmount {
			delta := config.MIN_ORDER_PRICE - newTotal
			if gc.Charged < delta {
				delta = gc.Charged
			}
			newTotal += delta
			draftOrder.GiftCards[i].Charged -= delta
		}
	}

	if newTotal < config.MIN_ORDER_PRICE {
		for i := len(draftOrder.GiftCards) - 1; i >= 0; i++ {
			if newTotal >= config.MIN_ORDER_PRICE {
				break
			}
			gc := draftOrder.GiftCards[i]
			delta := config.MIN_ORDER_PRICE - newTotal
			if gc.Charged < delta {
				delta = gc.Charged
			}
			newTotal += delta
			draftOrder.GiftCards[i].Charged -= delta
		}
	}

	return newTotal, usedGiftCardSum, usedPreGiftCardTotal
}

func SetTotalsAndEnsure(draftOrder *models.DraftOrder) error {
	draftOrder.PostDiscountTotal = draftOrder.Subtotal - draftOrder.OrderLevelDiscount
	draftOrder.PostTaxTotal = draftOrder.PostDiscountTotal + draftOrder.Shipping + draftOrder.Tax
	newPreGiftCardTotal := draftOrder.PostTaxTotal + draftOrder.Tip

	return EnsureGiftCardSum(draftOrder, 0, newPreGiftCardTotal, false)
}
