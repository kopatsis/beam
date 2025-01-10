package services

import (
	"beam/data/models"
	"beam/data/repositories"
	"beam/data/services/draftorderhelp"
	"beam/data/services/orderhelp"
	"errors"
)

type OrderService interface {
	SubmitOrder(draftID, guestID, newPaymentMethod string, customerID int, saveMethod bool, useExisting bool, ds *draftOrderService, cs *customerService, dts *discountService) error
}

type orderService struct {
	orderRepo repositories.OrderRepository
}

func NewOrderService(orderRepo repositories.OrderRepository) OrderService {
	return &orderService{orderRepo: orderRepo}
}

func (s *orderService) SubmitOrder(draftID, guestID, newPaymentMethod string, customerID int, saveMethod bool, useExisting bool, ds *draftOrderService, cs *customerService, dts *discountService) error {

	draft, err := ds.GetDraftPtl(draftID, guestID, customerID)
	if err != nil {
		return err
	}

	var cust *models.Customer
	if customerID > 0 && !draft.Guest {
		cust, err = cs.GetCustomerByID(customerID)
		if err != nil {
			return err
		}
	}

	if useExisting {
		if draft.ExistingPaymentMethod.ID == "" {
			return errors.New("requires a chosen payment method if using existing payment method")
		} else if !(customerID > 0 && !draft.Guest) {
			return errors.New("requires non guest order if using existing payment method")
		}
	}

	order := orderhelp.CreateOrderFromDraft(draft)

	if err := s.orderRepo.CreateOrder(order); err != nil {
		return err
	}

	if useExisting && order.Total > 0 {
		intent, err := draftorderhelp.CreateAndChargePaymentIntent(draft.ExistingPaymentMethod.ID, cust.StripeID, order.Total)
		if err != nil {
			return err
		}
		draft.StripePaymentIntentID = intent.ID
		order.StripePaymentIntentID = intent.ID
	} else if order.Guest && order.Total > 0 {
		intent, err := draftorderhelp.ChargePaymentIntent(order.StripePaymentIntentID, newPaymentMethod, false, *order.GuestStripeID)
		if err != nil {
			return err
		}
		order.StripePaymentIntentID = intent.ID
	} else if order.Total > 0 {
		intent, err := draftorderhelp.ChargePaymentIntent(order.StripePaymentIntentID, newPaymentMethod, saveMethod, cust.StripeID)
		if err != nil {
			return err
		}
		order.StripePaymentIntentID = intent.ID
	}

	return nil
}
