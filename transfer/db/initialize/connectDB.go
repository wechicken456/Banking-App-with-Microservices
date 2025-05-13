package initialize

import (
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func ConnectDB() *sqlx.DB {
	// connect to the database
	var dbConnectString = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("TRANSFER_DB_HOST"),
		os.Getenv("TRANSFER_DB_PORT"),
		os.Getenv("TRANSFER_DB_USER"),
		os.Getenv("TRANSFER_DB_PASSWORD"),
		os.Getenv("TRANSFER_DB_NAME"),
	)
	db, err := sqlx.Connect("postgres", dbConnectString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	return db
}
