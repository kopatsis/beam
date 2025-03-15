package services

import (
	"beam/background/apidata"
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
	"sync"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderService interface {
	SubmitOrder(dpi *DataPassIn, draftID, newPaymentMethod string, saveMethod bool, useExisting bool, ds DraftOrderService, dts DiscountService, cs CustomerService, ps ProductService, ors OrderService, tools *config.Tools, storeSettings *config.SettingsMutex) (error, error)
	SubmitPayment(dpi *DataPassIn, draftID, newPayment string, saveMethod bool, useExisting bool, ds DraftOrderService, dts DiscountService, cs CustomerService, ps ProductService, ors OrderService, tools *config.Tools, storeSettings *config.SettingsMutex) (error, error)
	CompleteOrder(store, orderID string, cs CustomerService, ds DraftOrderService, dts DiscountService, ls ListService, ps ProductService, ors OrderService, ss SessionService, mutexes *config.AllMutexes, tools *config.Tools)
	FailOrder(store, orderID string)

	UseDiscountsAndGiftCards(dpi *DataPassIn, order *models.Order, ds DiscountService, storeSettings *config.SettingsMutex, tools *config.Tools, cs CustomerService, ors OrderService) (error, error, bool)
	MarkOrderAndDraftAsSuccess(order *models.Order, draft *models.DraftOrder, ds DraftOrderService) error
	RenderOrder(dpi *DataPassIn, orderID string) (*models.Order, bool, error)
	GetOrdersList(dpi *DataPassIn, fromURL url.Values) (models.OrderRender, error)

	CheckInvDiscAndGiftCards(order *models.Order, draft *models.DraftOrder, dpi *DataPassIn, ps ProductService, ds DiscountService, cs CustomerService, storeSettings *config.SettingsMutex, tools *config.Tools, ors OrderService) error

	ShipOrder(store string, payload apidata.PackageShippedPF) error

	GetCheckDateOrders() ([]models.Order, error)
	AdjustCheckOrders(store string, sendEmail, delayCheck []string, tools *config.Tools) (error, error)

	MoveOrderToAccount(dpi *DataPassIn, orderID string) error

	GetOrdersByEmail(email string) (bool, error)
	GetOrdersByEmailAndCustomer(email string, custID int) (bool, error)
}

type orderService struct {
	orderRepo repositories.OrderRepository
}

func NewOrderService(orderRepo repositories.OrderRepository) OrderService {
	return &orderService{orderRepo: orderRepo}
}

// Charging error, internal error
func (s *orderService) SubmitPayment(dpi *DataPassIn, draftID, newPaymentMethod string, saveMethod bool, useExisting bool, ds DraftOrderService, dts DiscountService, cs CustomerService, ps ProductService, ors OrderService, tools *config.Tools, storeSettings *config.SettingsMutex) (error, error) {
	start := time.Now()

	var draft *models.DraftOrder
	var cust *models.Customer
	draftErr, custErr := error(nil), error(nil)

	var wg1 sync.WaitGroup
	wg1.Add(1)
	go func() {
		draft, draftErr = ds.GetDraftPtl(draftID, dpi.GuestID, dpi.CustomerID)
	}()

	if dpi.CustomerID > 0 && !draft.Guest {
		wg1.Add(1)
		go func() {
			cust, custErr = cs.GetCustomerByID(dpi.CustomerID)
		}()
	}

	wg1.Wait()

	if draftErr != nil {
		return nil, draftErr
	} else if custErr != nil {
		return nil, custErr
	}

	if draft.OrderID != "" || draft.Status == "Succceeded" || draft.Status == "Submitted" {
		return nil, errors.New("already paid for order based on draft")
	}

	if draft.Email == "" {
		if cust != nil {
			draft.Email = cust.Email
		} else {
			return nil, errors.New("no email supplied in draft order")
		}
	}

	if useExisting {
		if draft.ExistingPaymentMethod.ID == "" {
			return nil, errors.New("requires a chosen payment method if using existing payment method")
		} else if !(dpi.CustomerID > 0 && !draft.Guest) {
			return nil, errors.New("requires non guest order if using existing payment method")
		}
	}

	if err := s.CheckInvDiscAndGiftCards(nil, draft, dpi, ps, dts, cs, storeSettings, tools, ors); err != nil {
		return nil, err
	}

	orderID, err := s.orderRepo.CreateBlankOrder()
	if err != nil {
		return nil, err
	}

	if useExisting && draft.Total > 0 {
		if cust == nil {
			return nil, errors.New("nil customer for use existing payment method")
		} else if cust.StripeID == "" {
			return nil, errors.New("customer blank stripe id for use existing payment method")
		}
		pmid, err := draftorderhelp.CreatePaymentIntent(cust.StripeID, int64(draft.Total), "usd")
		if err != nil {
			return nil, err
		}
		draft.StripePaymentIntentID = pmid
	}

	if err := orderhelp.IntentToOrderSet(tools.Redis, draft.StripePaymentIntentID, dpi.Store, orderID); err != nil {
		return nil, err
	}

	orderErr, stripeErr := error(nil), error(nil)
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		if draft.Total > 0 {
			if draft.Total < config.MIN_ORDER_PRICE {
				stripeErr = errors.New("minimum order price not met")
				return
			}

			if draft.Guest {
				intent, err := draftorderhelp.ChargePaymentIntent(draft.StripePaymentIntentID, newPaymentMethod, false, draft.GuestStripeID)
				if err != nil {
					stripeErr = err
					return
				} else if intent.Status == "canceled" {
					stripeErr = errors.New("canceled status for intent immediately")
					return
				}
			} else if draft.Total > 0 {
				intent, err := draftorderhelp.ChargePaymentIntent(draft.StripePaymentIntentID, newPaymentMethod, saveMethod, cust.StripeID)
				if err != nil {
					stripeErr = err
					return
				} else if intent.Status == "canceled" {
					stripeErr = errors.New("canceled status for intent immediately")
					return
				}
			}
		}
	}()

	go func() {
		defer wg.Done()
		order := orderhelp.CreateOrderFromDraft(draft, dpi.SessionID, dpi.AffiliateCode, dpi.AffiliateID)
		hexID, err := primitive.ObjectIDFromHex(orderID)
		if err != nil {
			orderErr = err
			return
		}
		order.ID = hexID

		if err := s.orderRepo.Update(order); err != nil {
			orderErr = err
			return
		}

		draft.Status = "Submitted"
		draft.DateConverted = time.Now()
		draft.OrderID = orderID

		if err := ds.Update(draft); err != nil {
			go emails.AlertRecoverableOrderSubmitError(dpi.Store, draftID, order.ID.Hex(), "Error when updating draft for order on mongodb", tools, order, draft, nil, false, err)
		}
	}()

	wg.Wait()

	if stripeErr != nil || orderErr != nil {
		return stripeErr, orderErr
	}

	cancelOut := 10*time.Millisecond - time.Since(start)
	if cancelOut < 500*time.Millisecond {
		return nil, nil
	}

	orderMessage, err := s.orderRepo.PaymentListen(orderID, dpi.Store, cancelOut)
	if err != nil {
		return nil, err
	} else if orderMessage == config.FAILED_ORDER_MESSAGE {
		return errors.New("order failed from webhook side"), nil
	}

	return nil, nil

}

