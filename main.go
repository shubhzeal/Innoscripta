package main

import (
	"banking-ledger/api"
	"banking-ledger/config"
	"banking-ledger/db"
	"banking-ledger/queue"
	"log"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize PostgreSQL
	dbConn := db.InitPostgres(cfg.PostgresDSN)
	defer dbConn.Close()

	// Initialize MongoDB
	mongoClient := db.InitMongo(cfg.MongoURI)
	defer mongoClient.Disconnect(nil)

	// Initialize RabbitMQ
	rabbitMQConn := queue.InitRabbitMQ(cfg.RabbitMQURL)
	defer rabbitMQConn.Close()

	// Start API server
	router := api.SetupRouter(dbConn, mongoClient, rabbitMQConn)
	log.Println("Starting server on port 8080...")
	log.Fatal(router.Run(":8080"))
}
