package services

import (
	"beam/data/models"
	"beam/data/repositories"
)

type EventService interface {
	AddEvent(event models.Event) error
	GetEventByID(id string) (*models.Event, error)
	UpdateEvent(event models.Event) error
	DeleteEvent(id string) error
}

type eventService struct {
	eventRepo repositories.EventRepository
}

func NewEventService(eventRepo repositories.EventRepository) EventService {
	return &eventService{eventRepo: eventRepo}
}

func (s *eventService) AddEvent(event models.Event) error {
	return s.eventRepo.Create(event)
}

func (s *eventService) GetEventByID(id string) (*models.Event, error) {
	return s.eventRepo.Read(id)
}

func (s *eventService) UpdateEvent(event models.Event) error {
	return s.eventRepo.Update(event)
}

func (s *eventService) DeleteEvent(id string) error {
	return s.eventRepo.Delete(id)
}
