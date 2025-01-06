package repositories

import (
	"beam/data/models"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrderRepository interface {
	CreateDraft(order *models.DraftOrder) error
	ReadDraft(id string) (*models.DraftOrder, error)
	UpdateDraft(order *models.DraftOrder) error
	DeleteDraft(id string) error
}

type orderRepo struct {
	coll *mongo.Collection
}

func NewOrderRepository(mdb *mongo.Database) OrderRepository {
	collection := mdb.Collection("Draft_Order")
	return &orderRepo{coll: collection}
}

func (r *orderRepo) CreateDraft(order *models.DraftOrder) error {
	_, err := r.coll.InsertOne(context.Background(), order)
	return err
}

func (r *orderRepo) ReadDraft(id string) (*models.DraftOrder, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var order models.DraftOrder
	err = r.coll.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&order)
	return &order, err
}

func (r *orderRepo) UpdateDraft(order *models.DraftOrder) error {
	_, err := r.coll.UpdateOne(
		context.Background(),
		bson.M{"_id": order.ID},
		bson.M{"$set": order},
	)
	return err
}

func (r *orderRepo) DeleteDraft(id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.coll.DeleteOne(context.Background(), bson.M{"_id": objID})
	return err
}
