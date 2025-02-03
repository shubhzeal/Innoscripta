package processor

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"

	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Transaction struct {
	AccountID int     `json:"account_id"`
	Type      string  `json:"type"`
	Amount    float64 `json:"amount"`
}

func StartTransactionProcessor(db *sql.DB, mongoClient *mongo.Client, rabbitMQ *amqp.Connection) {
	ch, err := rabbitMQ.Channel()
	if err != nil {
		log.Fatalf("Failed to create channel: %v", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare("transactions", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to register consumer: %v", err)
	}

	collection := mongoClient.Database("banking").Collection("transactions")

	log.Println("Transaction processor started. Waiting for messages...")
	for d := range msgs {
		var t Transaction
		err := json.Unmarshal(d.Body, &t)
		if err != nil {
			log.Printf("Failed to parse transaction: %v", err)
			continue
		}

		// Process the transaction
		tx, err := db.Begin()
		if err != nil {
			log.Printf("Failed to start transaction: %v", err)
			continue
		}

		var balance float64
		err = tx.QueryRow("SELECT balance FROM accounts WHERE id = $1 FOR UPDATE", t.AccountID).Scan(&balance)
		if err != nil {
			log.Printf("Account not found: %v", err)
			tx.Rollback()
			continue
		}

		if t.Type == "withdraw" && balance < t.Amount {
			log.Printf("Insufficient balance for account %d", t.AccountID)
			tx.Rollback()
			continue
		}

		newBalance := balance
		if t.Type == "deposit" {
			newBalance += t.Amount
		} else {
			newBalance -= t.Amount
		}

		_, err = tx.Exec("UPDATE accounts SET balance = $1 WHERE id = $2", newBalance, t.AccountID)
		if err != nil {
			log.Printf("Failed to update balance: %v", err)
			tx.Rollback()
			continue
		}

		_, err = collection.InsertOne(context.TODO(), bson.M{
			"account_id": t.AccountID,
			"type":       t.Type,
			"amount":     t.Amount,
			"balance":    newBalance,
		})
		if err != nil {
			log.Printf("Failed to log transaction: %v", err)
			tx.Rollback()
			continue
		}

		err = tx.Commit()
		if err != nil {
			log.Printf("Failed to commit transaction: %v", err)
			continue
		}

		log.Printf("Processed transaction for account %d: %s %.2f", t.AccountID, t.Type, t.Amount)
	}
}
