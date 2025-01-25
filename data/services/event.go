package services

import (
	"beam/data/repositories"
)

type EventService interface {
	SaveEvent(
		customerID int,
		guestID, eventClassification, eventDescription, eventDetails, specialNote, orderID, draftOrderID, productID, variantID, favesID, savesID, lolistID, cartID, cartLineID, discountID, giftCardID string,
		errors []error,
	)
}

type eventService struct {
	eventRepo repositories.EventRepository
}

func NewEventService(eventRepo repositories.EventRepository) EventService {
	return &eventService{eventRepo: eventRepo}
}

func (s *eventService) SaveEvent(
	customerID int,
	guestID, eventClassification, eventDescription, eventDetails, specialNote, orderID, draftOrderID, productID, variantID, favesID, savesID, lolistID, cartID, cartLineID, discountID, giftCardID string,
	errors []error,
) {
	s.eventRepo.AddToBatch(
		customerID,
		guestID, eventClassification, eventDescription, eventDetails, specialNote, orderID, draftOrderID, productID, variantID, favesID, savesID, lolistID, cartID, cartLineID, discountID, giftCardID,
		errors)
}
