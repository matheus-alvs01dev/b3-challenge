package main

import (
	"b3challenge/config"
	"b3challenge/internal/adapter/db"
	"b3challenge/internal/api"
	"b3challenge/internal/di"
	"log"
)

func main() {
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	db, err := db.NewClient(config.GetDatabaseDSN())
	if err != nil {
		log.Fatalf("Error initializing database client: %v", err)
	}

	diContainer := di.NewContainer(db.DB())

	server := api.NewServer()
	server.ConfigureRoutes(
		diContainer.NewTradesHandler(),
	)
	server.Start(config.GetAPIPort())
}
