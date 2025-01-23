package repositories

import (
	"beam/config"
	"beam/data/models"
	"context"
	"log"
	"slices"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

var (
	validClassifications = []string{
		"Order", "DraftOrder", "Product", "List", "Cart", "Collection", "Discount", "GiftCard",
	}
)

type EventRepository interface {
	AddToBatch(
		customerID int,
		guestID, eventClassification, eventDescription, eventDetails, specialNote, orderID, draftOrderID, productID, listID, cartID, discountID, giftCardID string,
		errors []error,
	)
	FlushBatch()
}

type eventRepo struct {
	coll       *mongo.Collection
	mutex      sync.Mutex
	events     []*models.Event
	saveTicker *time.Ticker
	store      string
}

func NewEventRepository(mdb *mongo.Database, store string) EventRepository {
	repo := &eventRepo{
		coll:       mdb.Collection("Event"),
		events:     make([]*models.Event, 0),
		saveTicker: time.NewTicker(time.Duration(config.BATCH) * time.Second),
		store:      store,
	}

	go func() {
		for range repo.saveTicker.C {
			repo.FlushBatch()
		}
	}()
	defer func() {
		for range repo.saveTicker.C {
			repo.FlushBatch()
		}
		repo.saveTicker.Stop()
	}()

	return repo
}

func (r *eventRepo) AddToBatch(
	customerID int,
	guestID, eventClassification, eventDescription, eventDetails, specialNote, orderID, draftOrderID, productID, listID, cartID, discountID, giftCardID string,
	errors []error,
) {
	if !slices.Contains(validClassifications, eventClassification) {
		log.Printf("invalid event classification: %s\n", eventClassification)
		eventClassification = "Other"
	}

	event := models.Event{
		CustomerID:          customerID,
		GuestID:             guestID,
		Timestamp:           time.Now(),
		EventClassification: eventClassification,
		EventDescription:    eventDescription,
		EventDetails:        eventDetails,
		SpecialNote:         specialNote,
	}

	if orderID != "" {
		event.OrderID = &orderID
	}
	if draftOrderID != "" {
		event.DraftOrderID = &draftOrderID
	}
	if productID != "" {
		event.ProductID = &productID
	}
	if listID != "" {
		event.ListID = &listID
	}
	if cartID != "" {
		event.CartID = &cartID
	}
	if discountID != "" {
		event.DiscountID = &discountID
	}
	if giftCardID != "" {
		event.GiftCardID = &giftCardID
	}

	hasErr := false
	errList := []string{}

	for _, e := range errors {
		if e != nil {
			errList = append(errList, e.Error())
			hasErr = true
		}
	}

	event.AnyError = hasErr
	event.AllErrorsSt = errList

	r.events = append(r.events, &event)

}

func (r *eventRepo) FlushBatch() {
	r.mutex.Lock()
	if len(r.events) > 0 {
		docs := []interface{}{}
		for _, v := range r.events {
			docs = append(docs, v)
		}
		if res, err := r.coll.InsertMany(context.Background(), docs); err != nil {
			log.Printf("Unable to insert events for store: %s, error: %v\n", r.store, err)
		} else {
			log.Printf("Successfully inserted %d events\n", len(res.InsertedIDs))
			r.events = nil
		}
	}
	r.mutex.Unlock()
}
