package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"transfer/db/initialize"

	_ "github.com/lib/pq"
)

func main() {
	db := initialize.ConnectDB()
	defer db.Close()

	// start the server
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", os.Getenv("ACCOUNT_PORT")))
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer listener.Close()

	fmt.Printf("Server started on port %s\n", os.Getenv("ACCOUNT_PORT"))
}
