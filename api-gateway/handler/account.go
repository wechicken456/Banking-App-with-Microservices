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
	"strconv"

	proto "buf.build/gen/go/banking-app/account/protocolbuffers/go"
	"github.com/google/uuid"
)

type AccountHandler struct {
	Client *client.AccountClient
}

func NewAccountHandler(client *client.AccountClient) *AccountHandler {
	return &AccountHandler{Client: client}
}

// CreateAccountHandler creates a new account
func (h *AccountHandler) CreateAccountHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var createAccountReq model.CreateAccountRequest
	if err := DecodeJSONBody(w, r, &createAccountReq); err != nil {
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

	// get the userID from the request context (passed by AuthMiddleware)
	ctx := r.Context()
	requestingUserID := ctx.Value(middleware.UserIDContextKey).(string)
	if requestingUserID == "" {
		http.Error(w, "Missing user authentication", http.StatusUnauthorized)
		return
	}

	// use gRPC client to call the account microservice
	res, err := h.Client.CreateAccount(context.Background(), &proto.CreateAccountRequest{
		UserId:         requestingUserID,
		Balance:        createAccountReq.Balance,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		log.Printf("CreateAccountHandler: %v", err)
		utils.WriteGRPCErrorToHTTP(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(&model.CreateAccountResponse{AccountID: res.AccountId}); err != nil {
		log.Printf("CreateAccountHandler: couldn't encode response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	log.Println("CreateAccountHandler: successful")
}

// GetAccountsByUserIDHandler gets all accounts for the authenticated user
func (h *AccountHandler) GetAccountsByUserIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// get the userID from the request context (passed by AuthMiddleware)
	ctx := r.Context()
	requestingUserID := ctx.Value(middleware.UserIDContextKey).(string)
	if requestingUserID == "" {
		http.Error(w, "Missing user authentication", http.StatusUnauthorized)
		return
	}

	userIDBytes, err := uuid.Parse(requestingUserID)
	if err != nil {
		log.Printf("GetAccountsByUserIDHandler: Failed to parse user ID: %v", err)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// use gRPC client to call the account microservice
	res, err := h.Client.GetAccountsByUserId(context.Background(), &proto.GetAccountsByUserIdRequest{
		UserId: userIDBytes.String(),
	})
	if err != nil {
		log.Printf("GetAccountsByUserIDHandler: %v", err)
		utils.WriteGRPCErrorToHTTP(w, err)
		return
	}

	resp := model.GetAccountsByUserIDResponse{}
	for _, acc := range res.Accounts {
		tmp := model.Account{}
		tmp.AccountID = acc.AccountId
		tmp.AccountNumber = acc.AccountNumber
		tmp.Balance = acc.Balance
		tmp.UserID = acc.UserId
		resp.Accounts = append(resp.Accounts, tmp)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(&resp); err != nil {
		log.Printf("GetAccountsByUserIDHandler: couldn't encode response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	log.Println("GetAccountsByUserIDHandler: successful")
}

// GetAccountByAccountNumberHandler gets a specific account by account number
func (h *AccountHandler) GetAccountByAccountNumberHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// get account number from URL parameter
	u := r.URL
	queryParams := u.Query()
	accountNumber, err := strconv.ParseInt(queryParams.Get("accountNumber"), 10, 64)
	if err != nil {
		http.Error(w, "invalid accountNumber", http.StatusBadRequest)
		return
	}

	// get the userID from the request context (passed by AuthMiddleware)
	ctx := r.Context()
	requestingUserID := ctx.Value(middleware.UserIDContextKey).(string)
	if requestingUserID == "" {
		http.Error(w, "Missing user authentication", http.StatusUnauthorized)
		return
	}

	userIDBytes, err := uuid.Parse(requestingUserID)
	if err != nil {
		log.Printf("GetAccountByAccountNumberHandler: Failed to parse user ID: %v", err)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// use gRPC client to call the account microservice
	res, err := h.Client.GetAccountByAccountNumber(context.Background(), &proto.GetAccountByAccountNumberRequest{
		AccountNumber: accountNumber,
		UserId:        userIDBytes.String(),
	})
	if err != nil {
		log.Printf("GetAccountByAccountNumberHandler: %v", err)
		utils.WriteGRPCErrorToHTTP(w, err)
		return
	}
	resp := model.GetAccountResponse{
		Account: model.Account{
			AccountNumber: res.Account.AccountNumber,
			AccountID:     res.Account.AccountId,
			Balance:       res.Account.Balance,
			UserID:        res.Account.UserId,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(&resp); err != nil {
		log.Printf("GetAccountByAccountNumberHandler: couldn't encode response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	log.Println("GetAccountByAccountNumberHandler: successful")
}

// DeleteAccountByAccountNumberHandler deletes an account by account number
func (h *AccountHandler) DeleteAccountByAccountNumberHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.DeleteAccountByAccountNumberRequest
	if err := DecodeJSONBody(w, r, &req); err != nil {
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

	// get the userID from the request context (passed by AuthMiddleware)
	ctx := r.Context()
	requestingUserID := ctx.Value(middleware.UserIDContextKey).(string)
	if requestingUserID == "" {
		http.Error(w, "Missing user authentication", http.StatusUnauthorized)
		return
	}

	userIDBytes, err := uuid.Parse(requestingUserID)
	if err != nil {
		log.Printf("DeleteAccountByAccountNumberHandler: Failed to parse user ID: %v", err)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// use gRPC client to call the account microservice
	_, err = h.Client.DeleteAccountByAccountNumber(context.Background(), &proto.DeleteAccountByAccountNumberRequest{
		AccountNumber:  req.AccountNumber,
		IdempotencyKey: idempotencyKey,
		UserId:         userIDBytes.String(),
	})
	if err != nil {
		log.Printf("DeleteAccountByAccountNumberHandler: %v", err)
		utils.WriteGRPCErrorToHTTP(w, err)
		return
	}

	log.Println("DeleteAccountByAccountNumberHandler: successful")
}

// CreateTransactionHandler creates a new transaction
func (h *AccountHandler) CreateTransactionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var createTransactionReq model.CreateTransactionRequest
	if err := DecodeJSONBody(w, r, &createTransactionReq); err != nil {
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

	ctx := r.Context()
	requestingUserID := ctx.Value(middleware.UserIDContextKey).(string)
	if requestingUserID == "" {
		http.Error(w, "Missing user authentication", http.StatusUnauthorized)
		return
	}

	userIDBytes, err := uuid.Parse(requestingUserID)
	if err != nil {
		log.Printf("CreateTransactionHandler: Failed to parse user ID: %v", err)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	accountIDBytes, err := uuid.Parse(createTransactionReq.AccountID)
	if err != nil {
		log.Printf("CreateTransactionHandler: Failed to parse account ID: %v", err)
		http.Error(w, "Invalid account ID", http.StatusBadRequest)
		return
	}

	// use gRPC client to call the account microservice
	res, err := h.Client.CreateTransaction(context.Background(), &proto.CreateTransactionRequest{
		AccountId:      accountIDBytes.String(),
		Amount:         createTransactionReq.Amount,
		IdempotencyKey: idempotencyKey,
		UserId:         userIDBytes.String(),
	})
	if err != nil {
		log.Printf("CreateTransactionHandler: %v", err)
		utils.WriteGRPCErrorToHTTP(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(&model.CreateTransactionResponse{TransactionID: res.TransactionId}); err != nil {
		log.Printf("CreateTransactionHandler: couldn't encode response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	log.Println("CreateTransactionHandler: successful")
}
