package services

import (
	"beam/background/emails"
	"beam/config"
	"beam/data/models"
	"beam/data/repositories"
	"beam/data/services/discount"
	"errors"
)

type DiscountService interface {
	AddDiscount(discount models.Discount) error
	GetDiscountByID(id int) (*models.Discount, error)
	UpdateDiscount(discount models.Discount) error
	DeleteDiscount(id int) error
	CreateGiftCard(cents int, message string, store string, tools *config.Tools) (int, string, error)
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

func (s *discountService) CreateGiftCard(cents int, message string, store string, tools *config.Tools) (int, string, error) {
	if len(message) > 256 {
		message = message[:255]
	}

	var err error
	idSt, exists, iter := "", false, 0
	for !exists && iter < 10 {
		idSt = discount.GenerateCartID()
		exists, err = s.discountRepo.IDCodeExists(idSt)
		if err != nil {
			return 0, "", err
		}
		if !exists {
			emails.AlertGiftCardID(idSt, iter, store, tools)
		}
	}

	if !exists {
		return 0, "", errors.New("severe issue: could not create an id for gift card in 10 attempts")
	}

	idDB, err := s.discountRepo.CreateGiftCard(idSt, cents, message)
	if err != nil {
		return 0, "", err
	}

	return idDB, idSt, nil
}
