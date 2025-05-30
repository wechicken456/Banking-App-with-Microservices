package main

import (
	"auth/db/initialize"
	"auth/handler"
	"auth/proto"
	"auth/repository"
	"auth/service"
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
	authRepo := repository.NewAuthRepository(db)
	if authRepo == nil {
		log.Fatalf("Failed to create auth repository")
	}
	authService := service.NewAuthService(authRepo, db)
	if authService == nil {
		log.Fatalf("Failed to create auth service")
	}
	authHandler := handler.NewAuthHandler(authService)
	if authHandler == nil {
		log.Fatalf("Failed to create auth handler")
	}

	// start the server
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", os.Getenv("AUTH_PORT")))
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer listener.Close()

	grpcServer := grpc.NewServer()
	proto.RegisterAuthServiceServer(grpcServer, authHandler)
	fmt.Printf("Server started on port %s\n", os.Getenv("AUTH_PORT"))
	if grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
