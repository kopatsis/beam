package main

import (
	"log"
	"net/http"
	"beam/config"
	"beam/data"
	"beam/routing"
)

func main() {

	config.EnvVariables()

	mysql, redis := config.NewMySQLConnection(), config.NewRedisClient()

	fullService := data.NewMainService(mysql, redis)

	rtr := routing.New(fullService)

	port := config.GetPort()

	if err := http.ListenAndServe(":"+port, rtr); err != nil {
		log.Fatalf("There was an error with the http server: %v", err)
	}
}
