package services

import (
	"beam/config"
	"beam/data/models"
	"beam/data/repositories"
	"beam/data/services/draftorderhelp"
	"beam/data/services/orderhelp"
	"errors"
)

type OrderService interface {
	SubmitOrder(draftID, guestID, newPaymentMethod, store string, customerID int, saveMethod bool, useExisting bool, ds *draftOrderService, cs *customerService, dts *discountService, mutexes *config.AllMutexes, tools *config.Tools) error
	UseDiscountsAndGiftCards(order *models.Order, guestID string, customerID int, ds *discountService) (error, error, bool)
}

type orderService struct {
	orderRepo repositories.OrderRepository
}

func NewOrderService(orderRepo repositories.OrderRepository) OrderService {
	return &orderService{orderRepo: orderRepo}
}

func (s *orderService) SubmitOrder(draftID, guestID, newPaymentMethod, store string, customerID int, saveMethod bool, useExisting bool, ds *draftOrderService, cs *customerService, dts *discountService, mutexes *config.AllMutexes, tools *config.Tools) error {

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

	resp, err := orderhelp.PostOrderToPrintful(order, store, mutexes, tools)
	if err != nil {
		return err
	}

	if err := orderhelp.ConfirmOrderPostResponse(resp); err != nil {
		return err
	}

	gcErr, discErr, worked := s.UseDiscountsAndGiftCards(order, guestID, customerID, dts)
	if !worked {
		if gcErr != nil {
			return gcErr
		} else if discErr != nil {
			return discErr
		}
	}

	return nil
}

// Giftcard error, discount error, both worked
func (s *orderService) UseDiscountsAndGiftCards(order *models.Order, guestID string, customerID int, ds *discountService) (error, error, bool) {

	gcErr, discErr := error(nil), error(nil)

	if len(order.GiftCards) != 0 {

		gcsAndAmounts := map[string]int{}
		for _, gc := range order.GiftCards {
			gcsAndAmounts[gc.Code] = gc.Charged
		}

		gcErr = ds.UseMultipleGiftCards(gcsAndAmounts)

		if gcErr != nil {
			return gcErr, nil, false
		}

	}

	if order.OrderDiscount.DiscountCode != "" {

		discErr = ds.CheckDiscountCode(order.OrderDiscount.DiscountCode, order.Subtotal, customerID, order.Guest)

		if discErr != nil {
			return nil, discErr, false
		}
	}

	return nil, nil, true
}
