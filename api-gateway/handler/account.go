package handler

import (
	"api-gateway/initialize"
	"api-gateway/proto"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewAccountServiceClient() proto.AccountServiceClient {
	initialize.LoadDotEnv()

	// Connect to the account service
	connString := fmt.Sprintf("%s:%s", os.Getenv("ACCOUNT_SERVICE_HOST"), os.Getenv("ACCOUNT_SERVICE_PORT"))
	fmt.Printf("Connecting to account service at %s\n", connString)
	conn, err := grpc.NewClient(connString, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Printf("Failed to connect to account service: %v", err)
		return nil
	}
	client := proto.NewAccountServiceClient(conn)
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

	req := &proto.CreateAccountRequest{
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

func GetAccountsByUserID(userID string) []*proto.Account {
	client := NewAccountServiceClient()
	if client == nil {
		log.Fatalf("Failed to create account service client")
	}

	userIDBytes, err := uuid.Parse(userID)
	if err != nil {
		log.Printf("Failed to parse user ID: %v", err)
		return nil
	}

	req := &proto.GetAccountsByUserIdRequest{
		UserId: userIDBytes[:],
	}

	res, err := client.GetAccountsByUserId(context.Background(), req)
	if err != nil {
		log.Printf("Failed to get accounts by user ID: %v", err)
		return
	}
	log.Printf("Accounts: %v", res)
	return res
}
