package repositories

import (
	"beam/config"
	"beam/data/models"
	"context"
	"log"
	"os"
	"os/signal"
	"slices"
	"sync"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
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
		guestID, eventClassification, eventDescription, eventDetails, specialNote, orderID, draftOrderID, productID, variantID, favesID, savesID, lolistID, cartID, cartLineID, discountID, giftCardID string,
		errors []error,
	)
	AddToBatchNew(eventClassification, eventDescription, eventDetails, specialNote string, ids models.EventIDPassIn, errors []error)
	FlushBatch()
}

type eventRepo struct {
	coll       *mongo.Collection
	client     *redis.Client
	mutex      sync.Mutex
	events     []*models.Event
	eventsNew  []*models.EventNew
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

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		defer repo.saveTicker.Stop()
		defer repo.FlushBatch()

		for {
			select {
			case <-repo.saveTicker.C:
				repo.FlushBatch()
			case <-sigChan:
				return
			}
		}
	}()

	return repo
}

func (r *eventRepo) AddToBatch(
	customerID int,
	guestID, eventClassification, eventDescription, eventDetails, specialNote, orderID, draftOrderID, productID, variantID, favesID, savesID, lolistID, cartID, cartLineID, discountID, giftCardID string,
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
	if variantID != "" {
		event.VariantID = &variantID
	}
	if savesID != "" {
		event.SavesID = &savesID
	}
	if favesID != "" {
		event.FavesID = &favesID
	}
	if lolistID != "" {
		event.LOListID = &lolistID
	}
	if cartID != "" {
		event.CartID = &cartID
	}
	if cartLineID != "" {
		event.CartLineID = &cartLineID
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

func (r *eventRepo) AddToBatchNew(eventClassification, eventDescription, eventDetails, specialNote string, ids models.EventIDPassIn, errors []error) {
	if !slices.Contains(validClassifications, eventClassification) {
		log.Printf("invalid event classification: %s\n", eventClassification)
		eventClassification = "Other"
	}

	event := models.EventNew{
		Timestamp:           time.Now(),
		EventClassification: eventClassification,
		EventDescription:    eventDescription,
		EventDetails:        eventDetails,
		SpecialNote:         specialNote,
		CustomerID:          ids.CustomerID,
		GuestID:             ids.GuestID,
		OrderID:             ids.OrderID,
		DraftOrderID:        ids.DraftOrderID,
		ProductID:           ids.ProductID,
		ProductHandle:       ids.ProductHandle,
		VariantID:           ids.VariantID,
		SavesID:             ids.SavesID,
		FavesID:             ids.FavesID,
		LastOrderListID:     ids.LastOrderListID,
		CartID:              ids.CartID,
		CartLineID:          ids.CartLineID,
		DiscountID:          ids.DiscountID,
		DiscountCode:        ids.DiscountCode,
		GiftCardID:          ids.GiftCardID,
		GiftCardCode:        ids.GiftCardCode,
		SessionID:           ids.SessionID,
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

	r.eventsNew = append(r.eventsNew, &event)
}

func (r *eventRepo) FlushBatch() {
	r.mutex.Lock()
	if len(r.events) > 0 {
		docs := []interface{}{}
		for _, v := range r.eventsNew {
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
