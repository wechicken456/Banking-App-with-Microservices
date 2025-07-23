package handler

import (
	"account/model"
	"account/proto"
	"account/service"
	"context"
	"log"

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
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		log.Printf("gRPC CreateAccount: Failed to parse user ID: %v\n", err)
		return nil, model.ErrInvalidArgument
	}
	user := &model.User{
		UserID:  userID,
		Balance: req.Balance,
	}

	account, err := h.service.CreateAccount(ctx, user, req.IdempotencyKey, userID)
	if err != nil {
		log.Printf("gRPC CreateAccount: Failed to create account: %v\n", err)
		return nil, err
	}
	log.Printf("Created account: %v\n", account)
	return &proto.CreateAccountResponse{
		AccountId:     account.AccountID.String(),
		AccountNumber: account.AccountNumber,
	}, nil
}

func (h *AccountHandler) GetAccountsByUserId(ctx context.Context, req *proto.GetAccountsByUserIdRequest) (*proto.GetAccountsByUserIdResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		log.Printf("gRPC GetAccountsByUserID: Failed to parse user ID: %v\n", err)
		return nil, model.ErrInvalidArgument
	}

	accounts, err := h.service.GetAccountsByUserID(ctx, userID)
	if err != nil {
		log.Printf("gRPC GetAccountsByUserID: Failed to get accounts: %v\n", err)
		return nil, err
	}

	grpcAccounts := make([]*proto.Account, len(accounts))
	for i, account := range accounts {
		grpcAccounts[i] = &proto.Account{
			AccountId:     account.AccountID.String(),
			AccountNumber: account.AccountNumber,
			Balance:       account.Balance,
			UserId:        account.UserID.String(),
		}
	}

	return &proto.GetAccountsByUserIdResponse{
		Accounts: grpcAccounts,
	}, nil
}

func (h *AccountHandler) GetAccountByAccountNumber(ctx context.Context, req *proto.GetAccountByAccountNumberRequest) (*proto.Account, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		log.Printf("gRPC GetAccountByAccountNumber: Failed to parse user ID: %v\n", err)
		return nil, model.ErrInvalidArgument
	}

	account, err := h.service.GetAccountByAccountNumber(ctx, req.AccountNumber, userID)
	if err != nil {
		log.Printf("gRPC GetAccountByAccountNumber: Failed to get account: %v\n", err)
		return nil, err
	}

	return &proto.Account{
		AccountId:     account.AccountID.String(),
		AccountNumber: account.AccountNumber,
		Balance:       account.Balance,
		UserId:        account.UserID.String(),
	}, nil
}

func (h *AccountHandler) GetAccountByAccountId(ctx context.Context, req *proto.GetAccountByAccountIdRequest) (*proto.Account, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		log.Printf("gRPC GetAccountByAccountNumber: Failed to parse user ID: %v\n", err)
		return nil, model.ErrInvalidArgument
	}

	accountID, err := uuid.Parse(req.AccountId)
	if err != nil {
		log.Printf("gRPC GetAccountByAccountId: Failed to parse account ID: %v\n", err)
		return nil, model.ErrInvalidArgument
	}

	account, err := h.service.GetAccount(ctx, accountID, userID)
	if err != nil {
		log.Printf("gRPC GetAccountByAccountNumber: Failed to get account: %v\n", err)
		return nil, err
	}

	return &proto.Account{
		AccountId:     account.AccountID.String(),
		AccountNumber: account.AccountNumber,
		Balance:       account.Balance,
		UserId:        account.UserID.String(),
	}, nil
}

func (h *AccountHandler) DeleteAccountByAccountNumber(ctx context.Context, req *proto.DeleteAccountByAccountNumberRequest) (*proto.DeleteAccountByAccountNumberResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		log.Printf("gRPC DeleteAccountByAccountNumber: Failed to parse user ID: %v\n", err)
		return nil, model.ErrInvalidArgument
	}

	if err = h.service.DeleteAccountByAccountNumber(ctx, req.AccountNumber, req.IdempotencyKey, userID); err != nil {
		log.Printf("gRPC DeleteAccountByAccountNumber: Failed to delete account: %v\n", err)
		return nil, err
	}

	return &proto.DeleteAccountByAccountNumberResponse{}, nil
}

