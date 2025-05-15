package main

import (
	"account/db/initialize"
	"account/handler"
	"account/proto"
	"account/repository"
	"account/service"
	"fmt"
	"log"
	"net"
	"os"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

func main() {
	initialize.LoadDotEnv()
	db := initialize.ConnectDB()
	defer db.Close()
	accountRepo := repository.NewAccountRepository(db)
	if accountRepo == nil {
		log.Fatalf("Failed to create account repository")
	}
	accountService := service.NewAccountService(accountRepo, db)
	if accountService == nil {
		log.Fatalf("Failed to create account service")
	}
	accountHandler := handler.NewAccountHandler(accountService)
	if accountHandler == nil {
		log.Fatalf("Failed to create account handler")
	}

	// start the server
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", os.Getenv("ACCOUNT_PORT")))
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer listener.Close()

	grpcServer := grpc.NewServer()
	proto.RegisterAccountServiceServer(grpcServer, accountHandler)
	fmt.Printf("Server started on port %s\n", os.Getenv("ACCOUNT_PORT"))
	if grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
