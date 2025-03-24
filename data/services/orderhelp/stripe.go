package orderhelp

import (
	"beam/data/models"
	"beam/data/services/draftorderhelp"
)

func OrderPaymentMethodUpdate(order *models.Order, stripeID string) error {
	pms, err := draftorderhelp.GetAllPaymentMethods(stripeID)
	if err != nil {
		return err
	}

	order.PaymentMethodsForFailed = pms

	return nil
}
