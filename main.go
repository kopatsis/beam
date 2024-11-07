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

	mysql, redis := config.NewMySQLConnection(), config.NewRedisClient()

	fullService := data.NewMainService(mysql, redis)
	mutexes := config.LoadAllData()

	rtr := routing.New(fullService, mutexes)

	port := config.GetPort()

	if err := http.ListenAndServe(":"+port, rtr); err != nil {
		log.Fatalf("There was an error with the http server: %v", err)
	}
}
