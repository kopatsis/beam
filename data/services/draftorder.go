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
)

type DraftOrderService interface {
	CreateDraftOrder(name string, customerID int, guestID string, crs *cartService, pds *productService, cts *customerService) (*models.DraftOrder, error)
	GetDraftOrder(name, draftID, guestID string, customerID int, cts *customerService) (*models.DraftOrder, string, error)
	PostRenderUpdate(ip, name, guestID, draftID string, customerID int, cts *customerService, mutexes *config.AllMutexes, tools *config.Tools) (*models.DraftOrder, error)
	SaveAndUpdatePtl(draft *models.DraftOrder) error
	AddAddressToDraft(name, draftID, guestID, ip string, customerID int, cts *customerService, contact *models.Contact, addToCust bool, mutexes *config.AllMutexes, tools *config.Tools) (*models.DraftOrder, error)
	ChooseAddress(name, draftID, guestID, ip string, addrID, index, customerID int, cts *customerService, mutexes *config.AllMutexes, tools *config.Tools) (*models.DraftOrder, error)
	ChooseShipRate(draftID, guestID, rateName string, customerID int) (*models.DraftOrder, error)
	ChoosePaymentMethod(draftID, guestID, paymentMethodID string, customerID int, cts *customerService) (*models.DraftOrder, error)
	RemovePaymentMethod(draftID, guestID string, customerID int) (*models.DraftOrder, error)
	AddDiscountCode(draftID, guestID, discCode string, customerID int, ds *discountService) (*models.DraftOrder, error)
	RemoveDiscountCode(draftID, guestID string, customerID int) (*models.DraftOrder, error)
	SetTip(draftID, guestID string, customerID, tip int) (*models.DraftOrder, error)
	RemoveTip(draftID, guestID string, customerID int) (*models.DraftOrder, error)
	AddGiftSubjectAndMessage(draftID, guestID, subject, message string, customerID int) (*models.DraftOrder, error)
	CheckDiscountsAndGiftCards(draftID, guestID string, customerID int, ds *discountService) (error, error, error, bool)
}

type draftOrderService struct {
	draftOrderRepo repositories.DraftOrderRepository
}

func NewDraftOrderService(draftRepo repositories.DraftOrderRepository) DraftOrderService {
	return &draftOrderService{draftOrderRepo: draftRepo}
}

