package services

import (
	"beam/data/models"
	"beam/data/repositories"
)

type OrderService interface {
	AddOrder(order models.Order) error
	GetOrderByID(id string) (*models.Order, error)
	UpdateOrder(order models.Order) error
	DeleteOrder(id string) error
}

type orderService struct {
	orderRepo repositories.OrderRepository
}

func NewOrderService(orderRepo repositories.OrderRepository) OrderService {
	return &orderService{orderRepo: orderRepo}
}

func (s *orderService) AddOrder(order models.Order) error {
	return s.orderRepo.Create(order)
}

func (s *orderService) GetOrderByID(id string) (*models.Order, error) {
	return s.orderRepo.Read(id)
}

func (s *orderService) UpdateOrder(order models.Order) error {
	return s.orderRepo.Update(order)
}

func (s *orderService) DeleteOrder(id string) error {
	return s.orderRepo.Delete(id)
}
