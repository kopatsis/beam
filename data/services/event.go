package services

import (
	"beam/data/models"
	"beam/data/repositories"
)

type DataPassIn struct {
	Store         string
	CustomerID    int
	IsLoggedIn    bool
	GuestID       string
	CartID        int
	SessionID     string
	AffiliateID   int
	AffiliateCode string
	DeviceID      string
	Logger        EventService
}

type EventService interface {
	SaveEvent(
		customerID int,
		guestID, eventClassification, eventDescription, eventDetails, specialNote, orderID, draftOrderID, productID, variantID, favesID, savesID, lolistID, cartID, cartLineID, discountID, giftCardID string,
		errors []error,
	)
	SaveEventNew(eventClassification, eventDescription, eventDetails, specialNote string, ids models.EventIDPassIn, errors []error)
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

func (s *eventService) SaveEventNew(eventClassification, eventDescription, eventDetails, specialNote string, ids models.EventIDPassIn, errors []error) {
	s.eventRepo.AddToBatchNew(eventClassification, eventDescription, eventDetails, specialNote, ids, errors)
}
