package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware" // Chi's built-in middleware

	"api-gateway/handler"                 // my HTTP handlers
	myMiddleware "api-gateway/middleware" // my custom AuthMiddleware
)

func main() {
	// TODO: initialize gRPC clients to microservices
	// authConn, err := grpc.Dial("auth-service-address:port", grpc.WithInsecure()) // Use secure credentials in production
	// if err != nil {
	//     log.Fatalf("did not connect to auth service: %v", err)
	// }
	// defer authConn.Close()
	// authServiceClient := pb.NewAuthServiceClient(authConn)

	r := chi.NewRouter()

	// --- Global Middleware (applies to all routes) ---
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer) // Catches panics and returns 500
	r.Use(middleware.URLFormat)

	// --- Public Endpoints (do NOT require JWT validation) ---
	r.Post("/login", handler.LoginHandler)
	r.Post("/users", handler.CreateUserHandler) // Renamed from CreateUser for clarity
	r.Post("/renew-token", handler.RenewTokenHandler)

	// --- Protected Endpoints (require JWT validation) ---
	// Create a sub-router or group where the AuthMiddleware will be applied.
	r.Group(func(r chi.Router) {
		r.Use(myMiddleware.AuthMiddleware) // JWT valdiation happens in this middleware

		// TODO: add more routes to microservices
		r.Delete("/delete-user", handler.DeleteUserHandler)
		r.Get("/profile", handler.ProfileHandler)
	})

	// Start the HTTP server
	port := ":8080"
	log.Printf("API Gateway listening on port %s", port)
	log.Fatal(http.ListenAndServe(port, r)) // Use r (the Chi router) as the handler
}
