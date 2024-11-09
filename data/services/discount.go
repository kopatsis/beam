package services

import (
	"beam/data/models"
	"beam/data/repositories"
)

type DiscountService interface {
	AddDiscount(discount models.Discount) error
	GetDiscountByID(id int) (*models.Discount, error)
	UpdateDiscount(discount models.Discount) error
	DeleteDiscount(id int) error
}

type discountService struct {
	discountRepo repositories.DiscountRepository
}

func NewDiscountService(discountRepo repositories.DiscountRepository) DiscountService {
	return &discountService{discountRepo: discountRepo}
}

func (s *discountService) AddDiscount(discount models.Discount) error {
	return s.discountRepo.Create(discount)
}

func (s *discountService) GetDiscountByID(id int) (*models.Discount, error) {
	return s.discountRepo.Read(id)
}

func (s *discountService) UpdateDiscount(discount models.Discount) error {
	return s.discountRepo.Update(discount)
}

func (s *discountService) DeleteDiscount(id int) error {
	return s.discountRepo.Delete(id)
}
