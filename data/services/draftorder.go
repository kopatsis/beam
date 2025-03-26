package services

import (
	"beam/config"
	"beam/data/models"
	"beam/data/repositories"
	"beam/data/services/custhelp"
	"beam/data/services/draftorderhelp"
	"errors"
	"fmt"
	"sync"
	"time"
)

type DraftOrderService interface {
	CreateDraftOrder(dpi *DataPassIn, crs CartService, pds ProductService, cts CustomerService) (*models.DraftOrder, error)
	GetDraftOrder(dpi *DataPassIn, draftID string, cts CustomerService) (*models.DraftOrder, string, error)
	PostRenderUpdate(dpi *DataPassIn, ip, draftID string, cts CustomerService, mutexes *config.AllMutexes, tools *config.Tools) (*models.DraftOrder, error)
	SaveAndUpdatePtl(draft *models.DraftOrder) error
	GetDraftPtl(draftID, guestID string, custID int) (*models.DraftOrder, error)
	AddAddressToDraft(dpi *DataPassIn, draftID, ip string, cts CustomerService, contact *models.Contact, addToCust bool, mutexes *config.AllMutexes, tools *config.Tools) (*models.DraftOrder, error)
	ChooseAddress(dpi *DataPassIn, draftID, ip string, addrID, index, customerID int, cts CustomerService, mutexes *config.AllMutexes, tools *config.Tools) (*models.DraftOrder, error)
	ChooseShipRate(dpi *DataPassIn, draftID, rateName string) (*models.DraftOrder, error)
	ChoosePaymentMethod(dpi *DataPassIn, draftID, paymentMethodID string, cts CustomerService) (*models.DraftOrder, error)
	RemovePaymentMethod(dpi *DataPassIn, draftID string) (*models.DraftOrder, error)
	AddDiscountCode(dpi *DataPassIn, draftID, discCode string, ds DiscountService, storeSettings *config.SettingsMutex, tools *config.Tools, cs CustomerService, ors OrderService) (*models.DraftOrder, error)
	RemoveDiscountCode(dpi *DataPassIn, draftID string) (*models.DraftOrder, error)
	SetTip(dpi *DataPassIn, draftID string, tip int) (*models.DraftOrder, error)
	RemoveTip(dpi *DataPassIn, draftID string) (*models.DraftOrder, error)
	AddGiftSubjectAndMessage(dpi *DataPassIn, draftID, subject, message string) (*models.DraftOrder, error)
	AddGiftCard(dpi *DataPassIn, draftID, gcCode, pin string, ds DiscountService) (*models.DraftOrder, error)
	ApplyGiftCard(dpi *DataPassIn, draftID string, gcID, amount int, useMax bool) (*models.DraftOrder, error)
	DeApplyGiftCard(dpi *DataPassIn, draftID string, gcID int) (*models.DraftOrder, error)
	RemoveGiftCard(dpi *DataPassIn, draftID string, gcID int) (*models.DraftOrder, error)
	CheckDiscountsAndGiftCards(dpi *DataPassIn, draftID string, ds DiscountService, storeSettings *config.SettingsMutex, tools *config.Tools, cs CustomerService, ors OrderService) (error, error, error, bool)

	AddGuestInfoToDraft(dpi *DataPassIn, draftID, name, email string, tools *config.Tools) (*models.DraftOrder, error)
	ChangeCustDraftName(dpi *DataPassIn, draftID, name string) (*models.DraftOrder, error)

	Update(draft *models.DraftOrder) error
	MoveDraftToCustomer(dpi *DataPassIn, draftID string, cust *models.Customer, cs CartService, ds DiscountService, tools *config.Tools, storeSettings *config.SettingsMutex, cms CustomerService, ors OrderService) (int, error)
}

type draftOrderService struct {
	draftOrderRepo repositories.DraftOrderRepository
}

func NewDraftOrderService(draftRepo repositories.DraftOrderRepository) DraftOrderService {
	return &draftOrderService{draftOrderRepo: draftRepo}
}

