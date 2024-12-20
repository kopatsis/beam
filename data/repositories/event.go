package repositories

import (
	"beam/data/models"
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

var (
	eventQueue           chan models.Event
	once                 sync.Once
	validClassifications = []string{
		"Order", "DraftOrder", "Product", "List", "Cart", "Collection", "Discount", "GiftCard",
	}
)

type EventRepository interface {
	SaveEvent(
		customerID int,
		guestID, eventClassification, eventDescription, specialNote, orderID, draftOrderID, productID, listID, cartID, discountID, giftCardID string,
	) error
}

type eventRepo struct {
	coll *mongo.Collection
}

func NewEventRepository(mdb *mongo.Database) EventRepository {
	collection := mdb.Collection("Event")
	return &eventRepo{coll: collection}
}

func (repo *eventRepo) SaveEvent(
	customerID int,
	guestID, eventClassification, eventDescription, specialNote, orderID, draftOrderID, productID, listID, cartID, discountID, giftCardID string,
) error {
	if !slices.Contains(validClassifications, eventClassification) {
		return fmt.Errorf("invalid event classification: %s", eventClassification)
	}

	event := models.Event{
		CustomerID:          customerID,
		GuestID:             guestID,
		Timestamp:           time.Now(),
		EventClassification: eventClassification,
		EventDescription:    eventDescription,
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

	once.Do(func() {
		eventQueue = make(chan models.Event, 100)
		for i := 0; i < 5; i++ {
			go func() {
				for task := range eventQueue {
					_, err := repo.coll.InsertOne(context.Background(), task)
					if err != nil {
						fmt.Printf("ERROR saving event: %e\n", err)
					}
				}
			}()
		}
	})

	go func() {
		select {
		case eventQueue <- event:
		default:
		}
	}()

	return nil
}
