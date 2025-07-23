package handler

import (
	"api-gateway/client"
	"api-gateway/middleware"
	"api-gateway/model"
	"api-gateway/utils"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	proto "buf.build/gen/go/banking-app/auth/protocolbuffers/go"
)

type AuthHandler struct {
	Client *client.AuthClient
}

func NewAuthHandler(client *client.AuthClient) *AuthHandler {
	return &AuthHandler{client}
}

func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	var loginCreds model.LoginCreds
	if err := DecodeJSONBody(w, r, &loginCreds); err != nil {
		var mr *malformedRequest
		if errors.As(err, &mr) {
			log.Print(err.Error())
			http.Error(w, mr.msg, mr.status)
		} else {
			log.Print(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}
	// get idempotency key from header
	idempotencyKey := r.Header.Get("Idempotency-Key")

	res, err := h.Client.Login(context.Background(), &proto.LoginRequest{
		Email:          loginCreds.Email,
		Password:       loginCreds.Password,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		log.Printf("LoginHandler: %v\n", err)
		utils.WriteGRPCErrorToHTTP(w, err)
		return
	}
	fingerprintCookie := http.Cookie{
		Name:     model.FingerprintCookieName,
		Value:    res.Fingerprint,
		MaxAge:   int(model.TokenShortDuration),
		SameSite: http.SameSiteStrictMode,
		HttpOnly: true,
		Secure:   false, // TODO: set to true during production
	}

	http.SetCookie(w, &fingerprintCookie)
	w.Header().Set("Content-Type", "application/json")
	resBody := model.LoginResponse{
		UserID:               res.UserId,
		AccessToken:          res.AccessToken,
		Email:                loginCreds.Email,
		RefreshToken:         res.RefreshToken,
		AccessTokenDuration:  res.AccessTokenDuration,
		RefreshTokenDuration: res.RefreshTokenDuration,
	}
	if err := json.NewEncoder(w).Encode(&resBody); err != nil {
		log.Printf("LoginHandler: coudln't parse userId: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	log.Println("LoginHandler: successful")
}

func (h *AuthHandler) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	var loginCreds model.LoginCreds
	if err := DecodeJSONBody(w, r, &loginCreds); err != nil {
		var mr *malformedRequest
		if errors.As(err, &mr) {
			http.Error(w, mr.msg, mr.status)
		} else {
			log.Print(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	// get idempotency key from header
	idempotencyKey := r.Header.Get("Idempotency-Key")

	res, err := h.Client.CreateUser(context.Background(), &proto.CreateUserRequest{
		Email:          loginCreds.Email,
		Password:       loginCreds.Password,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		utils.WriteGRPCErrorToHTTP(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(&model.CreateUserResponse{UserID: res.UserId}); err != nil {
		log.Printf("CreateUserHandler: coudln't parse userId: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	log.Println("CreateUserHandler: successful")
}

func (h *AuthHandler) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	// get user id from the URL parameter
	u := r.URL
	queryParams := u.Query()
	userID := queryParams.Get("userId")
	if userID == "" {
		http.Error(w, "Missing argument userId", http.StatusBadRequest)
		return
	}

	// get idempotency key from header
	idempotencyKey := r.Header.Get("Idempotency-Key")

	// use gRPC client to call the auth microservice
	_, err := h.Client.DeleteUser(context.Background(), &proto.DeleteUserRequest{
		UserId:         userID,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		utils.WriteGRPCErrorToHTTP(w, err)
		return
	}

	log.Println("DeleteUserHandler: successful")
}

// Return a new JWT access token
// Requires the current JWT access token, and the refreshToken cookie
func (h *AuthHandler) RenewAccessTokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	// check presence of refreshToken
	refreshToken, err := r.Cookie(model.RefreshTokenCookieName)
	if err != nil {
		if err == http.ErrNoCookie {
			msg := "Missing refreshToken cookie"
			http.Error(w, msg, http.StatusBadRequest)
		} else {
			log.Print(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}
	// get idempotency key from header
	idempotencyKey := r.Header.Get("Idempotency-Key")

	// get the userID of the JWT access token attached to the request context which was passwd down by the AuthMiddleware
	ctx := r.Context()
	requestingUserID := ctx.Value(middleware.UserIDContextKey).(string)
	if requestingUserID == "" {
		msg := "Missing bearer token"
		http.Error(w, msg, http.StatusUnauthorized)
		return
	}

	// use gRPC client to call the auth microservice here
	res, err := h.Client.RenewAccessToken(r.Context(),
		&proto.RenewAccessTokenRequest{
			UserId:         requestingUserID,
			RefreshToken:   refreshToken.Value,
			IdempotencyKey: idempotencyKey,
		})
	if err != nil {
		log.Print(err.Error())
		utils.WriteGRPCErrorToHTTP(w, err)
		return
	}

	fingerprintCookie := http.Cookie{
		Name:     model.FingerprintCookieName,
		Value:    res.Fingerprint,
		MaxAge:   int(res.AccessTokenDuration),
		SameSite: http.SameSiteStrictMode,
		HttpOnly: true,
		Secure:   false, // TODO: set to true during production
	}
	http.SetCookie(w, &fingerprintCookie)
	w.Header().Set("Content-Type", "application/json")
	resBody := model.RenewAccessTokenResponse{
		AccessToken:         res.AccessToken,
		AccessTokenDuration: res.AccessTokenDuration,
	}
	if err := json.NewEncoder(w).Encode(&resBody); err != nil {
		log.Printf("RenewAccessTokenHandler: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	log.Println("RenewAccessTokenHandler: successful")
}