func (s *draftOrderService) CreateDraftOrder(dpi *DataPassIn, crs CartService, pds ProductService, cts CustomerService) (*models.DraftOrder, error) {
	var wg sync.WaitGroup

	cart := &models.Cart{}
	cartLines := []*models.CartLine{}
	contacts := []*models.Contact{}
	cust := &models.Customer{}
	pMap := map[int]*models.ProductRedis{}
	id := 0

	cartErr, customerErr, contactsErr, productErr := error(nil), error(nil), error(nil), error(nil)

	wg.Add(4)

	go func() {
		defer wg.Done()
		id, cart, cartLines, cartErr = crs.GetCartWithLinesAndVerify(dpi)
	}()

	go func() {
		defer wg.Done()
		cust, customerErr = cts.GetCustomerByID(dpi, dpi.CustomerID)
	}()

	go func() {
		defer wg.Done()
		contacts, contactsErr = cts.GetContactsWithDefault(dpi, dpi.CustomerID)
	}()

	go func() {
		defer wg.Done()
		pMap, productErr = pds.GetProductsMapFromCartLine(dpi, dpi.Store, cartLines)
	}()

	wg.Wait()

	if cartErr != nil {
		return nil, cartErr
	}
	if customerErr != nil {
		return nil, customerErr
	}
	if contactsErr != nil {
		return nil, contactsErr
	}
	if productErr != nil {
		return nil, productErr
	}
	if id != dpi.CartID {
		return nil, errors.New("no existing cart")
	}

	draft, err := draftorderhelp.CreateDraftOrder(cust, dpi.GuestID, cart, cartLines, pMap, contacts)
	if err != nil {
		return nil, err
	}

	_, custUpdate, err := draftorderhelp.ConfirmPaymentIntentDraft(draft, cust, dpi.GuestID)
	if err != nil {
		return nil, err
	}

	if custUpdate {
		go cts.Update(dpi, cust)
	}

	if err := s.draftOrderRepo.Create(draft); err != nil {
		return nil, err
	}

	return draft, nil
}

func (s *draftOrderService) GetDraftOrder(dpi *DataPassIn, draftID string, cts CustomerService) (*models.DraftOrder, string, error) {

	var wg sync.WaitGroup

	var draft *models.DraftOrder
	var cust *models.Customer
	var contacts []*models.Contact
	err, customerErr, contactErr := error(nil), error(nil), error(nil)

	wg.Add(3)

	go func() {
		defer wg.Done()
		draft, err = s.draftOrderRepo.Read(draftID)
	}()

	go func() {
		defer wg.Done()
		if dpi.CustomerID > 0 {
			cust, customerErr = cts.GetCustomerByID(dpi, dpi.CustomerID)
		}

	}()

	go func() {
		defer wg.Done()
		if dpi.CustomerID > 0 {
			contacts, contactErr = cts.GetContactsWithDefault(dpi, dpi.CustomerID)
		}
	}()

	wg.Wait()

	if err != nil {
		return nil, "", err
	} else if draft.Status == "Failed" || draft.Status == "Submitted" || draft.Status == "Expired" || draft.Status == "Abandoned" {
		return nil, draft.Status, nil
	} else if customerErr != nil {
		return draft, "", customerErr
	} else if contactErr != nil {
		return draft, "", contactErr
	}

	if dpi.CustomerID > 0 {
		if draft.CustomerID != dpi.CustomerID || draft.CustomerID != cust.ID {
			return draft, "", errors.New("incorrect id for customer")
		} else if cust.Status != "Active" {
			return draft, "", errors.New("draft order for inactive customer")
		} else if draftorderhelp.MergeAddresses(draft, contacts) {
			go s.draftOrderRepo.Update(draft)
		}
	} else {
		if draft.GuestID != dpi.GuestID {
			return draft, "", errors.New("incorrect id for guest customer")
		}
	}

	return draft, "", nil
}

