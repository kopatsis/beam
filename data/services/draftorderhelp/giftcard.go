package draftorderhelp

import (
	"beam/data/models"
	"errors"
)

func AddGiftCardToOrder(giftCard *models.GiftCard, draftOrder *models.DraftOrder) error {
	if draftOrder.Total <= 0 {
		return errors.New("nothing to pay for with this card")
	}

	draftOrder.GiftCards = append(draftOrder.GiftCards, models.OrderGiftCard{
		GiftCardID:      giftCard.ID,
		Code:            giftCard.IDCode,
		AmountAvailable: giftCard.LeftoverCents,
		Charged:         0,
		Message:         giftCard.ShortMessage,
	})

	return nil
}

func ApplyGiftCardToOrder(gcID int, cents int, draftOrder *models.DraftOrder) error {
	if draftOrder.Total <= 0 {
		return errors.New("nothing to pay for with this card")
	}

	if cents < 0 {
		return errors.New("must be positive cents")
	}

	var gc models.OrderGiftCard
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

	if cents > gc.AmountAvailable {
		cents = gc.AmountAvailable
	}

	delta := gc.Charged - cents
	if delta == 0 {
		return nil
	}

	if draftOrder.Total+delta < 0 {
		delta = -1 * draftOrder.Total
		cents = gc.Charged - delta
	}

	if draftOrder.Total+delta > 0 {
		if err := updateStripePaymentIntent(draftOrder.StripePaymentIntentID, draftOrder.Total+delta); err != nil {
			return err
		}
	}

	draftOrder.Total += delta
	draftOrder.GiftCardSum -= delta

	draftOrder.GiftCards[ind].Charged = cents

	return nil
}

func RemoveGiftCardFromOrder(gcID int, draftOrder *models.DraftOrder) error {

	var gc models.OrderGiftCard
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

	if gc.Charged != 0 {
		if err := updateStripePaymentIntent(draftOrder.StripePaymentIntentID, draftOrder.Total+gc.Charged); err != nil {
			return err
		}
		draftOrder.Total += gc.Charged
		draftOrder.GiftCardSum -= gc.Charged
	}

	draftOrder.GiftCards = append(draftOrder.GiftCards[:ind], draftOrder.GiftCards[ind+1:]...)

	return nil
}

func LowerGiftCardSum(draftOrder *models.DraftOrder, newAmount int) error {
	if newAmount >= draftOrder.GiftCardSum {
		return nil
	}

	if len(draftOrder.GiftCards) == 0 {
		return errors.New("no gift cards here")
	}

	i := len(draftOrder.GiftCards) - 1
	for i >= 0 && draftOrder.GiftCardSum > newAmount {
		if draftOrder.GiftCards[i].Charged <= 0 {
			continue
		}
		if draftOrder.GiftCardSum-newAmount > draftOrder.GiftCards[i].Charged {
			draftOrder.GiftCardSum -= draftOrder.GiftCards[i].Charged
			draftOrder.GiftCards[i].Charged = 0
		} else if draftOrder.GiftCardSum-newAmount == draftOrder.GiftCards[i].Charged {
			draftOrder.GiftCards[i].Charged = 0
			draftOrder.GiftCardSum = newAmount
		} else {
			draftOrder.GiftCards[i].Charged -= draftOrder.GiftCardSum - newAmount
			draftOrder.GiftCardSum = newAmount
		}
	}

	if draftOrder.GiftCardSum > newAmount {
		return errors.New("gift card amount zeroed out, but still above newAmount, somehow: SERIOUS")
	}

	return nil
}