// Charging error, internal error
func (s *orderService) SubmitOrder(dpi *DataPassIn, draftID, newPaymentMethod string, saveMethod bool, useExisting bool, ds DraftOrderService, dts DiscountService, cs CustomerService, ps ProductService, ors OrderService, tools *config.Tools, storeSettings *config.SettingsMutex) (error, error) {

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

	if draft.Email == "" {
		if cust != nil {
			draft.Email = cust.Email
		} else {
			return nil, errors.New("no email supplied in draft order")
		}
	}

	if useExisting {
		if draft.ExistingPaymentMethod.ID == "" {
			return errors.New("requires a chosen payment method if using existing payment method"), nil
		} else if !(dpi.CustomerID > 0 && !draft.Guest) {
			return errors.New("requires non guest order if using existing payment method"), nil
		}
	}

	if err := s.CheckInvDiscAndGiftCards(nil, draft, dpi, ps, dts, cs, storeSettings, tools, ors); err != nil {
		return nil, err
	}

	order := orderhelp.CreateOrderFromDraft(draft, dpi.SessionID, dpi.AffiliateCode, dpi.AffiliateID)

	if err := s.orderRepo.CreateOrder(order); err != nil {
		return nil, err
	}

	draft.Status = "Submitted"
	draft.DateConverted = time.Now()

	if err := ds.Update(draft); err != nil {
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

	if err := orderhelp.IntentToOrderSet(tools.Redis, order.StripePaymentIntentID, dpi.Store, order.ID.Hex()); err != nil {
		return nil, err
	}

	elapsed := time.Since(start)
	if elapsed < 5*time.Second {
		time.Sleep(5*time.Second - elapsed)
	}

	return nil, nil

}

func (s *orderService) CompleteOrder(store, orderID string, cs CustomerService, ds DraftOrderService, dts DiscountService, ls ListService, ps ProductService, ors OrderService, ss SessionService, mutexes *config.AllMutexes, tools *config.Tools) {

	order, err := s.orderRepo.Read(orderID)
	if err != nil {
		log.Printf("Unable to retrieve order from ID for order confirmation; store; %s; orderID: %s; err: %v\n", store, orderID, err)
		return
	}

	if timeOut, err := s.orderRepo.MarkOrderStatusUpdate(order, "Paid"); err != nil {
		log.Printf("Unable to mark order paid from ID for order confirmation; store; %s; orderID: %s; err: %v\n", store, orderID, err)
		return
	} else if timeOut {
		log.Printf("Unable to mark order paid from ID for order confirmation because of timeout awaiting change from status Blank; store; %s; orderID: %s\n", store, orderID)
		return
	}

	if err := s.orderRepo.PaymentPublish(orderID, store, "Success"); err != nil {
		log.Printf("Unable to publish to stream that order paid from ID for order confirmation; store; %s; orderID: %s; err: %v\n", store, orderID, err)
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

	gcErr, discErr, worked := s.UseDiscountsAndGiftCards(dpi, order, dts, &mutexes.Settings, tools, cs, ors)
	if !worked {
		if gcErr != nil {
			go emails.AlertRecoverableOrderSubmitError(dpi.Store, order.DraftOrderID, order.ID.Hex(), "Unable to post order to mark charging of gift cards after using", tools, order, draft, nil, true, gcErr)
		}
		if discErr != nil {
			go emails.AlertRecoverableOrderSubmitError(dpi.Store, order.DraftOrderID, order.ID.Hex(), "Unable to post order to mark use of of discount after using", tools, order, draft, nil, false, discErr)
		}
	}

	resp, err := orderhelp.PostOrderToPrintful(order, dpi.Store, mutexes, tools)
	if err != nil {
		go emails.AlertRecoverableOrderSubmitError(dpi.Store, order.DraftOrderID, order.ID.Hex(), "Unable to post order to printful after charging", tools, order, draft, resp, true, err)
	}

	if err := orderhelp.ConfirmOrderPostResponse(resp, order); err != nil {
		go emails.AlertRecoverableOrderSubmitError(dpi.Store, order.DraftOrderID, order.ID.Hex(), "Bad response from posting order to printful after charging", tools, order, draft, resp, false, err)
	}

	if err := s.MarkOrderAndDraftAsSuccess(order, draft, ds); err != nil {
		go emails.AlertRecoverableOrderSubmitError(dpi.Store, order.DraftOrderID, order.ID.Hex(), "Unable to save order and draft order after successful creation", tools, order, draft, nil, false, err)
	}

	if err := orderhelp.OrderEmailWithProfit(resp, order, tools, dpi.Store); err != nil {
		go emails.AlertRecoverableOrderSubmitError(dpi.Store, order.DraftOrderID, order.ID.Hex(), "Unable to send email of success to creat the order", tools, order, draft, nil, false, err)
	}

	if err := ls.UpdateLastOrdersList(dpi, order.DateCreated, order.ID.Hex(), vids, ps); err != nil {
		log.Printf("Unable to update last orders list for order: %s, in store: %s; error: %v\n", order.ID.Hex(), dpi.Store, err)
	}

	ss.AddAffiliateSale(dpi, order.ID.Hex())
}

func (s *orderService) FailOrder(store, orderID string) {
	order, err := s.orderRepo.Read(orderID)
	if err != nil {
		log.Printf("Unable to retrieve order from ID for order confirmation; store; %s; orderID: %s; err: %v\n", store, orderID, err)
		return
	}

	if timeOut, err := s.orderRepo.MarkOrderStatusUpdate(order, "Failed"); err != nil {
		log.Printf("Unable to mark order paid from ID for order confirmation; store; %s; orderID: %s; err: %v\n", store, orderID, err)
		return
	} else if timeOut {
		log.Printf("Unable to mark order paid from ID for order confirmation because of timeout awaiting change from status Blank; store; %s; orderID: %s\n", store, orderID)
		return
	}

	if err := s.orderRepo.PaymentPublish(orderID, store, config.FAILED_ORDER_MESSAGE); err != nil {
		log.Printf("Unable to publish to stream that order paid from ID for order confirmation; store; %s; orderID: %s; err: %v\n", store, orderID, err)
		return
	}
}

// Giftcard error, discount error, both worked
func (s *orderService) UseDiscountsAndGiftCards(dpi *DataPassIn, order *models.Order, ds DiscountService, storeSettings *config.SettingsMutex, tools *config.Tools, cs CustomerService, ors OrderService) (error, error, bool) {

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

		discErr = ds.UseDiscountCode(order.OrderDiscount.DiscountCode, dpi.GuestID, order.ID.Hex(), dpi.SessionID, dpi.Store, order.Subtotal, dpi.CustomerID, order.Guest, order.Email, storeSettings, tools, cs, ors)

		if discErr != nil {
			return nil, discErr, false
		}
	}

	return nil, nil, true
}

func (s *orderService) MarkOrderAndDraftAsSuccess(order *models.Order, draft *models.DraftOrder, ds DraftOrderService) error {
	now := time.Now()

	order.Status = "Procesed"
	order.DateProcessedPrintful = now

	draft.Status = "Succeeded"
	draft.DateSucceeded = now

	if err := s.orderRepo.Update(order); err != nil {
		return err
	}

	if err := ds.Update(draft); err != nil {
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

func (s *orderService) CheckInvDiscAndGiftCards(order *models.Order, draft *models.DraftOrder, dpi *DataPassIn, ps ProductService, ds DiscountService, cs CustomerService, storeSettings *config.SettingsMutex, tools *config.Tools, ors OrderService) error {
	dvids := []int{}
	vinv := map[int]int{}

	var giftCards [3]*models.OrderGiftCard
	var discCode string
	var orderGuest bool
	var subtotal int

	if order != nil {

		for _, l := range order.Lines {
			dvids = append(dvids, l.VariantID)
			vinv[l.VariantID] += l.Quantity
		}

		giftCards = order.GiftCards
		discCode = order.OrderDiscount.DiscountCode
		orderGuest = order.Guest
		subtotal = order.Subtotal

	} else if draft != nil {

		for _, l := range draft.Lines {
			dvids = append(dvids, l.VariantID)
			vinv[l.VariantID] += l.Quantity
		}

		giftCards = draft.GiftCards
		discCode = draft.OrderDiscount.DiscountCode
		orderGuest = draft.Guest
		subtotal = draft.Subtotal

	} else {
		return errors.New("nil order and draft order passed in")
	}

	mapped, works, err := ps.ConfirmDraftOrderProducts(dpi, vinv, dvids)
	if err != nil {
		return fmt.Errorf("unable to confirm variants for draft order: %s, store: %s,  err: %v", draft.ID.Hex(), dpi.Store, err)
	} else if !works {
		var builder strings.Builder
		for id, val := range mapped {
			if !val.Possible {
				builder.WriteString(strconv.Itoa(id))
				builder.WriteString(",")
			}
		}
		falseVarIDs := builder.String()
		return fmt.Errorf("nonexistent or low inventory vars for draft order: %s, store: %s, list: %s", draft.ID.Hex(), dpi.Store, falseVarIDs)
	}

	if discCode == "" && len(giftCards) == 0 {
		return nil
	}

	if discCode != "" && len(giftCards) != 0 {

		gcsAndAmounts := map[[2]string]int{}
		for _, gc := range giftCards {
			gcsAndAmounts[[2]string{gc.Code, gc.Pin}] = gc.Charged
		}

		gcErr, draftErr := ds.CheckGiftCardsAndDiscountCodes(gcsAndAmounts, discCode, dpi.Store, subtotal, dpi.CustomerID, orderGuest, draft.Email, storeSettings, tools, cs, ors)
		if gcErr == nil && draftErr == nil {
			return nil
		} else if gcErr == nil {
			return draftErr
		} else if draftErr == nil {
			return gcErr
		}

		return fmt.Errorf("errors from both gc and disc; gc: %v; disc: %v", gcErr, draftErr)

	} else if len(giftCards) != 0 {

		gcsAndAmounts := map[[2]string]int{}
		for _, gc := range giftCards {
			gcsAndAmounts[[2]string{gc.Code, gc.Pin}] = gc.Charged
		}

		gcErr := ds.CheckMultipleGiftCards(gcsAndAmounts)
		if gcErr == nil {
			return nil
		}

		return gcErr

	} else if discCode != "" {

		draftErr := ds.CheckDiscountCode(discCode, dpi.Store, subtotal, dpi.CustomerID, orderGuest, draft.Email, storeSettings, tools, cs, ors)
		if draftErr == nil {
			return nil
		}

		return draftErr

	}

	return nil
}

func (s *orderService) ShipOrder(store string, payload apidata.PackageShippedPF) error {
	orderID := payload.Data.Order.ExternalID
	order, err := s.orderRepo.Read(orderID)
	if err != nil {
		return err
	} else if order == nil {
		return errors.New("nil order with ID: " + orderID)
	}
	if order.Status == "Blank" || order.Status == "Cancelled" || order.Status == "AdminError" {
		return errors.New("not allowed to ship an order under status: " + order.Status)
	}

	shipment := payload.Data.Shipment
	t, err := time.Parse("2006-01-02", shipment.ShipDate)
	if err != nil {
		log.Printf("Failed to parse date '2020-05-05': %v", err)
	}
	fulfillmentID := "FL-" + uuid.NewString()

	order.Fulfillments = append(order.Fulfillments, models.OrderFulfillment{
		ID:             fulfillmentID,
		Status:         "Active",
		PrintfulID:     shipment.ID,
		Carrier:        shipment.Carrier,
		Service:        shipment.Service,
		TrackingNumber: shipment.TrackingNumber,
		TrackingURL:    shipment.TrackingURL,
		Created:        time.Unix(int64(shipment.Created), 0),
		ShipDate:       t,
		ShippedAt:      time.Unix(int64(shipment.ShippedAt), 0),
	})

outerItem:
	for _, item := range shipment.Items {

		lineID := item.ItemID
		quantityLeft := item.Quantity
		if quantityLeft <= 0 {
			continue
		}

		for _, ol := range order.Lines {

			for _, originalPF := range ol.PrintfulID {

				fulfillment := originalPF.Fulfillment
				needed := fulfillment.SubLineQuantity - len(fulfillment.OrderFulfillmentIDs)

				if fulfillment.LineItemID == lineID && needed > 0 {

					if needed < quantityLeft {
						for i := 0; i < needed; i++ {
							fulfillment.OrderFulfillmentIDs = append(fulfillment.OrderFulfillmentIDs, fulfillmentID)
						}
						quantityLeft -= needed
					} else {
						for i := 0; i < quantityLeft; i++ {
							fulfillment.OrderFulfillmentIDs = append(fulfillment.OrderFulfillmentIDs, fulfillmentID)
						}
						quantityLeft = 0
					}
				}

				if quantityLeft <= 0 {
					continue outerItem
				}

			}
		}

		if quantityLeft > 0 {
			log.Printf("Unable to allocate shipped variant per line id: %d; fulfillment ID: %s; PF fullfillment ID: %d; orderID: %s; stores: %s\n", lineID, fulfillmentID, shipment.ID, orderID, store)
		}
	}

	return nil
}

func (s *orderService) GetCheckDateOrders() ([]models.Order, error) {
	return s.orderRepo.GetCheckOrders()
}

func (s *orderService) AdjustCheckOrders(store string, sendEmail, delayCheck []string, tools *config.Tools) (error, error) {
	sendError, delayError := error(nil), error(nil)
	if len(sendEmail) > 0 {
		orders, err := s.orderRepo.GetOrdersByIDs(sendEmail)
		if err != nil {
			sendError = err
		} else {
			for _, o := range orders {
				emails.OrderConfirmAndRate(store, o.Email, &o, tools)
			}
			sendError = s.orderRepo.UpdateCheckEmailSent(sendEmail)
		}

	}

	if len(delayCheck) > 0 {
		delayError = s.orderRepo.UpdateCheckDeliveryDate(delayCheck)
	}

	return sendError, delayError
}

func (s *orderService) MoveOrderToAccount(dpi *DataPassIn, orderID string) error {
	order, err := s.orderRepo.Read(orderID)
	if err != nil {
		return err
	} else if order == nil {
		return errors.New("nil order")
	}

	if order.CustomerID == dpi.CustomerID {
		if order.Guest || !order.MovedToAccount {
			if order.Guest {
				order.Guest = false
			}
			if !order.MovedToAccount {
				order.MovedToAccount = true
				order.MovedToAccountDate = time.Now()
			}
			return s.orderRepo.Update(order)
		} else {
			return nil
		}
	}

	if !order.Guest || order.CustomerID != 0 {
		return errors.New("order belongs to a customer already")
	}

	order.CustomerID = dpi.CustomerID
	order.Guest = false
	order.MovedToAccount = true
	order.MovedToAccountDate = time.Now()

	return s.orderRepo.Update(order)
}

func (s *orderService) GetOrdersByEmail(email string) (bool, error) {
	return s.orderRepo.GetOrdersByEmail(email)
}
func (s *orderService) GetOrdersByEmailAndCustomer(email string, custID int) (bool, error) {
	return s.orderRepo.GetOrdersByEmailAndCustomer(email, custID)
}