func (h *AccountHandler) CreateTransaction(ctx context.Context, req *proto.CreateTransactionRequest) (*proto.CreateTransactionResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		log.Printf("gRPC CreateTransaction: Failed to parse user ID: %v\n", err)
		return nil, model.ErrInvalidArgument
	}

	accountID, err := uuid.Parse(req.AccountId)
	if err != nil {
		log.Printf("gRPC CreateTransaction: Failed to parse account ID: %v\n", err)
		return nil, model.ErrInvalidArgument
	}

	var transferID uuid.NullUUID
	if req.TransferId != "" {
		transferUUID, err := uuid.Parse(req.TransferId)
		if err != nil {
			log.Printf("gRPC CreateTransaction: Failed to parse transfer ID: %v\n", err)
			return nil, model.ErrInvalidArgument
		}
		transferID = uuid.NullUUID{UUID: transferUUID, Valid: true}
	}

	transaction := &model.Transaction{
		AccountID:       accountID,
		Amount:          req.Amount,
		TransactionType: req.TransactionType,
		Status:          "PENDING",
		TransferID:      transferID,
	}

	

	createdTransaction, err := h.service.CreateTransaction(ctx, transaction, req.IdempotencyKey, userID)
	if err != nil {
		log.Printf("gRPC CreateTransaction: Failed to create transaction: %v\n", err)
		return nil, err
	}

	return &proto.CreateTransactionResponse{
		TransactionId: createdTransaction.TransactionID.String(),
	}, nil
}

func (h *AccountHandler) GetTransactionsByAccountId(ctx context.Context, req *proto.GetTransactionsByAccountIdRequest) (*proto.GetTransactionsByAccountIdResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		log.Printf("gRPC GetTransactionsByAccountId: Failed to parse user ID: %v\n", err)
		return nil, model.ErrInvalidArgument
	}

	accountID, err := uuid.Parse(req.AccountId)
	if err != nil {
		log.Printf("gRPC GetTransactionsByAccountId: Failed to parse account ID: %v\n", err)
		return nil, model.ErrInvalidArgument
	}

	transactions, err := h.service.GetTransactionsByAccountID(ctx, accountID, userID)
	if err != nil {
		log.Printf("gRPC GetTransactionsByAccountId: Failed to get transactions: %v\n", err)
		return nil, err
	}

	grpcTransactions := make([]*proto.Transaction, len(transactions))
	for i, transaction := range transactions {
		grpcTransactions[i] = &proto.Transaction{
			TransactionId:   transaction.TransactionID.String(),
			AccountId:       transaction.AccountID.String(),
			Amount:          transaction.Amount,
			TransactionType: transaction.TransactionType,
			Status:          transaction.Status,
			TransferId:      transaction.TransferID.UUID.String(),
		}
	}

	return &proto.GetTransactionsByAccountIdResponse{
		Transactions: grpcTransactions,
	}, nil
}

func (h *AccountHandler) ValidateAccountNumber(ctx context.Context, req *proto.ValidateAccountNumberRequest) (*proto.ValidateAccountNumberResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		log.Printf("gRPC ValidateAccountNumber: Failed to parse user ID: %v\n", err)
		return nil, model.ErrInvalidArgument
	}

	valid, err := h.service.ValidateAccountNumber(ctx, req.AccountNumber, userID)
	if err != nil {
		log.Printf("gRPC ValidateAccountNumber: Failed to validate account: %v\n", err)
		return nil, err
	}

	return &proto.ValidateAccountNumberResponse{
		Valid: valid,
	}, nil
}

func (h *AccountHandler) HasSufficientBalance(ctx context.Context, req *proto.HasSufficientBalanceRequest) (*proto.HasSufficientBalanceResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		log.Printf("gRPC HasSufficientBalance: Failed to parse user ID: %v\n", err)
		return nil, model.ErrInvalidArgument
	}

	sufficient, err := h.service.HasSufficientBalance(ctx, req.AccountNumber, req.Amount, userID)
	if err != nil {
		log.Printf("gRPC HasSufficientBalance: Failed to check balance: %v\n", err)
		return nil, err
	}

	return &proto.HasSufficientBalanceResponse{
		Sufficient: sufficient,
	}, nil
}