// Re-updates the payment methods, the shipping options, the order estimates, and the CA tax if done that way -> check payment intent/update w/ new $, OR create new one
func (s *draftOrderService) PostRenderUpdate(dpi *DataPassIn, ip, draftID string, cts CustomerService, mutexes *config.AllMutexes, tools *config.Tools) (*models.DraftOrder, error) {

	var wg sync.WaitGroup

	var draft *models.DraftOrder
	var cust *models.Customer
	draftErr, customerErr := error(nil), error(nil)

	wg.Add(2)

	go func() {
		defer wg.Done()
		draft, draftErr = s.draftOrderRepo.Read(draftID)
		if draftErr == nil && (draft.Status == "Failed" || draft.Status == "Submitted" || draft.Status == "Expired" || draft.Status == "Abandoned") {
			draftErr = fmt.Errorf("incorrect status for actions with draft: %s", draft.Status)
		}
	}()

	go func() {
		defer wg.Done()
		cust, customerErr = cts.GetCustomerByID(dpi, dpi.CustomerID)
	}()

	paymentMethodErr, shipRateErr, taxRateErr, orderEstErr, paymentIntentErr := error(nil), error(nil), error(nil), error(nil), error(nil)

	wg.Wait()

	if draftErr != nil {
		return nil, draftErr
	} else if customerErr != nil {
		return draft, customerErr
	}

	wg.Add(5)

	go func() {
		defer wg.Done()
		paymentMethodErr = draftorderhelp.DraftPaymentMethodUpdate(draft, cust.StripeID)
	}()

	go func() {
		defer wg.Done()
		shipRateErr = draftorderhelp.UpdateShippingRates(draft, draft.ShippingContact, mutexes, dpi.Store, ip, tools)
	}()

	go func() {
		defer wg.Done()
		taxRateErr = draftorderhelp.ModifyTaxRate(draft, tools, mutexes)
	}()

	go func() {
		defer wg.Done()
		orderEstErr = draftorderhelp.DraftOrderEstimateUpdate(draft, draft.ShippingContact, mutexes, dpi.Store, ip, tools)
	}()

	go func() {
		defer wg.Done()
		custUpd := false
		_, custUpd, paymentIntentErr = draftorderhelp.ConfirmPaymentIntentDraft(draft, cust, dpi.GuestID)
		if custUpd {
			go cts.Update(dpi, cust)
		}
	}()

	wg.Wait()

	if paymentMethodErr != nil {
		return nil, paymentMethodErr
	} else if shipRateErr != nil {
		return draft, shipRateErr
	} else if taxRateErr != nil {
		return draft, taxRateErr
	} else if orderEstErr != nil {
		return draft, orderEstErr
	} else if paymentIntentErr != nil {
		return draft, paymentIntentErr
	}

	err := s.SaveAndUpdatePtl(draft)

	return draft, err
}

// For use by other methods
func (s *draftOrderService) SaveAndUpdatePtl(draft *models.DraftOrder) error {
	if err := draftorderhelp.UpdateTaxFromRate(draft); err != nil {
		return err
	}

	if err := draftorderhelp.SetTotalsAndEnsure(draft); err != nil {
		return err
	}

	return s.draftOrderRepo.Update(draft)
}

// For use by other methods
func (s *draftOrderService) GetDraftPtl(draftID, guestID string, custID int) (*models.DraftOrder, error) {
	draft, err := s.draftOrderRepo.Read(draftID)
	if err == nil && (draft.Status == "Failed" || draft.Status == "Submitted" || draft.Status == "Expired" || draft.Status == "Abandoned") {
		err = fmt.Errorf("incorrect status for actions with draft: %s", draft.Status)
	}
	if err != nil {
		return draft, err
	}
	if custID > 0 {
		if draft.CustomerID != custID {
			return draft, errors.New("draft does not belong to customer")
		}
	} else {
		if draft.GuestID != guestID {
			return draft, errors.New("draft does not belong to customer")
		}
	}
	return draft, nil
}

