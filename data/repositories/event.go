package repositories

import (
	"beam/data/models"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type EventRepository interface {
	Create(event models.Event) error
	Read(id string) (*models.Event, error)
	Update(event models.Event) error
	Delete(id string) error
}

type eventRepo struct {
	coll *mongo.Collection
}

func NewEventRepository(mdb *mongo.Database) EventRepository {
	collection := mdb.Collection("Event")
	return &eventRepo{coll: collection}
}

func (r *eventRepo) Create(event models.Event) error {
	_, err := r.coll.InsertOne(context.Background(), event)
	return err
}

func (r *eventRepo) Read(id string) (*models.Event, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var event models.Event
	err = r.coll.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&event)
	return &event, err
}

func (r *eventRepo) Update(event models.Event) error {
	_, err := r.coll.UpdateOne(
		context.Background(),
		bson.M{"_id": event.ID},
		bson.M{"$set": event},
	)
	return err
}

func (r *eventRepo) Delete(id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.coll.DeleteOne(context.Background(), bson.M{"_id": objID})
	return err
}