func (s *draftOrderService) CreateDraftOrder(name string, customerID int, guestID string, crs *cartService, pds *productService, cts *customerService) (*models.DraftOrder, error) {
	var wg sync.WaitGroup

	cart := models.Cart{}
	cartLines := []models.CartLine{}
	contacts := []*models.Contact{}
	cust := &models.Customer{}
	pMap := map[int]*models.ProductRedis{}
	exists := false

	cartErr, customerErr, contactsErr, productErr := error(nil), error(nil), error(nil), error(nil)

	wg.Add(4)

	go func() {
		defer wg.Done()
		if customerID == 0 && guestID != "" {
			cart, cartLines, exists, cartErr = crs.cartRepo.GetCartWithLinesByGuestID(guestID)
		} else {
			cart, cartLines, exists, cartErr = crs.cartRepo.GetCartWithLinesByCustomerID(customerID)
		}
	}()

	go func() {
		defer wg.Done()
		cust, customerErr = cts.GetCustomerByID(customerID)
	}()

	go func() {
		defer wg.Done()
		contacts, contactsErr = cts.customerRepo.GetContactsWithDefault(customerID)
	}()

	go func() {
		defer wg.Done()
		pMap, productErr = pds.GetProductsMapFromCartLine(name, cartLines)
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
	if !exists {
		return nil, errors.New("no existing cart")
	}

	draft, err := draftorderhelp.CreateDraftOrder(cust, guestID, cart, cartLines, pMap, contacts)
	if err != nil {
		return nil, err
	}

	_, custUpdate, err := draftorderhelp.ConfirmPaymentIntentDraft(draft, cust, guestID)
	if err != nil {
		return nil, err
	}

	if custUpdate {
		go cts.customerRepo.Update(*cust)
	}

	if err := s.draftOrderRepo.Create(draft); err != nil {
		return nil, err
	}

	return draft, nil
}

func (s *draftOrderService) GetDraftOrder(name, draftID, guestID string, customerID int, cts *customerService) (*models.DraftOrder, string, error) {

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
		if customerID > 0 {
			cust, customerErr = cts.GetCustomerByID(customerID)
		}

	}()

	go func() {
		defer wg.Done()
		if customerID > 0 {
			contacts, contactErr = cts.customerRepo.GetContactsWithDefault(customerID)
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

	if customerID > 0 {
		if draft.CustomerID != customerID || draft.CustomerID != cust.ID {
			return draft, "", errors.New("incorrect id for customer")
		} else if cust.Status != "Active" {
			return draft, "", errors.New("draft order for inactive customer")
		} else if draftorderhelp.MergeAddresses(draft, contacts) {
			go s.draftOrderRepo.Update(draft)
		}
	} else {
		if draft.GuestID != &guestID {
			return draft, "", errors.New("incorrect id for guest customer")
		}
	}

	return draft, "", nil
}

// Re-updates the payment methods, the shipping options, the order estimates, and the CA tax if done that way -> check payment intent/update w/ new $, OR create new one
func (s *draftOrderService) PostRenderUpdate(ip, name, guestID, draftID string, customerID int, cts *customerService, mutexes *config.AllMutexes, tools *config.Tools) (*models.DraftOrder, error) {

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
		cust, customerErr = cts.GetCustomerByID(customerID)
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
		shipRateErr = draftorderhelp.UpdateShippingRates(draft, draft.ShippingContact, mutexes, name, ip, tools)
	}()

	go func() {
		defer wg.Done()
		taxRateErr = draftorderhelp.ModifyTaxRate(draft, tools, mutexes)
	}()

	go func() {
		defer wg.Done()
		orderEstErr = draftorderhelp.DraftOrderEstimateUpdate(draft, draft.ShippingContact, mutexes, name, ip, tools)
	}()

	go func() {
		defer wg.Done()
		custUpd := false
		_, custUpd, paymentIntentErr = draftorderhelp.ConfirmPaymentIntentDraft(draft, cust, guestID)
		if custUpd {
			go cts.customerRepo.Update(*cust)
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
		if draft.GuestID != &guestID {
			return draft, errors.New("draft does not belong to customer")
		}
	}
	return draft, nil
}

func (s *draftOrderService) AddAddressToDraft(name, draftID, guestID, ip string, customerID int, cts *customerService, contact *models.Contact, addToCust bool, mutexes *config.AllMutexes, tools *config.Tools) (*models.DraftOrder, error) {
	draft, err := s.GetDraftPtl(draftID, guestID, customerID)
	if err != nil {
		return draft, err
	}

	contact.CustomerID = customerID
	if err := custhelp.VerifyContact(contact, mutexes); err != nil {
		return draft, err
	}

	var custErr error = nil
	if addToCust && customerID > 0 {
		custErr = cts.customerRepo.AddContactToCustomer(contact)
	}

	draft.ShippingContact = contact
	draft.ListedContacts = append([]*models.Contact{contact}, draft.ListedContacts...)

	if err := draftorderhelp.UpdateShippingRates(draft, contact, mutexes, name, ip, tools); err != nil {
		return draft, err
	}

	err = s.SaveAndUpdatePtl(draft)
	if err != nil {
		return draft, err
	}
	return draft, custErr
}

func (s *draftOrderService) ChooseAddress(name, draftID, guestID, ip string, addrID, index, customerID int, cts *customerService, mutexes *config.AllMutexes, tools *config.Tools) (*models.DraftOrder, error) {
	draft, err := s.GetDraftPtl(draftID, guestID, customerID)
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

	if err := draftorderhelp.UpdateShippingRates(draft, draft.ShippingContact, mutexes, name, ip, tools); err != nil {
		return draft, err
	}

	err = s.SaveAndUpdatePtl(draft)

	return draft, err
}

func (s *draftOrderService) ChooseShipRate(draftID, guestID, rateName string, customerID int) (*models.DraftOrder, error) {
	draft, err := s.GetDraftPtl(draftID, guestID, customerID)
	if err != nil {
		return draft, err
	}

	if err := draftorderhelp.UpdateActualShippingRate(draft, rateName); err != nil {
		return draft, err
	}

	err = s.SaveAndUpdatePtl(draft)

	return draft, err
}

func (s *draftOrderService) ChoosePaymentMethod(draftID, guestID, paymentMethodID string, customerID int, cts *customerService) (*models.DraftOrder, error) {
	if customerID <= 0 {
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
		cust, customerErr = cts.GetCustomerByID(customerID)
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

func (s *draftOrderService) RemovePaymentMethod(draftID, guestID string, customerID int) (*models.DraftOrder, error) {
	draft, err := s.GetDraftPtl(draftID, guestID, customerID)
	if err != nil {
		return draft, err
	}

	draft.ExistingPaymentMethod = models.PaymentMethodStripe{}

	err = s.SaveAndUpdatePtl(draft)

	return draft, err
}

func (s *draftOrderService) AddDiscountCode(draftID, guestID, discCode string, customerID int, ds *discountService) (*models.DraftOrder, error) {
	draft, err := s.GetDraftPtl(draftID, guestID, customerID)
	if err != nil {
		return draft, err
	}

	disc, users, err := ds.GetDiscountCodeForDraft(discCode, draft.Subtotal, customerID, customerID <= 1 && guestID != "")
	if err != nil {
		return draft, err
	}

	if err := draftorderhelp.ApplyDiscountToOrder(disc, users, draft); err != nil {
		return draft, err
	}

	err = s.SaveAndUpdatePtl(draft)

	return draft, err
}

func (s *draftOrderService) RemoveDiscountCode(draftID, guestID string, customerID int) (*models.DraftOrder, error) {
	draft, err := s.GetDraftPtl(draftID, guestID, customerID)
	if err != nil {
		return draft, err
	}

	if err := draftorderhelp.RemoveDiscountFromOrder(draft); err != nil {
		return draft, err
	}

	err = s.SaveAndUpdatePtl(draft)

	return draft, err
}

func (s *draftOrderService) SetTip(draftID, guestID string, customerID, tip int) (*models.DraftOrder, error) {
	draft, err := s.GetDraftPtl(draftID, guestID, customerID)
	if err != nil {
		return draft, err
	}

	if err := draftorderhelp.AddTipToOrder(draft, tip); err != nil {
		return draft, err
	}

	err = s.SaveAndUpdatePtl(draft)

	return draft, err
}

func (s *draftOrderService) RemoveTip(draftID, guestID string, customerID int) (*models.DraftOrder, error) {
	draft, err := s.GetDraftPtl(draftID, guestID, customerID)
	if err != nil {
		return draft, err
	}

	if err := draftorderhelp.DeleteTipFromOrder(draft); err != nil {
		return draft, err
	}

	err = s.SaveAndUpdatePtl(draft)

	return draft, err
}

func (s *draftOrderService) AddGiftSubjectAndMessage(draftID, guestID, subject, message string, customerID int) (*models.DraftOrder, error) {
	draft, err := s.GetDraftPtl(draftID, guestID, customerID)
	if err != nil {
		return draft, err
	}

	draftorderhelp.SetGiftMessage(draft, subject, message)

	err = s.draftOrderRepo.Update(draft)

	return draft, err
}

func (s *draftOrderService) AddGiftCard(draftID, guestID, gcCode string, customerID int, ds *discountService) (*models.DraftOrder, error) {
	draft, err := s.GetDraftPtl(draftID, guestID, customerID)
	if err != nil {
		return draft, err
	}

	gc, err := ds.RetrieveGiftCard(gcCode)
	if err != nil {
		return draft, err
	}

	if err := draftorderhelp.AddGiftCardToOrder(gc, draft); err != nil {
		return draft, err
	}

	err = s.SaveAndUpdatePtl(draft)

	return draft, err
}

func (s *draftOrderService) ApplyGiftCard(draftID, guestID string, gcID, customerID, amount int, useMax bool) (*models.DraftOrder, error) {
	draft, err := s.GetDraftPtl(draftID, guestID, customerID)
	if err != nil {
		return draft, err
	}

	if err := draftorderhelp.ApplyGiftCardToOrder(gcID, amount, useMax, draft); err != nil {
		return draft, err
	}

	err = s.SaveAndUpdatePtl(draft)

	return draft, err
}

func (s *draftOrderService) DeApplyGiftCard(draftID, guestID string, gcID, customerID int) (*models.DraftOrder, error) {
	draft, err := s.GetDraftPtl(draftID, guestID, customerID)
	if err != nil {
		return draft, err
	}

	if err := draftorderhelp.ApplyGiftCardToOrder(gcID, 0, false, draft); err != nil {
		return draft, err
	}

	err = s.SaveAndUpdatePtl(draft)

	return draft, err
}

func (s *draftOrderService) RemoveGiftCard(draftID, guestID string, gcID, customerID int) (*models.DraftOrder, error) {
	draft, err := s.GetDraftPtl(draftID, guestID, customerID)
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
func (s *draftOrderService) CheckDiscountsAndGiftCards(draftID, guestID string, customerID int, ds *discountService) (error, error, error, bool) {
	draft, err := s.GetDraftPtl(draftID, guestID, customerID)
	if err != nil {
		return err, nil, nil, false
	}

	if draft.OrderDiscount.DiscountCode != "" && len(draft.GiftCards) != 0 {

		gcsAndAmounts := map[string]int{}
		for _, gc := range draft.GiftCards {
			gcsAndAmounts[gc.Code] = gc.Charged
		}

		gcErr, draftErr := ds.CheckGiftCardsAndDiscountCodes(gcsAndAmounts, draft.OrderDiscount.DiscountCode, draft.Subtotal, customerID, draft.Guest)
		if gcErr == nil && draftErr == nil {
			return nil, nil, nil, true
		}

		return nil, gcErr, draftErr, false

	} else if len(draft.GiftCards) != 0 {

		gcsAndAmounts := map[string]int{}
		for _, gc := range draft.GiftCards {
			gcsAndAmounts[gc.Code] = gc.Charged
		}

		gcErr := ds.CheckMultipleGiftCards(gcsAndAmounts)
		if gcErr == nil {
			return nil, nil, nil, true
		}

		return nil, gcErr, nil, false

	} else if draft.OrderDiscount.DiscountCode != "" {

		draftErr := ds.CheckDiscountCode(draft.OrderDiscount.DiscountCode, draft.Subtotal, customerID, draft.Guest)
		if draftErr == nil {
			return nil, nil, nil, true
		}

		return nil, nil, draftErr, false

	}

	return nil, nil, nil, true
}
