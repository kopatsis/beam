package config

import (
	"context"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func MongoDisconnect(client *mongo.Client) {
	err := client.Disconnect(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}

func MongoConnect() (*mongo.Client, *mongo.Database) {

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	connectStr := os.Getenv("MONGOSTRING")
	clientOptions := options.Client().ApplyURI(connectStr).SetServerAPIOptions(serverAPI)

	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatalf("failed to establish mongo client: %v", err)
	}

	// Check the connection
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatalf("failed to connect to mongo database: %v", err)
	}

	db := os.Getenv("MONGO_DBNAME")
	if db == "" {
		db = "test"
	}

	// Specify the database and collection
	database := client.Database(db)

	return client, database
}
