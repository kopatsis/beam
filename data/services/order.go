package services

import (
	"beam/background/emails"
	"beam/config"
	"beam/data/models"
	"beam/data/repositories"
	"beam/data/services/draftorderhelp"
	"beam/data/services/orderhelp"
	"errors"
	"log"
	"net/url"
	"strconv"
	"time"
)

type OrderService interface {
	SubmitOrder(draftID, guestID, newPaymentMethod, store string, customerID int, saveMethod bool, useExisting bool, ds *draftOrderService, cs *customerService, dts *discountService, ls *listService, ps *productService, mutexes *config.AllMutexes, tools *config.Tools) (error, error)
	UseDiscountsAndGiftCards(order *models.Order, guestID string, customerID int, ds *discountService) (error, error, bool)
	MarkOrderAndDraftAsSuccess(order *models.Order, draft *models.DraftOrder, ds *draftOrderService) error
	RenderOrder(orderID, guestID string, customerID int) (*models.Order, bool, error)
	GetOrdersList(customerID int, fromURL url.Values) (models.OrderRender, error)
}

type orderService struct {
	orderRepo repositories.OrderRepository
}

func NewOrderService(orderRepo repositories.OrderRepository) OrderService {
	return &orderService{orderRepo: orderRepo}
}

// Charging error, internal error
func (s *orderService) SubmitOrder(draftID, guestID, newPaymentMethod, store string, customerID int, saveMethod bool, useExisting bool, ds *draftOrderService, cs *customerService, dts *discountService, ls *listService, ps *productService, mutexes *config.AllMutexes, tools *config.Tools) (error, error) {

	draft, err := ds.GetDraftPtl(draftID, guestID, customerID)
	if err != nil {
		return nil, err
	}

	var cust *models.Customer
	if customerID > 0 && !draft.Guest {
		cust, err = cs.GetCustomerByID(customerID)
		if err != nil {
			return nil, err
		}
	}

	if useExisting {
		if draft.ExistingPaymentMethod.ID == "" {
			return errors.New("requires a chosen payment method if using existing payment method"), nil
		} else if !(customerID > 0 && !draft.Guest) {
			return errors.New("requires non guest order if using existing payment method"), nil
		}
	}

	order := orderhelp.CreateOrderFromDraft(draft)

	if err := s.orderRepo.CreateOrder(order); err != nil {
		return nil, err
	}

	draft.Status = "Submitted"
	draft.DateConverted = time.Now()

	if err := ds.draftOrderRepo.Update(draft); err != nil {
		go emails.AlertRecoverableOrderSubmitError(store, draftID, order.ID.Hex(), "Error when updating draft for order on mongodb", tools, order, draft, nil, false, err)
	}

	if useExisting && order.Total > 0 {
		intent, err := draftorderhelp.CreateAndChargePaymentIntent(draft.ExistingPaymentMethod.ID, cust.StripeID, order.Total)
		if err != nil {
			return err, nil
		}
		draft.StripePaymentIntentID = intent.ID
		order.StripePaymentIntentID = intent.ID
	} else if order.Guest && order.Total > 0 {
		intent, err := draftorderhelp.ChargePaymentIntent(order.StripePaymentIntentID, newPaymentMethod, false, *order.GuestStripeID)
		if err != nil {
			return err, nil
		}
		order.StripePaymentIntentID = intent.ID
	} else if order.Total > 0 {
		intent, err := draftorderhelp.ChargePaymentIntent(order.StripePaymentIntentID, newPaymentMethod, saveMethod, cust.StripeID)
		if err != nil {
			return err, nil
		}
		order.StripePaymentIntentID = intent.ID
	}

	resp, err := orderhelp.PostOrderToPrintful(order, store, mutexes, tools)
	if err != nil {
		go emails.AlertRecoverableOrderSubmitError(store, draftID, order.ID.Hex(), "Unable to post order to printful after charging", tools, order, draft, resp, true, err)
	}

	if err := orderhelp.ConfirmOrderPostResponse(resp, order); err != nil {
		go emails.AlertRecoverableOrderSubmitError(store, draftID, order.ID.Hex(), "Bad response from posting order to printful after charging", tools, order, draft, resp, false, err)
	}

	gcErr, discErr, worked := s.UseDiscountsAndGiftCards(order, guestID, customerID, dts)
	if !worked {
		if gcErr != nil {
			go emails.AlertRecoverableOrderSubmitError(store, draftID, order.ID.Hex(), "Unable to post order to mark charging of gift cards after using", tools, order, draft, nil, true, gcErr)
		}
		if discErr != nil {
			go emails.AlertRecoverableOrderSubmitError(store, draftID, order.ID.Hex(), "Unable to post order to mark use of of discount after using", tools, order, draft, nil, false, discErr)
		}
	}

	if err := s.MarkOrderAndDraftAsSuccess(order, draft, ds); err != nil {
		go emails.AlertRecoverableOrderSubmitError(store, draftID, order.ID.Hex(), "Unable to save order and draft order after successful creation", tools, order, draft, nil, false, err)
	}

	if err := orderhelp.OrderEmailWithProfit(resp, order, tools, store); err != nil {
		go emails.AlertRecoverableOrderSubmitError(store, draftID, order.ID.Hex(), "Unable to send email of success to creat the order", tools, order, draft, nil, false, err)
	}

	vids := []int{}
	for _, l := range order.Lines {
		varInt, err := strconv.Atoi(l.VariantID)
		if err != nil {
			log.Printf("Unable to convert variant ID: %s on order: %s, in store: %s to int\n", l.VariantID, order.ID.Hex(), store)
			continue
		}
		vids = append(vids, varInt)
	}

	if err := ls.UpdateLastOrdersList(store, customerID, order.DateCreated, order.ID.Hex(), vids, ps); err != nil {
		log.Printf("Unable to update last orders list for order: %s, in store: %s; error: %v\n", order.ID.Hex(), store, err)
	}

	return nil, nil
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

func (s *orderService) MarkOrderAndDraftAsSuccess(order *models.Order, draft *models.DraftOrder, ds *draftOrderService) error {
	now := time.Now()

	order.Status = "Procesed"
	order.DateProcessedPrintful = now

	draft.Status = "Succeeded"
	draft.DateSucceeded = now

	if err := s.orderRepo.Update(order); err != nil {
		return err
	}

	if err := ds.draftOrderRepo.Update(draft); err != nil {
		return err
	}

	return nil
}

// Actual order, display order doesn't belong to this account (guest), error
func (s *orderService) RenderOrder(orderID, guestID string, customerID int) (*models.Order, bool, error) {
	o, err := s.orderRepo.Read(orderID)
	if err != nil {
		return nil, false, err
	}

	if !o.Guest && o.CustomerID != customerID {
		return nil, false, errors.New("order does not belong to customer")
	}

	if o.Status == "Created" {
		return nil, false, errors.New("order does not exist yet")
	}

	if o.Guest && customerID > 0 {
		return o, true, nil
	}

	return o, false, nil
}

func (s *orderService) GetOrdersList(customerID int, fromURL url.Values) (models.OrderRender, error) {
	ret := models.OrderRender{}

	sort, desc, page := orderhelp.ParseQueryParams(fromURL)

	perPage := config.ORDERLEN

	offset := (perPage * page) - perPage

	orders, err := s.orderRepo.GetOrders(customerID, perPage+1, offset, sort, desc)
	if err != nil {
		return ret, err
	}

	more := false
	if len(orders) > perPage {
		orders = orders[:perPage]
		more = true
	}

	ret.Orders = orders
	ret.Descending = desc
	ret.SortColumn = sort
	ret.Next = more
	ret.Previous = page > 1

	return ret, nil
}
