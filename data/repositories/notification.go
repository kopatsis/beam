package repositories

import (
	"beam/data/models"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type NotificationRepository interface {
	Create(notification models.Notification) error
	Read(id string) (*models.Notification, error)
	Update(notification models.Notification) error
	Delete(id string) error
}

type notificationRepo struct {
	coll *mongo.Collection
}

func NewNotificationRepository(mdb *mongo.Database) NotificationRepository {
	collection := mdb.Collection("Notification")
	return &notificationRepo{coll: collection}
}

func (r *notificationRepo) Create(notification models.Notification) error {
	_, err := r.coll.InsertOne(context.Background(), notification)
	return err
}

func (r *notificationRepo) Read(id string) (*models.Notification, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var notification models.Notification
	err = r.coll.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&notification)
	return &notification, err
}

func (r *notificationRepo) Update(notification models.Notification) error {
	_, err := r.coll.UpdateOne(
		context.Background(),
		bson.M{"_id": notification.ID},
		bson.M{"$set": notification},
	)
	return err
}

func (r *notificationRepo) Delete(id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.coll.DeleteOne(context.Background(), bson.M{"_id": objID})
	return err
}
