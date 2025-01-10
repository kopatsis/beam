package services

import (
	"beam/data/repositories"
)

type OrderService interface {
	SubmitOrder(draftID, guestID, newPaymentMethod string, customerID int, saveMethod bool, useExisting bool, ds *draftOrderService) error
}

type orderService struct {
	orderRepo repositories.OrderRepository
}

func NewOrderService(orderRepo repositories.OrderRepository) OrderService {
	return &orderService{orderRepo: orderRepo}
}

func (s *orderService) SubmitOrder(draftID, guestID, newPaymentMethod string, customerID int, saveMethod bool, useExisting bool, ds *draftOrderService) error {
	panic("reee")
}
