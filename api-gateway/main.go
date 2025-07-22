package main

import (
	"api-gateway/client"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"api-gateway/handler" // my HTTP handlers

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware" // Chi's built-in middleware
	"github.com/go-chi/cors"

	myMiddleware "api-gateway/middleware" // my custom AuthMiddleware
)

func main() {
	// TODO: initialize gRPC clients to microservices
	// authConn, err := grpc.Dial("auth-service-address:port", grpc.WithInsecure()) // Use secure credentials in production
	authClient := client.NewAuthClient(os.Getenv("AUTH_SERVICE_URL"))
	authHandler := handler.NewAuthHandler(authClient)
	accountClient := client.NewAccountClient(os.Getenv("ACCOUNT_SERVICE_URL"))
	accountHandler := handler.NewAccountHandler(accountClient)

	r := chi.NewRouter()

	// --- Setup CORS for frontend access ---
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"}, // frontend origin
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "Accept", "Idempotency-Key"},
		ExposedHeaders:   []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Route("/api", func(r chi.Router) {
		// --- Global Middleware (applies to all routes) ---
		r.Use(middleware.RequestID)
		r.Use(middleware.Logger)
		r.Use(middleware.Recoverer) // Catches panics and returns 500
		r.Use(middleware.URLFormat)

		// --- Public Endpoints (do NOT require JWT validation) ---
		r.Post("/login", authHandler.LoginHandler)
		r.Post("/register", authHandler.CreateUserHandler)

		// --- Protected Endpoints (require JWT validation) ---
		// Create a sub-router or group where the AuthMiddleware will be applied.
		r.Group(func(r chi.Router) {
			r.Use(myMiddleware.AuthMiddleware) // JWT valdiation happens in this middleware

			// TODO: add more routes to microservices
			// user management
			r.Delete("/delete-user", authHandler.DeleteUserHandler)
			r.Post("/renew-token", authHandler.RenewAccessTokenHandler)

			// account management
			r.Get("/all-accounts", accountHandler.GetAccountsByUserIDHandler)
			r.Get("/account", accountHandler.GetAccountByAccountNumberHandler)
			r.Post("/create-account", accountHandler.CreateAccountHandler)
			r.Post("/delete-account", accountHandler.DeleteAccountByAccountNumberHandler)

			// transaction management
			r.Post("/create-transaction", accountHandler.CreateTransactionHandler)
			r.Get("/transactions", accountHandler.GetTransactionsByAccountIDHandler)
		})
	})

	// Start the HTTP server
	_port := os.Getenv("HTTP_PORT")
	var port int
	var err error
	if port, err = strconv.Atoi(_port); err != nil {
		port = 3000
	}
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), r)) // Use r (the Chi router) as the handler
	log.Printf("API Gateway listening on port %s", port)
}
