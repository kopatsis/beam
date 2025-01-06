package services

import (
	"beam/data/repositories"
)

type OrderService interface {
}

type orderService struct {
	orderRepo repositories.OrderRepository
}

func NewOrderService(orderRepo repositories.OrderRepository) OrderService {
	return &orderService{orderRepo: orderRepo}
}