func (s *draftOrderService) AddAddressToDraft(dpi *DataPassIn, draftID, ip string, cts CustomerService, contact *models.Contact, addToCust bool, mutexes *config.AllMutexes, tools *config.Tools) (*models.DraftOrder, error) {
	draft, err := s.GetDraftPtl(draftID, dpi.GuestID, dpi.CustomerID)
	if err != nil {
		return draft, err
	}

	contact.CustomerID = dpi.CustomerID
	if err := custhelp.VerifyContact(contact, mutexes); err != nil {
		return draft, err
	}

	var custErr error = nil
	if addToCust && dpi.CustomerID > 0 {
		custErr = cts.AddContactToCustomer(dpi, contact)
	}

	draft.ShippingContact = contact
	draft.ListedContacts = append([]*models.Contact{contact}, draft.ListedContacts...)

	if err := draftorderhelp.UpdateShippingRates(draft, contact, mutexes, dpi.Store, ip, tools); err != nil {
		return draft, err
	}

	err = s.SaveAndUpdatePtl(draft)
	if err != nil {
		return draft, err
	}
	return draft, custErr
}

func (s *draftOrderService) ChooseAddress(dpi *DataPassIn, draftID, ip string, addrID, index, customerID int, cts CustomerService, mutexes *config.AllMutexes, tools *config.Tools) (*models.DraftOrder, error) {
	draft, err := s.GetDraftPtl(draftID, dpi.GuestID, customerID)
	if err != nil {
		return draft, err
	}

	if addrID == 0 {
		for _, c := range draft.ListedContacts {
			if c.ID == addrID {
				draft.ShippingContact = c
				break
			}
		}
	} else {
		if index < 0 {
			return draft, errors.New("choice of address without id must have index > 0")
		} else if index >= len(draft.ListedContacts) {
			return draft, errors.New("choice of address without id must have index < length of list")
		}
		draft.ShippingContact = draft.ListedContacts[index]
	}

	if err := draftorderhelp.UpdateShippingRates(draft, draft.ShippingContact, mutexes, dpi.Store, ip, tools); err != nil {
		return draft, err
	}

	err = s.SaveAndUpdatePtl(draft)

	return draft, err
}

func (s *draftOrderService) ChooseShipRate(dpi *DataPassIn, draftID, rateName string) (*models.DraftOrder, error) {
	draft, err := s.GetDraftPtl(draftID, dpi.GuestID, dpi.CustomerID)
	if err != nil {
		return draft, err
	}

	if err := draftorderhelp.UpdateActualShippingRate(draft, rateName); err != nil {
		return draft, err
	}

	err = s.SaveAndUpdatePtl(draft)

	return draft, err
}

func (s *draftOrderService) ChoosePaymentMethod(dpi *DataPassIn, draftID, paymentMethodID string, cts CustomerService) (*models.DraftOrder, error) {
	if dpi.CustomerID <= 0 {
		return nil, errors.New("unable to do action for nonexistent customer id")
	}

	var wg sync.WaitGroup

	var draft *models.DraftOrder
	var cust *models.Customer
	err, customerErr := error(nil), error(nil)

	wg.Add(2)

	go func() {
		defer wg.Done()
		draft, err = s.draftOrderRepo.Read(draftID)
	}()

	go func() {
		defer wg.Done()
		cust, customerErr = cts.GetCustomerByID(dpi, dpi.CustomerID)
	}()

	wg.Wait()
	if err != nil {
		return draft, err
	} else if customerErr != nil {
		return draft, customerErr
	}

	found := false
	for _, pm := range draft.AllPaymentMethods {
		if pm.ID == paymentMethodID {
			if e := draftorderhelp.ValidatePaymentMethod(cust.StripeID, paymentMethodID); e != nil {
				return draft, e
			}
			draft.ExistingPaymentMethod = pm
			found = true
			break
		}
	}

	if !found {
		return draft, errors.New("no payment method with that id listed")
	}

	err = s.SaveAndUpdatePtl(draft)

	return draft, err
}

