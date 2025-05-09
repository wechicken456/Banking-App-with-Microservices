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
		os.Getenv("AUTH_DB_HOST"),
		os.Getenv("AUTH_DB_PORT"),
		os.Getenv("AUTH_DB_USER"),
		os.Getenv("AUTH_DB_PASSWORD"),
		os.Getenv("AUTH_DB_NAME"),
	)
	db, err := sqlx.Connect("postgres", dbConnectString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	return db
}
