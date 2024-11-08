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

func MongoConnect(mutex *AllMutexes) (*mongo.Client, map[string]*mongo.Database) {

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	connectStr := os.Getenv("MONGOSTRING")
	clientOptions := options.Client().ApplyURI(connectStr).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatalf("failed to establish mongo client: %v", err)
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatalf("failed to connect to mongo database: %v", err)
	}

	ret := map[string]*mongo.Database{}

	mutex.Store.Mu.RLock()

	for dbName := range mutex.Store.Store.ToDomain {
		database := client.Database(dbName)
		ret[dbName] = database
	}

	return client, ret
}
