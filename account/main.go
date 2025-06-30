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
	"strconv"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

func main() {
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

	_port := os.Getenv("GRPC_PORT")
	var port int
	var err error
	if port, err = strconv.Atoi(_port); err != nil {
		port = 50002
	}

	// start the server
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer listener.Close()

	grpcServer := grpc.NewServer()
	proto.RegisterAccountServiceServer(grpcServer, accountHandler)
	fmt.Printf("Server started on port %d\n", port)
	if grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
