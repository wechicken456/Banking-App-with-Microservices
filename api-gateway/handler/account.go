package handler

import (
	"api-gateway/client"
	"api-gateway/initialize"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var dotEnvFilename = ".env"

func NewAccountServiceClient() client.AccountServiceClient {
	initialize.LoadDotEnv(dotEnvFilename)

	// Connect to the account service
	connString := fmt.Sprintf("%s:%s", os.Getenv("ACCOUNT_SERVICE_HOST"), os.Getenv("ACCOUNT_SERVICE_PORT"))
	fmt.Printf("Connecting to account service at %s\n", connString)
	conn, err := grpc.NewClient(connString, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Failed to connect to account service: %v", err)
		return nil
	}
	client := client.NewAccountServiceClient(conn)
	return client
}

func CreateAccount(userID string, balance int64) {
	client := NewAccountServiceClient()
	if client == nil {
		log.Fatalf("Failed to create account service client")
	}

	userIDBytes, err := uuid.Parse(userID)
	if err != nil {
		log.Printf("Failed to parse user ID: %v", err)
		return
	}

	req := &client.CreateAccountRequest{
		UserId:  userIDBytes[:],
		Balance: balance,
	}

	res, err := client.CreateAccount(context.Background(), req)
	if err != nil {
		log.Printf("Failed to create account: %v", err)
		return
	}
	log.Printf("Account created: %v", res)
}

func GetAccountsByUserID(userID string) []*client.Account {
	client := NewAccountServiceClient()
	if client == nil {
		log.Fatalf("Failed to create account service client")
	}

	userIDBytes, err := uuid.Parse(userID)
	if err != nil {
		log.Printf("Failed to parse user ID: %v", err)
		return nil
	}

	req := &client.GetAccountsByUserIdRequest{
		UserId: userIDBytes[:],
	}

	res, err := client.GetAccountsByUserId(context.Background(), req)
	if err != nil {
		log.Printf("Failed to get accounts by user ID: %v", err)
		return nil
	}
	log.Printf("Accounts: %v", res)
	return res.Accounts
}
