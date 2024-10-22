package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gigimon/solstat/pkg/config"
	"github.com/gigimon/solstat/pkg/database"
	"github.com/gigimon/solstat/pkg/web"
)

func main() {
	cfg := config.GetServerConfig()
	log.Println("Connect to MongoDB: ", cfg.Database.MongoUri)
	mongoClient := database.InitializeDatabase(cfg.Database.MongoUri, cfg.Database.DBName)
	defer func() {
		if err := mongoClient.Client().Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	app := &web.App{
		DB: mongoClient,
	}

	http.HandleFunc("/api/fee/blocks", app.GetBlocks)

	log.Print("Starting server on port 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
