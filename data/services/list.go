package services

import (
	"beam/data/models"
	"beam/data/repositories"
)

type ListService interface {
	AddList(list models.List) error
	GetListByID(id int) (*models.List, error)
	UpdateList(list models.List) error
	DeleteList(id int) error
}

type listService struct {
	listRepo repositories.ListRepository
}

func NewListService(listRepo repositories.ListRepository) ListService {
	return &listService{listRepo: listRepo}
}

func (s *listService) AddList(list models.List) error {
	return s.listRepo.Create(list)
}

func (s *listService) GetListByID(id int) (*models.List, error) {
	return s.listRepo.Read(id)
}

func (s *listService) UpdateList(list models.List) error {
	return s.listRepo.Update(list)
}

func (s *listService) DeleteList(id int) error {
	return s.listRepo.Delete(id)
}
