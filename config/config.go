package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	PostgresDSN string
	MongoURI    string
	RabbitMQURL string
}

func LoadConfig() Config {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found. Using system environment variables.")
	}

	return Config{
		PostgresDSN: os.Getenv("POSTGRES_DSN"),
		MongoURI:    os.Getenv("MONGO_URI"),
		RabbitMQURL: os.Getenv("RABBITMQ_URL"),
	}
}
