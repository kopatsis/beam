package repositories

import (
	"beam/data/models"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type DraftOrderRepository interface {
	Create(draftOrder *models.DraftOrder) error
	Read(id string) (*models.DraftOrder, error)
	Update(draftOrder *models.DraftOrder) error
	Delete(id string) error
}

type draftOrderRepo struct {
	coll *mongo.Collection
}

func NewDraftOrderRepository(mdb *mongo.Database) DraftOrderRepository {
	collection := mdb.Collection("DraftOrder")
	return &draftOrderRepo{coll: collection}
}

func (r *draftOrderRepo) Create(draftOrder *models.DraftOrder) error {
	ctx := context.Background()
	res, err := r.coll.InsertOne(ctx, draftOrder)
	if err != nil {
		return err
	}

	draftOrder.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *draftOrderRepo) Read(id string) (*models.DraftOrder, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var draftOrder models.DraftOrder
	err = r.coll.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&draftOrder)
	return &draftOrder, err
}

func (r *draftOrderRepo) Update(draftOrder *models.DraftOrder) error {
	_, err := r.coll.UpdateOne(
		context.Background(),
		bson.M{"_id": draftOrder.ID},
		bson.M{"$set": draftOrder},
	)
	return err
}

func (r *draftOrderRepo) Delete(id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.coll.DeleteOne(context.Background(), bson.M{"_id": objID})
	return err
}
