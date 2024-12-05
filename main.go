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
	mutexes := config.LoadAllData()

	pgDBs, redis := config.PostgresConnect(mutexes), config.NewRedisClient()
	mongoClient, mongoDBs := config.MongoConnect(mutexes)
	defer config.MongoDisconnect(mongoClient)

	fullService := data.NewMainService(pgDBs, redis, mongoDBs, mutexes)
	tools := config.NewTools()

	rtr := routing.New(fullService, tools)

	port := config.GetPort()

	if err := http.ListenAndServe(":"+port, rtr); err != nil {
		log.Fatalf("There was an error with the http server: %v", err)
	}
}
