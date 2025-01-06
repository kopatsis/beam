package services

import (
	"beam/data/models"
	"beam/data/repositories"
	"beam/data/services/draftorderhelp"
	"errors"
	"sync"
)

type DraftOrderService interface {
	CreateDraftOrder(name string, customerID int, guestID string, crs *cartService, pds *productService, cts *customerService) (*models.DraftOrder, error)
	GetDraftOrder(name, draftID, guestID string, customerID int, cts *customerService) (*models.DraftOrder, string, error)
}

type draftOrderService struct {
	draftOrderRepo repositories.DraftOrderRepository
}

func NewDraftOrderService(draftRepo repositories.DraftOrderRepository) DraftOrderService {
	return &draftOrderService{draftOrderRepo: draftRepo}
}

func (s *draftOrderService) CreateDraftOrder(name string, customerID int, guestID string, crs *cartService, pds *productService, cts *customerService) (*models.DraftOrder, error) {
	var cart models.Cart
	var cartLines []models.CartLine
	var contacts []*models.Contact
	var cust *models.Customer
	var pMap map[int]*models.ProductRedis
	var err error
	var exists bool
	var wg sync.WaitGroup

	wg.Add(4)

	go func() {
		defer wg.Done()
		if customerID == 0 && guestID != "" {
			cart, cartLines, exists, err = crs.cartRepo.GetCartWithLinesByGuestID(guestID)
		} else {
			cart, cartLines, exists, err = crs.cartRepo.GetCartWithLinesByCustomerID(customerID)
		}
	}()

	go func() {
		defer wg.Done()
		cust, err = cts.GetCustomerByID(customerID)
	}()

	go func() {
		defer wg.Done()
		contacts, err = cts.customerRepo.GetContactsWithDefault(customerID)
	}()

	go func() {
		defer wg.Done()
		pMap, err = pds.GetProductsMapFromCartLine(name, cartLines)
	}()

	wg.Wait()

	if err != nil {
		return nil, err
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