func (s *draftOrderService) RemovePaymentMethod(dpi *DataPassIn, draftID string) (*models.DraftOrder, error) {
	draft, err := s.GetDraftPtl(draftID, dpi.GuestID, dpi.CustomerID)
	if err != nil {
		return draft, err
	}

	draft.ExistingPaymentMethod = models.PaymentMethodStripe{}

	err = s.SaveAndUpdatePtl(draft)

	return draft, err
}

func (s *draftOrderService) AddDiscountCode(dpi *DataPassIn, draftID, discCode string, ds DiscountService, storeSettings *config.SettingsMutex, tools *config.Tools, cs CustomerService, ors OrderService) (*models.DraftOrder, error) {
	draft, err := s.GetDraftPtl(draftID, dpi.GuestID, dpi.CustomerID)
	if err != nil {
		return draft, err
	}

	disc, users, err := ds.GetDiscountCodeForDraft(dpi, discCode, dpi.Store, draft.Subtotal, dpi.CustomerID, dpi.CustomerID < 1 && dpi.GuestID != "", draft.Email, storeSettings, tools, cs, ors)
	if err != nil {
		return draft, err
	}

	if err := draftorderhelp.ApplyDiscountToOrder(disc, users, draft); err != nil {
		return draft, err
	}

	err = s.SaveAndUpdatePtl(draft)

	return draft, err
}

func (s *draftOrderService) RemoveDiscountCode(dpi *DataPassIn, draftID string) (*models.DraftOrder, error) {
	draft, err := s.GetDraftPtl(draftID, dpi.GuestID, dpi.CustomerID)
	if err != nil {
		return draft, err
	}

	if err := draftorderhelp.RemoveDiscountFromOrder(draft); err != nil {
		return draft, err
	}

	err = s.SaveAndUpdatePtl(draft)

	return draft, err
}

func (s *draftOrderService) SetTip(dpi *DataPassIn, draftID string, tip int) (*models.DraftOrder, error) {
	draft, err := s.GetDraftPtl(draftID, dpi.GuestID, dpi.CustomerID)
	if err != nil {
		return draft, err
	}

	if err := draftorderhelp.AddTipToOrder(draft, tip); err != nil {
		return draft, err
	}

	err = s.SaveAndUpdatePtl(draft)

	return draft, err
}

func (s *draftOrderService) RemoveTip(dpi *DataPassIn, draftID string) (*models.DraftOrder, error) {
	draft, err := s.GetDraftPtl(draftID, dpi.GuestID, dpi.CustomerID)
	if err != nil {
		return draft, err
	}

	if err := draftorderhelp.DeleteTipFromOrder(draft); err != nil {
		return draft, err
	}

	err = s.SaveAndUpdatePtl(draft)

	return draft, err
}

func (s *draftOrderService) AddGiftSubjectAndMessage(dpi *DataPassIn, draftID, subject, message string) (*models.DraftOrder, error) {
	draft, err := s.GetDraftPtl(draftID, dpi.GuestID, dpi.CustomerID)
	if err != nil {
		return draft, err
	}

	draftorderhelp.SetGiftMessage(draft, subject, message)

	err = s.draftOrderRepo.Update(draft)

	return draft, err
}

func (s *draftOrderService) AddGiftCard(dpi *DataPassIn, draftID, gcCode, pin string, ds DiscountService) (*models.DraftOrder, error) {
	draft, err := s.GetDraftPtl(draftID, dpi.GuestID, dpi.CustomerID)
	if err != nil {
		return draft, err
	}

	if !draftorderhelp.MoreGiftCardsAllowed(draft) {
		return draft, errors.New("maximum number of gift cards to add reached")
	}

	gc, err := ds.RetrieveGiftCard(dpi, gcCode, pin)
	if err != nil {
		return draft, err
	}

	if err := draftorderhelp.AddGiftCardToOrder(gc, draft); err != nil {
		return draft, err
	}

	err = s.SaveAndUpdatePtl(draft)

	return draft, err
}

