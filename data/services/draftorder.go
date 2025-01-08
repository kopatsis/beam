package services

import (
	"beam/config"
	"beam/data/models"
	"beam/data/repositories"
	"beam/data/services/draftorderhelp"
	"errors"
	"fmt"
	"sync"
)

type DraftOrderService interface {
	CreateDraftOrder(name string, customerID int, guestID string, crs *cartService, pds *productService, cts *customerService) (*models.DraftOrder, error)
	GetDraftOrder(name, draftID, guestID string, customerID int, cts *customerService) (*models.DraftOrder, string, error)
	AddAddressToDraft(name, draftID, guestID, ip string, customerID int, cts *customerService, contact *models.Contact, addToCust bool, mutexes *config.AllMutexes, tools *config.Tools) (*models.DraftOrder, error)
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

	if err := s.draftOrderRepo.Create(draft); err != nil {
		return nil, err
	}

	return draft, nil
}

func (s *draftOrderService) GetDraftOrder(name, draftID, guestID string, customerID int, cts *customerService) (*models.DraftOrder, string, error) {

	draft, err := s.draftOrderRepo.Read(draftID)
	if err != nil {
		return nil, "", err
	}

	if draft.Status == "Failed" || draft.Status == "Submitted" || draft.Status == "Expired" || draft.Status == "Abandoned" {
		return nil, draft.Status, nil
	}

	if customerID > 0 {
		contacts, err := cts.customerRepo.GetContactsWithDefault(customerID)
		if err == nil {
			if len(contacts) < 1 {
				draft.ShippingContact = nil
				draft.ListedContacts = nil
			} else {
				found := false
				for _, c := range contacts {
					if c.ID == draft.ShippingContact.ID {
						found = true
						break
					}
				}
				if !found {
					draft.ShippingContact = contacts[0]
				}
				draft.ListedContacts = contacts
			}
			go s.draftOrderRepo.Update(draft)
		}
	}

	return draft, "", nil
}

func (s *draftOrderService) PostRenderUpdate(draftID string, customerID int, cts *customerService) (*models.DraftOrder, error) {

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

	wg.Wait()

	if draftErr != nil {
		return nil, draftErr
	}

	if customerErr != nil {
		return nil, customerErr
	}

	pms, err := draftorderhelp.GetAllPaymentMethods(cust.StripeID)
	if err != nil {
		return nil, err
	}

	draft.AllPaymentMethods = pms

	in := false
	for _, pm := range pms {
		if pm.ID == draft.ExistingPaymentMethod.ID {
			in = true
		}
	}

	if !in {
		draft.ExistingPaymentMethod = models.PaymentMethodStripe{}
	}

	return draft, nil
}

func (s *draftOrderService) AddAddressToDraft(name, draftID, guestID, ip string, customerID int, cts *customerService, contact *models.Contact, addToCust bool, mutexes *config.AllMutexes, tools *config.Tools) (*models.DraftOrder, error) {
	draft, err := s.draftOrderRepo.Read(draftID)
	if err == nil && (draft.Status == "Failed" || draft.Status == "Submitted" || draft.Status == "Expired" || draft.Status == "Abandoned") {
		err = fmt.Errorf("incorrect status for actions with draft: %s", draft.Status)
	}
	if err != nil {
		return draft, err
	}

	var custErr error = nil
	if addToCust && customerID > 0 {
		custErr = cts.customerRepo.AddContactToCustomer(customerID, contact)
	}

	draft.ShippingContact = contact
	draft.ListedContacts = append([]*models.Contact{contact}, draft.ListedContacts...)

	if err := draftorderhelp.UpdateShippingRates(draft, contact, mutexes, name, ip, tools); err != nil {
		return draft, err
	}

	go s.draftOrderRepo.Update(draft)
	return draft, custErr
}
