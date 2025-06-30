package main

import (
	"auth/db/initialize"
	"auth/handler"
	"auth/proto"
	"auth/repository"
	"auth/service"
	"crypto/rand"
	"encoding/base64"
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

	// Set JWT secret key
	jwtSecretKey := make([]byte, 32)
	_, err := rand.Read(jwtSecretKey)
	if err != nil {
		log.Fatalf("failed to generate JWT secret key: %w", err)
	}
	os.Setenv("JWT_SECRET_KEY", base64.StdEncoding.EncodeToString(jwtSecretKey))

	_port := os.Getenv("GRPC_PORT")
	var port int
	if port, err = strconv.Atoi(_port); err != nil {
		port = 50001
	}

	// start the server
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer listener.Close()

	grpcServer := grpc.NewServer()
	proto.RegisterAuthServiceServer(grpcServer, authHandler)
	fmt.Printf("Server started on port %d\n", port)
	if grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
