package api

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupRouter(db *sql.DB, mongoClient *mongo.Client, rabbitMQ *amqp.Connection) *gin.Engine {
	router := gin.Default()

	// Initialize handlers
	accountHandler := NewAccountHandler(db)
	transactionHandler := NewTransactionHandler(db, mongoClient, rabbitMQ)

	// Routes
	router.POST("/accounts", accountHandler.CreateAccount)
	router.POST("/transactions", transactionHandler.ProcessTransaction)
	router.GET("/transactions/:account_id", transactionHandler.GetTransactionHistory)

	return router
}
