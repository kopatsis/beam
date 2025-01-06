package repositories

import (
	"go.mongodb.org/mongo-driver/mongo"
)

type OrderRepository interface {
}

type orderRepo struct {
	coll *mongo.Collection
}

func NewOrderRepository(mdb *mongo.Database) OrderRepository {
	collection := mdb.Collection("Order")
	return &orderRepo{coll: collection}
}
