package services

import (
	"beam/data/models"
	"beam/data/repositories"
	"beam/data/services/draftorderhelp"
	"errors"
	"sync"
)

type OrderService interface {
	CreateDraftOrder(name string, customerID int, guestID string, crs *cartService, pds *productService, cts *customerService) (*models.DraftOrder, error)
}

type orderService struct {
	orderRepo repositories.OrderRepository
}

func NewOrderService(orderRepo repositories.OrderRepository) OrderService {
	return &orderService{orderRepo: orderRepo}
}

func (s *orderService) CreateDraftOrder(name string, customerID int, guestID string, crs *cartService, pds *productService, cts *customerService) (*models.DraftOrder, error) {
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

	if err := s.orderRepo.CreateDraft(draft); err != nil {
		return nil, err
	}

	return draft, nil
}
