package services

import (
	"beam/data/models"
	"beam/data/repositories"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
)

type DataPassIn struct {
	Store         string
	CustomerID    int
	IsLoggedIn    bool
	GuestID       string
	CartID        int
	SessionID     string
	SessionLineID string
	AffiliateID   int
	AffiliateCode string
	IPAddress     string
	TimeStarted   time.Time
	Logger        EventService
	Logs          []models.EventFinal
	LogsMutex     sync.Mutex
}

func (d *DataPassIn) AddLog(modelName, funcName, errorDesc, extraNote string, err error, ids models.EventPassInFinal) {
	new := models.EventFinal{
		ID:              "EV-" + uuid.NewString(),
		Store:           d.Store,
		Timestamp:       time.Now(),
		SessionID:       d.SessionID,
		SessionLineID:   d.SessionLineID,
		CustomerID:      d.CustomerID,
		GuestID:         d.GuestID,
		ModelName:       modelName,
		FunctionName:    modelName + "." + funcName,
		HasError:        err != nil,
		OptionalNote:    extraNote,
		OrderID:         ids.OrderID,
		DraftOrderID:    ids.DraftOrderID,
		ProductID:       ids.ProductID,
		ProductHandle:   ids.ProductHandle,
		VariantID:       ids.VariantID,
		SavesID:         ids.SavesID,
		FavesID:         ids.FavesID,
		LastOrderListID: ids.LastOrderListID,
		CartID:          ids.CartID,
		CartLineID:      ids.CartLineID,
		DiscountID:      ids.DiscountID,
		DiscountCode:    ids.DiscountCode,
		GiftCardID:      ids.GiftCardID,
		GiftCardCode:    ids.GiftCardCode,
	}

	if err != nil {
		new.Level = "Warn"
		new.ErrorValueSt = err.Error()
		new.ErrorDescription = errorDesc
	} else {
		new.Level = "Trace"
	}

	d.LogsMutex.Lock()
	d.Logs = append(d.Logs, new)
	d.LogsMutex.Unlock()
}

func (d *DataPassIn) MarshalLogs() ([]byte, error) {
	d.LogsMutex.Lock()
	defer d.LogsMutex.Unlock()

	if len(d.Logs) == 0 {
		return nil, nil
	}

	payload, err := json.Marshal(d.Logs)
	if err != nil {
		return nil, err
	}
	return payload, nil
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
