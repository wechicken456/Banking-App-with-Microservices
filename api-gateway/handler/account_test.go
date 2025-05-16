package handler

import (
	"api-gateway/proto"
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func loadDotEnvForTest() {
	fmt.Println("Loading .env file for test")
	dotEnvFilename = "../.env"
	if err := godotenv.Load(dotEnvFilename); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

func createClient() (proto.AccountServiceClient, error) {
	loadDotEnvForTest()

	accountClient := NewAccountServiceClient()
	if accountClient == nil {
		panic("Failed to create account service client")
	}
	return accountClient, nil
}

func TestCreateAcocunt_Success(t *testing.T) {
	accountClient, err := createClient()
	if err != nil {
		t.Fatalf("Failed to create account service client: %v", err)
	}

	id := uuid.New()
	idBytes, _ := id.MarshalBinary()

	args := &proto.CreateAccountRequest{
		UserId:  idBytes,
		Balance: 100,
	}

	res, err := accountClient.CreateAccount(context.Background(), args)
	if err != nil {
		log.Printf("Failed to create account: %v", err)
	}
	log.Printf("Account created: %v", res)
}
