package services

import (
	"beam/data/repositories"
)

type EventService interface {
	SaveEvent(customerID int, guestID, eventClassification, eventDescription, specialNote, otherID string) error
}

type eventService struct {
	eventRepo repositories.EventRepository
}

func NewEventService(eventRepo repositories.EventRepository) EventService {
	return &eventService{eventRepo: eventRepo}
}

func (s *eventService) SaveEvent(customerID int, guestID, eventClassification, eventDescription, specialNote, otherID string) error {
	return s.eventRepo.SaveEvent(customerID, guestID, eventClassification, eventDescription, specialNote, otherID)
}
