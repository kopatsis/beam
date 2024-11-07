package main

import (
	"beam/config"
	"beam/data"
	"beam/routing"
	"log"
	"net/http"
)

func main() {

	config.EnvVariables()

	pgDB, redis := config.PostgresConnect(), config.NewRedisClient()
	mongoClient, mongoDB := config.MongoConnect()
	defer config.MongoDisconnect(mongoClient)

	fullService := data.NewMainService(pgDB, redis, mongoDB)
	mutexes := config.LoadAllData()

	rtr := routing.New(fullService, mutexes)

	port := config.GetPort()

	if err := http.ListenAndServe(":"+port, rtr); err != nil {
		log.Fatalf("There was an error with the http server: %v", err)
	}
}
