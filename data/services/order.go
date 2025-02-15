package services

import (
	"beam/background/emails"
	"beam/config"
	"beam/data/models"
	"beam/data/repositories"
	"beam/data/services/draftorderhelp"
	"beam/data/services/orderhelp"
	"errors"
	"fmt"
	"log"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"
)

type OrderService interface {
	SubmitOrder(dpi *DataPassIn, draftID, newPaymentMethod string, saveMethod bool, useExisting bool, ds *draftOrderService, cs *customerService, ps *productService, tools *config.Tools) (error, error)
	CompleteOrder(store, orderID string, ds *draftOrderService, dts *discountService, ls *listService, ps *productService, ss *sessionService, mutexes *config.AllMutexes, tools *config.Tools)
	UseDiscountsAndGiftCards(dpi *DataPassIn, order *models.Order, ds *discountService) (error, error, bool)
	MarkOrderAndDraftAsSuccess(order *models.Order, draft *models.DraftOrder, ds *draftOrderService) error
	RenderOrder(dpi *DataPassIn, orderID string) (*models.Order, bool, error)
	GetOrdersList(dpi *DataPassIn, fromURL url.Values) (models.OrderRender, error)
}

type orderService struct {
	orderRepo repositories.OrderRepository
}

func NewOrderService(orderRepo repositories.OrderRepository) OrderService {
	return &orderService{orderRepo: orderRepo}
}

