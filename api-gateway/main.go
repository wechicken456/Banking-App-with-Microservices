package main

import (
	"api-gateway/handler"
	"api-gateway/initialize"
	"api-gateway/proto"
	"context"
	"log"

	"github.com/google/uuid"
)

func main() {
	initialize.LoadDotEnv()

	accountClient := handler.NewAccountServiceClient()
	if accountClient == nil {
		panic("Failed to create account service client")
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