func (s *draftOrderService) ApplyGiftCard(dpi *DataPassIn, draftID string, gcID, amount int, useMax bool) (*models.DraftOrder, error) {
	draft, err := s.GetDraftPtl(draftID, dpi.GuestID, dpi.CustomerID)
	if err != nil {
		return draft, err
	}

	if err := draftorderhelp.ApplyGiftCardToOrder(gcID, amount, useMax, draft); err != nil {
		return draft, err
	}

	err = s.SaveAndUpdatePtl(draft)

	return draft, err
}

func (s *draftOrderService) DeApplyGiftCard(dpi *DataPassIn, draftID string, gcID int) (*models.DraftOrder, error) {
	draft, err := s.GetDraftPtl(draftID, dpi.GuestID, dpi.CustomerID)
	if err != nil {
		return draft, err
	}

	if err := draftorderhelp.ApplyGiftCardToOrder(gcID, 0, false, draft); err != nil {
		return draft, err
	}

	err = s.SaveAndUpdatePtl(draft)

	return draft, err
}

func (s *draftOrderService) RemoveGiftCard(dpi *DataPassIn, draftID string, gcID int) (*models.DraftOrder, error) {
	draft, err := s.GetDraftPtl(draftID, dpi.GuestID, dpi.CustomerID)
	if err != nil {
		return draft, err
	}

	if err := draftorderhelp.RemoveGiftCardFromOrder(gcID, draft); err != nil {
		return draft, err
	}

	err = s.SaveAndUpdatePtl(draft)

	return draft, err
}

// draftErr error, gcErr error, draftErr error, passes bool
func (s *draftOrderService) CheckDiscountsAndGiftCards(dpi *DataPassIn, draftID string, ds DiscountService, storeSettings *config.SettingsMutex, tools *config.Tools, cs CustomerService, ors OrderService) (error, error, error, bool) {
	draft, err := s.GetDraftPtl(draftID, dpi.GuestID, dpi.CustomerID)
	if err != nil {
		return err, nil, nil, false
	}

	if draft.OrderDiscount.DiscountCode != "" && len(draft.GiftCards) != 0 {

		gcsAndAmounts := map[[2]string]int{}
		for _, gc := range draft.GiftCards {
			gcsAndAmounts[[2]string{gc.Code, gc.Pin}] = gc.Charged
		}

		gcErr, draftErr := ds.CheckGiftCardsAndDiscountCodes(dpi, gcsAndAmounts, draft.OrderDiscount.DiscountCode, dpi.Store, draft.Subtotal, dpi.CustomerID, draft.Guest, draft.Email, storeSettings, tools, cs, ors)
		if gcErr == nil && draftErr == nil {
			return nil, nil, nil, true
		}

		return nil, gcErr, draftErr, false

	} else if len(draft.GiftCards) != 0 {

		gcsAndAmounts := map[[2]string]int{}
		for _, gc := range draft.GiftCards {
			gcsAndAmounts[[2]string{gc.Code, gc.Pin}] = gc.Charged
		}

		gcErr := ds.CheckMultipleGiftCards(dpi, gcsAndAmounts)
		if gcErr == nil {
			return nil, nil, nil, true
		}

		return nil, gcErr, nil, false

	} else if draft.OrderDiscount.DiscountCode != "" {

		draftErr := ds.CheckDiscountCode(dpi, draft.OrderDiscount.DiscountCode, dpi.Store, draft.Subtotal, dpi.CustomerID, draft.Guest, draft.Email, storeSettings, tools, cs, ors)
		if draftErr == nil {
			return nil, nil, nil, true
		}

		return nil, nil, draftErr, false

	}

	return nil, nil, nil, true
}

func (s *draftOrderService) AddGuestInfoToDraft(dpi *DataPassIn, draftID string, email, name string, tools *config.Tools) (*models.DraftOrder, error) {
	if !custhelp.VerifyEmail(email, tools) {
		return nil, errors.New("invalid email")
	}

	if len(name) > 140 {
		name = name[:139]
	}

	draft, err := s.GetDraftPtl(draftID, dpi.GuestID, dpi.CustomerID)
	if err != nil {
		return nil, err
	}

	if !draft.Guest {
		return nil, errors.New("must be guest order to change both name and email")
	}

	draft.Name = name
	draft.Email = email

	err = s.SaveAndUpdatePtl(draft)
	return draft, err
}

