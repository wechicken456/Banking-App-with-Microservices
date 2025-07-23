package middleware

import (
	"api-gateway/model"
	"api-gateway/utils"
	"context"
	"fmt"
	"net/http"
	"strings"
)

// Define a custom type for context keys to avoid collisions
type contextKey string

const (
	// UserIDContextKey is the key for the **authenticated** user ID to pass down to handlers
	UserIDContextKey contextKey = "requestingUserID"
)

// AuthMiddleware validates the JWT token from the Authorization header.
// If valid, it extracts claims and adds them to the request context.
// If invalid or missing, it writes an HTTP error response.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Extract JWT and fingerprint from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization token is not provided", http.StatusUnauthorized)
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Invalid authorization format. Must be 'Bearer <token>'", http.StatusUnauthorized)
			return
		}
		jwtToken := strings.TrimPrefix(authHeader, "Bearer ")

		fingerprintCookie, err := r.Cookie(model.FingerprintCookieName)
		if err != nil {
			http.Error(w, "Invalid fingerprint cookie", http.StatusUnauthorized)
			return
		}
		fingerprint := fingerprintCookie.Value

		// 2. Validate the JWT
		claims, err := utils.ValidateJWT(jwtToken, fingerprint) // Your utility function to validate JWT
		if err != nil {
			// Log the error for debugging on the server side
			fmt.Printf("JWT validation failed for request to %s: %v\n", r.URL.Path, err)
			http.Error(w, "Invalid or expired JWT token", http.StatusUnauthorized)
			return
		}

		if claims.Subject == "" {
			http.Error(w, "Invalid JWT subject", http.StatusUnauthorized)
			return
		}

		// 3. Inject validated claims into the request context
		// This makes the userID (and other claims) available to downstream handlers.
		ctx := r.Context()
		ctx = context.WithValue(ctx, UserIDContextKey, claims.Subject) // Using Subject for user ID

		// 4. Call the next handler in the chain with the updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
