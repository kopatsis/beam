package repositories

import (
	"beam/config"
	"beam/data/models"
	"context"
	"encoding/json"
	"log"
	"slices"
	"strconv"
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
	rdb        *redis.Client
	key        string
	saveTicker *time.Ticker
	store      string
}

func NewEventRepository(mdb *mongo.Database, store string) EventRepository {
	repo := &eventRepo{
		coll:       mdb.Collection("Event"),
		saveTicker: time.NewTicker(time.Duration(config.BATCH) * time.Second),
		store:      store,
	}

	go func() {
		defer repo.saveTicker.Stop()

		for range repo.saveTicker.C {
			repo.FlushBatch()
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

	data, err := json.Marshal(event)
	if err == nil {
		if err := r.rdb.LPush(context.Background(), r.store+"::EVE::"+r.key, data); err != nil {
			log.Printf("Unable to push event to redis for store: %s; err: %v\n", r.store, err)
		}
	} else {
		log.Printf("Unable to create event for redis for store: %s; err: %v\n", r.store, err)
	}
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

	data, err := json.Marshal(event)
	if err == nil {
		if err := r.rdb.LPush(context.Background(), r.store+"::EVE::"+r.key, data); err != nil {
			log.Printf("Unable to push event to redis for store: %s; err: %v\n", r.store, err)
		}
	} else {
		log.Printf("Unable to create event for redis for store: %s; err: %v\n", r.store, err)
	}
}

func (r *eventRepo) FlushBatch() {
	oldKey := r.key
	r.key = strconv.FormatInt(time.Now().Unix(), 10)

	ctx := context.Background()

	eventsData, err := r.rdb.LRange(ctx, r.store+"::EVE::"+oldKey, 0, -1).Result()
	if err != nil {
		log.Printf("Error retrieving events for store: %s; key: %s; error: %v", r.store, oldKey, err)
	} else {
		if err := r.rdb.Del(ctx, r.store+"::EVE::"+oldKey).Err(); err != nil {
			log.Printf("Error deleting events for store: %s; key: %s; error: %v", r.store, oldKey, err)
		}
	}

	var eventsToSave []models.Event

	if len(eventsData) == 0 {
		return
	}

	for _, event := range eventsData {
		var e models.Event
		if err := json.Unmarshal([]byte(event), &e); err != nil {
			log.Printf("Error unmarshalling event for store: %s; key: %s; error: %v", r.store, oldKey, err)
		} else {
			eventsToSave = append(eventsToSave, e)
		}
	}

	if len(eventsToSave) == 0 {
		return
	}

	var docs []interface{}
	for _, e := range eventsToSave {
		docs = append(docs, e)
	}

	if _, err := r.coll.InsertMany(ctx, docs); err != nil {
		log.Printf("Unable to save events for store: %s; key: %s; error: %v", r.store, oldKey, err)
		if len(eventsToSave) > 0 {
			if err := r.rdb.LPush(ctx, r.store+"::EVE::"+r.key, eventsData).Err(); err != nil {
				log.Printf("Error pushing back events for store: %s; key: %s; error: %v", r.store, r.key, err)
			}
		}
	}
}
