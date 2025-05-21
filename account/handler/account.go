package handler

import (
	"account/model"
	"account/proto"
	"account/service"
	"context"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type AccountHandler struct {
	proto.UnimplementedAccountServiceServer
	service *service.AccountService
}

func NewAccountHandler(service *service.AccountService) *AccountHandler {
	return &AccountHandler{service: service}
}

func (h *AccountHandler) CreateAccount(ctx context.Context, req *proto.CreateAccountRequest) (*proto.CreateAccountResponse, error) {
	var (
		user   *model.User
		userID uuid.UUID
		err    error
	)

	if userID, err = uuid.FromBytes(req.UserId); err != nil {
		log.Printf("gRPC CreateAccount: Failed to convert user ID: %v\n", err)
		return nil, err
	}
	user = &model.User{
		UserID:  userID,
		Balance: req.Balance,
	}
	account, err := h.service.CreateAccount(ctx, user)
	if err != nil {
		log.Printf("gRPC CreateAccount: Failed to create account: %v\n", err)
		return nil, err
	}
	return &proto.CreateAccountResponse{
		Status:    http.StatusOK,
		AccountId: account.AccountID[:],
		Error:     nil,
	}, nil
}

func (h *AccountHandler) GetAccountsByUserID(ctx context.Context, req *proto.GetAccountsByUserIdRequest) (*proto.GetAccountsByUserIdResponse, error) {
	var (
		accountID    uuid.UUID
		accounts     []*model.Account
		grpcAccounts []*proto.Account
		err          error
	)

	if accountID, err = uuid.FromBytes(req.UserId); err != nil {
		log.Printf("gRPC GetAccount: Failed to convert account ID: %v\n", err)
		return nil, err
	}
	accounts, err = h.service.GetAccountsByUserID(ctx, accountID)
	if err != nil {
		log.Printf("gRPC GetAccount: Failed to get account: %v\n", err)
		return nil, err
	}

	grpcAccounts = make([]*proto.Account, len(accounts))
	for i, account := range accounts {
		grpcAccounts[i] = &proto.Account{
			AccountId:     account.AccountID[:],
			AccountNumber: account.AccountNumber,
			Balance:       account.Balance,
			UserId:        account.UserID[:],
		}
	}
	return &proto.GetAccountsByUserIdResponse{
		Status:   http.StatusOK,
		Accounts: grpcAccounts,
		Error:    nil,
	}, nil
}

func (h *AccountHandler) CreateTransaction(ctx context.Context, req *proto.CreateTransactionRequest) (*proto.CreateTransactionResponse, error) {
	var (
		userID      uuid.UUID
		accountID   uuid.UUID
		transaction *model.Transaction
		err         error
	)

	if accountID, err = uuid.FromBytes(req.AccountId); err != nil {
		log.Printf("gRPC AddToAccountBalance: Failed to convert account ID: %v\n", err)
		return nil, err
	}
	if userID, err = uuid.FromBytes(req.ReqUserId); err != nil {
		log.Printf("gRPC AddToAccountBalance: Failed to convert req user ID: %v\n", err)
		return nil, err
	}
	transaction = &model.Transaction{
		AccountID:       accountID,
		Amount:          req.Amount,
		TransactionType: string(req.TransactionType),
		Status:          "PENDING",
		TransferID:      uuid.NullUUID{uuid.UUID(req.TransferId), true},
	}

	if transaction, err = h.service.CreateTransaction(ctx, transaction, uuid.UUID(req.IdempotentKey), userID); err != nil {
		log.Printf("gRPC AddToAccountBalance: Failed to create transaction: %v\n", err)
		return nil, err
	}
	return &proto.CreateTransactionResponse{
		Status:        http.StatusOK,
		Error:         nil,
		TransactionId: transaction.TransactionID[:],
	}, nil
}

func (h *AccountHandler) DeleteAccountByAccountNumber(ctx context.Context, req *proto.DeleteAccountByAccountNumberRequest) (*proto.DeleteAccountByAccountNumberResponse, error) {
	var (
		userID uuid.UUID
		err    error
	)

	if userID, err = uuid.FromBytes(req.UserId); err != nil {
		log.Printf("gRPC DeleteAccountByAccountNumber: Failed to convert account ID: %v\n", err)
		return nil, err
	}

	if err = h.service.DeleteAccountByAccountNumber(ctx, req.AccountNumber, uuid.UUID(req.IdempotentKey), userID); err != nil {
		log.Printf("gRPC DeleteAccountByAccountNumber: Failed to delete account: %v\n", err)
		return nil, err
	}
	return &proto.DeleteAccountByAccountNumberResponse{
		Status: http.StatusOK,
		Error:  nil,
	}, nil
}

func (h *AccountHandler) GetAccountByAccountNumber(ctx context.Context, req *proto.GetAccountByAccountNumberRequest) (*proto.GetAccountByAccountNumberResponse, error) {
	var (
		accountID uuid.UUID
		account   *model.Account
		err       error
	)

	if accountID, err = uuid.FromBytes(req.UserId); err != nil {
		log.Printf("gRPC GetAccountByAccountNumber: Failed to convert account ID: %v\n", err)
		return nil, err
	}
	if account, err = h.service.GetAccountByAccountNumber(ctx, req.AccountNumber, accountID); err != nil {
		log.Printf("gRPC GetAccountByAccountNumber: Failed to get account: %v\n", err)
		return nil, err
	}
	return &proto.GetAccountByAccountNumberResponse{
		Status: http.StatusOK,
		Error:  nil,
		Account: &proto.Account{
			AccountId:     account.AccountID[:],
			AccountNumber: account.AccountNumber,
			UserId:        account.UserID[:],
			Balance:       account.Balance,
		},
	}, nil
}
