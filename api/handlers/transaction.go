package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/context"
)

type TransactionHandler struct {
	db          *sql.DB
	mongoClient *mongo.Client
	rabbitMQ    *amqp.Connection
}

func NewTransactionHandler(db *sql.DB, mongoClient *mongo.Client, rabbitMQ *amqp.Connection) *TransactionHandler {
	return &TransactionHandler{db: db, mongoClient: mongoClient, rabbitMQ: rabbitMQ}
}

// ProcessTransaction handles deposit/withdraw requests
func (h *TransactionHandler) ProcessTransaction(c *gin.Context) {
	var req struct {
		AccountID int     `json:"account_id" binding:"required"`
		Type      string  `json:"type" binding:"required,oneof=deposit withdraw"`
		Amount    float64 `json:"amount" binding:"required,gt=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Publish transaction to RabbitMQ
	ch, err := h.rabbitMQ.Channel()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create channel"})
		return
	}
	defer ch.Close()

	q, err := ch.QueueDeclare("transactions", true, false, false, false, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to declare queue"})
		return
	}

	body, _ := json.Marshal(req)
	err = ch.Publish("", q.Name, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish transaction"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "Transaction queued"})
}

// GetTransactionHistory retrieves transactions for an account
func (h *TransactionHandler) GetTransactionHistory(c *gin.Context) {
	accountID := c.Param("account_id")
	collection := h.mongoClient.Database("banking").Collection("transactions")

	filter := bson.M{"account_id": accountID}
	cursor, err := collection.Find(context.TODO(), filter, options.Find())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transaction history"})
		return
	}
	defer cursor.Close(context.TODO())

	var transactions []bson.M
	if err := cursor.All(context.TODO(), &transactions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode transactions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"transactions": transactions})
}