// Charging error, internal error
func (s *orderService) SubmitOrder(dpi *DataPassIn, draftID, newPaymentMethod string, saveMethod bool, useExisting bool, ds *draftOrderService, cs *customerService, ps *productService, tools *config.Tools) (error, error) {

	start := time.Now()

	draft, err := ds.GetDraftPtl(draftID, dpi.GuestID, dpi.CustomerID)
	if err != nil {
		return nil, err
	}

	var cust *models.Customer
	if dpi.CustomerID > 0 && !draft.Guest {
		cust, err = cs.GetCustomerByID(dpi.CustomerID)
		if err != nil {
			return nil, err
		}
	}

	if useExisting {
		if draft.ExistingPaymentMethod.ID == "" {
			return errors.New("requires a chosen payment method if using existing payment method"), nil
		} else if !(dpi.CustomerID > 0 && !draft.Guest) {
			return errors.New("requires non guest order if using existing payment method"), nil
		}
	}

	dvids := []int{}
	vinv := map[int]int{}
	for _, l := range draft.Lines {
		dvids = append(dvids, l.VariantID)
		vinv[l.VariantID] += l.Quantity
	}

	mapped, works, err := ps.ConfirmDraftOrderProducts(dpi, vinv, dvids)
	if err != nil {
		return nil, fmt.Errorf("unable to query lim vars for draft order: %s, store: %s,  err: %v", draftID, dpi.Store, err)
	} else if !works {
		var builder strings.Builder
		for id, val := range mapped {
			if !val.Possible {
				builder.WriteString(strconv.Itoa(id))
				builder.WriteString(",")
			}
		}
		falseVarIDs := builder.String()
		return nil, fmt.Errorf("nonexistent or low inventory vars for draft order: %s, store: %s, list: %s", draftID, dpi.Store, falseVarIDs)
	}

	order := orderhelp.CreateOrderFromDraft(draft, dpi.SessionID, dpi.AffiliateCode, dpi.AffiliateID)

	if err := s.orderRepo.CreateOrder(order); err != nil {
		return nil, err
	}

	draft.Status = "Submitted"
	draft.DateConverted = time.Now()

	if err := ds.draftOrderRepo.Update(draft); err != nil {
		go emails.AlertRecoverableOrderSubmitError(dpi.Store, draftID, order.ID.Hex(), "Error when updating draft for order on mongodb", tools, order, draft, nil, false, err)
	}

	if useExisting && order.Total > 0 {
		intent, err := draftorderhelp.CreateAndChargePaymentIntent(draft.ExistingPaymentMethod.ID, cust.StripeID, order.Total)
		if err != nil {
			return err, nil
		}
		draft.StripePaymentIntentID = intent.ID
		order.StripePaymentIntentID = intent.ID
	} else if order.Guest && order.Total > 0 {
		intent, err := draftorderhelp.ChargePaymentIntent(order.StripePaymentIntentID, newPaymentMethod, false, order.GuestStripeID)
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

	elapsed := time.Since(start)
	if elapsed < 5*time.Second {
		time.Sleep(5*time.Second - elapsed)
	}

	return nil, nil

}

func (s *orderService) CompleteOrder(store, orderID string, ds *draftOrderService, dts *discountService, ls *listService, ps *productService, ss *sessionService, mutexes *config.AllMutexes, tools *config.Tools) {

	order, err := s.orderRepo.Read(orderID)
	if err != nil {
		log.Printf("Unable to retrieve order from ID for order confirmation; store; %s; orderID: %s; err: %v\n", store, orderID, err)
		return
	}

	draft, err := ds.GetDraftPtl(order.DraftOrderID, order.GuestID, order.CustomerID)
	if err != nil {
		log.Printf("Unable to retrieve draft order from ID for order confirmation; store; %s; orderID: %s; draft orderID: %s; err: %v\n", store, orderID, order.DraftOrderID, err)
		return
	}

	dpi := &DataPassIn{
		Store:         store,
		GuestID:       order.GuestID,
		CustomerID:    order.CustomerID,
		SessionID:     order.SessionID,
		AffiliateID:   order.AffiliateID,
		AffiliateCode: order.AffiliateCode,
	}

	resp, err := orderhelp.PostOrderToPrintful(order, dpi.Store, mutexes, tools)
	if err != nil {
		go emails.AlertRecoverableOrderSubmitError(dpi.Store, order.DraftOrderID, order.ID.Hex(), "Unable to post order to printful after charging", tools, order, draft, resp, true, err)
	}

	if err := orderhelp.ConfirmOrderPostResponse(resp, order); err != nil {
		go emails.AlertRecoverableOrderSubmitError(dpi.Store, order.DraftOrderID, order.ID.Hex(), "Bad response from posting order to printful after charging", tools, order, draft, resp, false, err)
	}

	gcErr, discErr, worked := s.UseDiscountsAndGiftCards(dpi, order, dts)
	if !worked {
		if gcErr != nil {
			go emails.AlertRecoverableOrderSubmitError(dpi.Store, order.DraftOrderID, order.ID.Hex(), "Unable to post order to mark charging of gift cards after using", tools, order, draft, nil, true, gcErr)
		}
		if discErr != nil {
			go emails.AlertRecoverableOrderSubmitError(dpi.Store, order.DraftOrderID, order.ID.Hex(), "Unable to post order to mark use of of discount after using", tools, order, draft, nil, false, discErr)
		}
	}

	if err := s.MarkOrderAndDraftAsSuccess(order, draft, ds); err != nil {
		go emails.AlertRecoverableOrderSubmitError(dpi.Store, order.DraftOrderID, order.ID.Hex(), "Unable to save order and draft order after successful creation", tools, order, draft, nil, false, err)
	}

	if err := orderhelp.OrderEmailWithProfit(resp, order, tools, dpi.Store); err != nil {
		go emails.AlertRecoverableOrderSubmitError(dpi.Store, order.DraftOrderID, order.ID.Hex(), "Unable to send email of success to creat the order", tools, order, draft, nil, false, err)
	}

	vids := []int{}
	dec := map[int]int{}
	handles := []string{}
	for _, l := range order.Lines {
		vids = append(vids, l.VariantID)

		dec[l.VariantID] = l.Quantity

		if !slices.Contains(handles, l.Handle) {
			handles = append(handles, l.Handle)
		}
	}

	if err := ps.SetInventoryFromOrder(dpi, dec, handles, order.ID.Hex(), tools); err != nil {
		log.Printf("Unable to update inventory for order: %s, in store: %s; error: %v\n", order.ID.Hex(), dpi.Store, err)
	}

	if err := ls.UpdateLastOrdersList(dpi, order.DateCreated, order.ID.Hex(), vids, ps); err != nil {
		log.Printf("Unable to update last orders list for order: %s, in store: %s; error: %v\n", order.ID.Hex(), dpi.Store, err)
	}

	ss.AddAffiliateSale(dpi, order.ID.Hex())
}

// Giftcard error, discount error, both worked
func (s *orderService) UseDiscountsAndGiftCards(dpi *DataPassIn, order *models.Order, ds *discountService) (error, error, bool) {

	gcErr, discErr := error(nil), error(nil)

	if len(order.GiftCards) != 0 {

		gcsAndAmounts := map[[2]string]int{}
		for _, gc := range order.GiftCards {
			gcsAndAmounts[[2]string{gc.Code, gc.Pin}] = gc.Charged
		}

		gcErr = ds.UseMultipleGiftCards(gcsAndAmounts, dpi.CustomerID, dpi.GuestID, order.ID.Hex(), dpi.SessionID)

		if gcErr != nil {
			return gcErr, nil, false
		}

	}

	if order.OrderDiscount.DiscountCode != "" {

		discErr = ds.UseDiscountCode(order.OrderDiscount.DiscountCode, dpi.GuestID, order.ID.Hex(), dpi.SessionID, order.Subtotal, dpi.CustomerID, order.Guest)

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
func (s *orderService) RenderOrder(dpi *DataPassIn, orderID string) (*models.Order, bool, error) {
	o, err := s.orderRepo.Read(orderID)
	if err != nil {
		return nil, false, err
	}

	if !o.Guest && o.CustomerID != dpi.CustomerID {
		return nil, false, errors.New("order does not belong to customer")
	}

	if o.Status == "Created" {
		return nil, false, errors.New("order does not exist yet")
	}

	if o.Guest && dpi.CustomerID > 0 {
		return o, true, nil
	}

	return o, false, nil
}

func (s *orderService) GetOrdersList(dpi *DataPassIn, fromURL url.Values) (models.OrderRender, error) {
	ret := models.OrderRender{}

	sort, desc, page := orderhelp.ParseQueryParams(fromURL)

	perPage := config.ORDERLEN

	offset := (perPage * page) - perPage

	orders, err := s.orderRepo.GetOrders(dpi.CustomerID, perPage+1, offset, sort, desc)
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
