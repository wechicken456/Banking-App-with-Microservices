package initialize

import (
	"log"

	"github.com/joho/godotenv"
)

func LoadDotEnv(dotEnvFilename string) {
	err := godotenv.Load(dotEnvFilename)
	if err != nil {
		log.Fatalf("Failed to load .env file: %v", err)
	}
}
