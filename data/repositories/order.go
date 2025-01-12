package repositories

import (
	"beam/data/models"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrderRepository interface {
	CreateOrder(order *models.Order) error
	Update(order *models.Order) error
}

type orderRepo struct {
	coll *mongo.Collection
}

func NewOrderRepository(mdb *mongo.Database) OrderRepository {
	collection := mdb.Collection("Order")
	return &orderRepo{coll: collection}
}

func (r *orderRepo) CreateOrder(order *models.Order) error {
	ctx := context.Background()
	res, err := r.coll.InsertOne(ctx, order)
	if err != nil {
		return err
	}

	order.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *orderRepo) Update(order *models.Order) error {
	_, err := r.coll.UpdateOne(
		context.Background(),
		bson.M{"_id": order.ID},
		bson.M{"$set": order},
	)
	return err
}
