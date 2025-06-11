package handler

import (
	"api-gateway/client"
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
		utils.WriteGRPCErrorToHTTP(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(&model.LoginUserResponse{
		UserID:       res.UserId,
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
		Fingerprint:  res.Fingerprint,
	}); err != nil {
		log.Printf("LoginUserHandler: coudln't parse user_id: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	log.Println("LoginUserHandler: successful")
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
		log.Printf("CreateUserHandler: coudln't parse user_id: %v\n", err)
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
	userID := queryParams.Get("user_id")
	if userID == "" {
		http.Error(w, "Missing argument user_id", http.StatusBadRequest)
		return
	}

	// get idempotency key from header
	idempotencyKey := r.Header.Get("Idempotency-Key")

	// TODO: use gRPC client to call the auth microservice here
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
