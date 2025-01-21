package services

import (
	"beam/data/repositories"
)

type EventService interface {
	SaveEvent(customerID int, guestID, eventClassification, eventDescription, specialNote, orderID, draftOrderID, productID, listID, cartID, discountID, giftCardID string)
}

type eventService struct {
	eventRepo repositories.EventRepository
}

func NewEventService(eventRepo repositories.EventRepository) EventService {
	return &eventService{eventRepo: eventRepo}
}

func (s *eventService) SaveEvent(customerID int, guestID, eventClassification, eventDescription, specialNote, orderID, draftOrderID, productID, listID, cartID, discountID, giftCardID string) {
	s.eventRepo.AddToBatch(
		customerID,
		guestID, eventClassification, eventDescription, specialNote, orderID, draftOrderID, productID, listID, cartID, discountID, giftCardID,
	)
}