func (s *draftOrderService) ChangeCustDraftName(dpi *DataPassIn, draftID, name string) (*models.DraftOrder, error) {
	if len(name) > 140 {
		name = name[:139]
	}

	draft, err := s.GetDraftPtl(draftID, dpi.GuestID, dpi.CustomerID)
	if err != nil {
		return nil, err
	}

	if draft.Guest {
		return nil, errors.New("must be customer order, not guest")
	}

	draft.Name = name

	err = s.SaveAndUpdatePtl(draft)
	return draft, err
}

func (s *draftOrderService) Update(draft *models.DraftOrder) error {
	return s.draftOrderRepo.Update(draft)
}

func (s *draftOrderService) MoveDraftToCustomer(dpi *DataPassIn, draftID string, cust *models.Customer, cs CartService, ds DiscountService, tools *config.Tools, storeSettings *config.SettingsMutex, cms CustomerService, ors OrderService) (int, error) {
	draft, err := s.draftOrderRepo.Read(draftID)
	if err != nil {
		return 0, err
	}

	if draft.Status == "Failed" || draft.Status == "Submitted" || draft.Status == "Expired" || draft.Status == "Abandoned" {
		return 0, fmt.Errorf("incorrect status for actions with draft: %s", draft.Status)
	}

	if draft.CustomerID == dpi.CustomerID {
		return draft.CartID, nil
	}

	if err := cs.MoveCart(&DataPassIn{CustomerID: dpi.CustomerID, CartID: draft.CartID}); err != nil {
		return 0, err
	}

	draft.Guest = false
	draft.GuestID = dpi.GuestID
	draft.CustomerID = dpi.CustomerID
	draft.Email = cust.Email
	draft.MovedToAccount = true
	draft.MovedToAccountDate = time.Now()

	draft.GiftCards = [3]*models.OrderGiftCard{}
	draft.GiftCardSum = 0
	draft.PostGiftCardTotal = draft.PreGiftCardTotal
	draft.Total = draft.PostGiftCardTotal + draft.GiftCardBuyTotal

	draft.StripeMethodID = ""
	draft.NewPaymentMethodID = ""
	draft.AllPaymentMethods = []models.PaymentMethodStripe{}
	draft.ExistingPaymentMethod = models.PaymentMethodStripe{}
	draft.StripePaymentIntentID = ""
	draft.GuestStripeID = ""

	draft.ShippingContact = nil
	draft.ActualRate = models.ShippingRate{}
	draft.CurrentShipping = []models.ShippingRate{}
	draft.AllShippingRates = map[string][]models.ShippingRate{}
	draft.AllOrderEstimates = map[string]models.OrderEstimateCost{}
	draft.CATax = false
	draft.CATaxRate = 0
	draft.ListedContacts = []*models.Contact{}

	if draft.OrderDiscount.DiscountCode != "" {
		disc, users, err := ds.GetDiscountCodeForDraft(dpi, draft.OrderDiscount.DiscountCode, dpi.Store, draft.Subtotal, dpi.CustomerID, false, draft.Email, storeSettings, tools, cms, ors)
		if err != nil {
			return 0, err
		}

		if err := draftorderhelp.ApplyDiscountToOrder(disc, users, draft); err != nil {
			if err := draftorderhelp.RemoveDiscountFromOrder(draft); err != nil {
				return 0, err
			}
		}
	}

	contacts, err := cms.GetContactsWithDefault(dpi, dpi.CustomerID)
	if err != nil {
		return 0, err
	}

	draftorderhelp.MergeAddresses(draft, contacts)

	_, custUpdate, err := draftorderhelp.ConfirmPaymentIntentDraft(draft, cust, dpi.GuestID)
	if err != nil {
		return 0, err
	}

	if custUpdate {
		go cms.Update(dpi, cust)
	}

	return draft.CartID, s.SaveAndUpdatePtl(draft)
}
