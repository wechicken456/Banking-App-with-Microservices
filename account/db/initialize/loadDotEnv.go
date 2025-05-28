package initialize

import (
	"log"

	"github.com/joho/godotenv"
)

// meant to be called from the root of `account`, since it loads .env from CWD
func LoadDotEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Failed to load .env file: %v", err)
	}
}
