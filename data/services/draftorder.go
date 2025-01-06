package services

import (
	"beam/data/repositories"
)

type DraftOrderService interface {
}

type draftOrderService struct {
	draftOrderRepo repositories.DraftOrderRepository
}

func NewDraftOrderService(draftRepo repositories.DraftOrderRepository) DraftOrderService {
	return &draftOrderService{draftOrderRepo: draftRepo}
}
