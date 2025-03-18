package repositories

import (
	"beam/data/models"
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type OrderRepository interface {
	CreateOrder(order *models.Order) error
	CreateBlankOrder() (string, error)
	Update(order *models.Order) error
	Read(id string) (*models.Order, error)
	ReadStatus(id string) (string, error)
	GetOrders(customerID, limit, offset int, sortColumn string, desc bool) ([]*models.Order, error)

	PaymentListen(orderID, store string, cancelOut time.Duration) (string, error)
	PaymentPublish(orderID, store, message string) error
	MarkOrderStatusUpdate(order *models.Order, status string) (bool, error)

	GetCheckOrders() ([]models.Order, error)
	UpdateCheckDeliveryDate(ids []string) error
	UpdateCheckEmailSent(ids []string) error
	GetOrdersByIDs(ids []string) ([]models.Order, error)

	GetOrdersByEmail(email string) (bool, error)
	GetOrdersByEmailAndCustomer(email string, custID int) (bool, error)
}

type orderRepo struct {
	coll *mongo.Collection
	rdb  *redis.Client
}

func NewOrderRepository(mdb *mongo.Database, rdb *redis.Client) OrderRepository {
	collection := mdb.Collection("Order")
	return &orderRepo{coll: collection, rdb: rdb}
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

func (r *orderRepo) CreateBlankOrder() (string, error) {
	order := &models.Order{Status: "Blank"}

	ctx := context.Background()
	res, err := r.coll.InsertOne(ctx, order)
	if err != nil {
		return "", err
	}

	return res.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (r *orderRepo) Update(order *models.Order) error {
	_, err := r.coll.UpdateOne(
		context.Background(),
		bson.M{"_id": order.ID},
		bson.M{"$set": order},
	)
	return err
}

func (r *orderRepo) Read(id string) (*models.Order, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var order models.Order
	err = r.coll.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&order)
	return &order, err
}

func (r *orderRepo) ReadStatus(id string) (string, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return "", err
	}
	var result struct {
		Status string `bson:"status"`
	}
	err = r.coll.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&result)
	return result.Status, err
}

func (r *orderRepo) PaymentListen(orderID, store string, cancelOut time.Duration) (string, error) {
	if cancelOut < 10*time.Millisecond {
		return "", nil
	}

	streams, err := r.rdb.XRead(context.Background(), &redis.XReadArgs{
		Streams: []string{store + "::PMLN::" + orderID, "0"},
		Count:   1,
		Block:   cancelOut,
	}).Result()

	if err != nil {
		return "", err
	}

	for _, stream := range streams {
		for _, msg := range stream.Messages {
			return msg.Values["message"].(string), nil
		}
	}

	return "", nil
}

func (r *orderRepo) PaymentPublish(orderID, store, message string) error {
	stream := store + "::PMLN::" + orderID
	_, err := r.rdb.XAdd(context.Background(), &redis.XAddArgs{
		Stream: stream,
		Values: map[string]interface{}{"message": message},
	}).Result()
	return err
}

func (r *orderRepo) MarkOrderStatusUpdate(order *models.Order, status string) (bool, error) {
	if order.Status == "Blank" {
		start := time.Now()
		for {
			if time.Since(start) < 1*time.Second {
				time.Sleep(250 * time.Millisecond)
			} else if time.Since(start) < 3*time.Second {
				time.Sleep(500 * time.Millisecond)
			} else {
				time.Sleep(1000 * time.Millisecond)
			}

			status, err := r.ReadStatus(order.ID.Hex())
			if err != nil {
				return false, err
			}
			if status != "Blank" {
				break
			}
			if time.Since(start) >= 10*time.Second {
				return true, nil
			}
		}
	}

	_, err := r.coll.UpdateOne(
		context.Background(),
		bson.M{"_id": order.ID},
		bson.M{"$set": bson.M{"status": status}},
	)
	return false, err
}

// sortColumn in "date_created", "subtotal", "total"; defaults to "date_created"
func (r *orderRepo) GetOrders(customerID, limit, offset int, sortColumn string, desc bool) ([]*models.Order, error) {
	filter := bson.M{"customer_id": customerID}

	sortOrder := 1
	if desc {
		sortOrder = -1
	}

	if sortColumn != "total" && sortColumn != "subtotal" {
		sortColumn = "date_created"
	}

	findOptions := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.M{"date_created": sortOrder})

	cursor, err := r.coll.Find(context.Background(), filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var orders []*models.Order
	if err := cursor.All(context.Background(), &orders); err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *orderRepo) GetCheckOrders() ([]models.Order, error) {
	filter := bson.M{
		"check_sent": false,
		"status":     "Shipped",
		"check_date": bson.M{"$lt": time.Now()},
	}

	findOptions := options.Find()

	cursor, err := r.coll.Find(context.Background(), filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var orders []models.Order
	if err := cursor.All(context.Background(), &orders); err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *orderRepo) UpdateCheckDeliveryDate(ids []string) error {
	var objectIDs []primitive.ObjectID
	for _, id := range ids {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return err
		}
		objectIDs = append(objectIDs, objID)
	}

	filter := bson.M{"_id": bson.M{"$in": objectIDs}}
	update := bson.M{"$set": bson.M{"check_date": time.Now().Add(72 * time.Hour)}}

	_, err := r.coll.UpdateMany(context.Background(), filter, update)
	return err
}

func (r *orderRepo) UpdateCheckEmailSent(ids []string) error {
	var objectIDs []primitive.ObjectID
	for _, id := range ids {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return err
		}
		objectIDs = append(objectIDs, objID)
	}

	filter := bson.M{"_id": bson.M{"$in": objectIDs}}
	update := bson.M{"$set": bson.M{"check_sent": true, "status": "Delivered"}}

	_, err := r.coll.UpdateMany(context.Background(), filter, update)
	return err
}

func (r *orderRepo) GetOrdersByIDs(ids []string) ([]models.Order, error) {
	var objectIDs []primitive.ObjectID
	for _, id := range ids {
		objectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return nil, err
		}
		objectIDs = append(objectIDs, objectID)
	}

	filter := bson.M{
		"_id": bson.M{"$in": objectIDs},
	}

	findOptions := options.Find()

	cursor, err := r.coll.Find(context.Background(), filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var orders []models.Order
	if err := cursor.All(context.Background(), &orders); err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *orderRepo) GetOrdersByEmail(email string) (bool, error) {
	filter := bson.M{"status": bson.M{"$ne": "Cancelled"}, "email": email, "guest": true} // Update statuses

	var result struct {
		ID primitive.ObjectID `bson:"_id"`
	}
	err := r.coll.FindOne(context.Background(), filter, options.FindOne().SetProjection(bson.M{"_id": 1})).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return false, nil
	}
	return err == nil, err
}

func (r *orderRepo) GetOrdersByEmailAndCustomer(email string, custID int) (bool, error) {
	filter := bson.M{
		"$or": bson.A{
			bson.M{"status": bson.M{"$ne": "Cancelled"}, "email": email, "guest": true},         // Update statuses
			bson.M{"status": bson.M{"$ne": "Cancelled"}, "customer_id": custID, "guest": false}, // Update statuses
		},
	}

	var result struct {
		ID primitive.ObjectID `bson:"_id"`
	}
	err := r.coll.FindOne(context.Background(), filter, options.FindOne().SetProjection(bson.M{"_id": 1})).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return false, nil
	}
	return err == nil, err
}
