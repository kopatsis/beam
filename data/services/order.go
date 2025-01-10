package services

import (
	"beam/data/repositories"
)

type OrderService interface {
	SubmitOrder(draftID, guestID string, customerID int, ds *draftOrderService)
}

type orderService struct {
	orderRepo repositories.OrderRepository
}

func NewOrderService(orderRepo repositories.OrderRepository) OrderService {
	return &orderService{orderRepo: orderRepo}
}

func (s *orderService) SubmitOrder(draftID, guestID string, customerID int, ds *draftOrderService) {
	panic("reee")
}
