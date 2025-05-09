package main

import (
	"auth/db/initialize"
	"fmt"
	"log"
	"net"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	initialize.LoadDotEnv()
	db := initialize.ConnectDB()
	defer db.Close()

	// start the server
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", os.Getenv("AUTH_PORT")))
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer listener.Close()

	fmt.Printf("Server started on port %s\n", os.Getenv("AUTH_PORT"))
}
