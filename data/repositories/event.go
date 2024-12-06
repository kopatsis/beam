package repositories

import (
	"beam/data/models"
	"context"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

var (
	eventQueue chan models.Event
	once       sync.Once
)

type EventRepository interface {
	SaveEvent(customerID int, guestID, eventClassification, eventDescription, specialNote, otherID string) error
}

type eventRepo struct {
	coll *mongo.Collection
}

func NewEventRepository(mdb *mongo.Database) EventRepository {
	collection := mdb.Collection("Event")
	return &eventRepo{coll: collection}
}

func (repo *eventRepo) SaveEvent(customerID int, guestID, eventClassification, eventDescription, specialNote, otherID string) error {
	event := models.Event{
		CustomerID:          customerID,
		GuestID:             guestID,
		Timestamp:           time.Now(),
		EventClassification: eventClassification,
		EventDescription:    eventDescription,
		SpecialNote:         specialNote,
	}

	switch eventClassification {
	case "Order":
		event.OrderID = &otherID
	case "Product":
		event.ProductID = &otherID
	case "List":
		event.ListID = &otherID
	case "Cart":
		event.CartID = &otherID
	case "Collection":
		event.CollectionID = &otherID
	case "Discount":
		event.DiscountID = &otherID
	}

	once.Do(func() {
		eventQueue = make(chan models.Event, 100)
		for i := 0; i < 5; i++ {
			go func() {
				for task := range eventQueue {
					_, err := repo.coll.InsertOne(context.Background(), task)
					if err != nil {
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
