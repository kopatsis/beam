package services

import (
	"beam/data/models"
	"beam/data/repositories"
)

type CartService interface {
	AddCart(cart models.Cart) error
	GetCartByID(id int) (*models.Cart, error)
	UpdateCart(cart models.Cart) error
	DeleteCart(id int) error
}

type cartService struct {
	cartRepo repositories.CartRepository
}

func NewCartService(cartRepo repositories.CartRepository) CartService {
	return &cartService{cartRepo: cartRepo}
}

func (s *cartService) AddCart(cart models.Cart) error {
	return s.cartRepo.Create(cart)
}

func (s *cartService) GetCartByID(id int) (*models.Cart, error) {
	return s.cartRepo.Read(id)
}

func (s *cartService) UpdateCart(cart models.Cart) error {
	return s.cartRepo.Update(cart)
}

func (s *cartService) DeleteCart(id int) error {
	return s.cartRepo.Delete(id)
}
