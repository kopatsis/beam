package repositories

import (
	"beam/data/models"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type SessionRepository interface {
	Create(session models.Session) error
	Read(id string) (*models.Session, error)
	Update(session models.Session) error
	Delete(id string) error
}

type sessionRepo struct {
	coll *mongo.Collection
}

func NewSessionRepository(mdb *mongo.Database) SessionRepository {
	collection := mdb.Collection("Session")
	return &sessionRepo{coll: collection}
}

func (r *sessionRepo) Create(session models.Session) error {
	_, err := r.coll.InsertOne(context.Background(), session)
	return err
}

func (r *sessionRepo) Read(id string) (*models.Session, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var session models.Session
	err = r.coll.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&session)
	return &session, err
}

func (r *sessionRepo) Update(session models.Session) error {
	_, err := r.coll.UpdateOne(
		context.Background(),
		bson.M{"_id": session.ID},
		bson.M{"$set": session},
	)
	return err
}

func (r *sessionRepo) Delete(id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.coll.DeleteOne(context.Background(), bson.M{"_id": objID})
	return err
}
