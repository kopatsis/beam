package services

import (
	"beam/data/models"
	"beam/data/repositories"
)

type DraftOrderService interface {
	AddDraftOrder(draftOrder models.DraftOrder) error
	GetDraftOrderByID(id string) (*models.DraftOrder, error)
	UpdateDraftOrder(draftOrder models.DraftOrder) error
	DeleteDraftOrder(id string) error
}

type draftOrderService struct {
	draftOrderRepo repositories.DraftOrderRepository
}

func NewDraftOrderService(draftOrderRepo repositories.DraftOrderRepository) DraftOrderService {
	return &draftOrderService{draftOrderRepo: draftOrderRepo}
}

func (s *draftOrderService) AddDraftOrder(draftOrder models.DraftOrder) error {
	return s.draftOrderRepo.Create(draftOrder)
}

func (s *draftOrderService) GetDraftOrderByID(id string) (*models.DraftOrder, error) {
	return s.draftOrderRepo.Read(id)
}

func (s *draftOrderService) UpdateDraftOrder(draftOrder models.DraftOrder) error {
	return s.draftOrderRepo.Update(draftOrder)
}

func (s *draftOrderService) DeleteDraftOrder(id string) error {
	return s.draftOrderRepo.Delete(id)
}
