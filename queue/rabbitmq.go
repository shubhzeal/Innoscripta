package queue

import (
	"log"

	"github.com/streadway/amqp"
)

func InitRabbitMQ(url string) *amqp.Connection {
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	log.Println("Connected to RabbitMQ.")
	return conn
}
