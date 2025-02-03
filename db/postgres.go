package db

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func InitPostgres(dsn string) *sql.DB {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	// Verify connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("PostgreSQL connection error: %v", err)
	}

	log.Println("Connected to PostgreSQL.")
	return db
}
