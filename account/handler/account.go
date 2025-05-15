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

func (s *AccountHandler) CreateAccount(ctx context.Context, req *proto.CreateAccountRequest) (*proto.CreateAcocuntResponse, error) {
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
	account, err := s.service.CreateAccount(ctx, user)
	if err != nil {
		log.Printf("gRPC CreateAccount: Failed to create account: %v\n", err)
		return nil, err
	}
	return &proto.CreateAcocuntResponse{
		Status:    http.StatusOK,
		AccountId: account.AccountID[:],
		Error:     nil,
	}, nil
}

func (s *AccountHandler) GetAccountsByUserID(ctx context.Context, req *proto.GetAccountsByUserIdRequest) (*proto.GetAccountsByUserIdResponse, error) {
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
	accounts, err = s.service.GetAccountsByUserID(ctx, accountID)
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

func (s *AccountHandler) UpdateAccountBalance(ctx context.Context, req *proto.UpdateAccountBalanceRequest) (*proto.UpdateAccountBalanceResponse, error) {
	var (
		accountID uuid.UUID
		account   *model.Account
		err       error
	)

	if accountID, err = uuid.FromBytes(req.AccountId); err != nil {
		log.Printf("gRPC AddToAccountBalance: Failed to convert account ID: %v\n", err)
		return nil, err
	}
	if account, err = s.service.AddToAccountBalance(ctx, req.AccountNumber, req.Amount, accountID); err != nil {
		log.Printf("gRPC AddToAccountBalance: Failed to add to account balance: %v\n", err)
		return nil, err
	}
	return &proto.UpdateAccountBalanceResponse{
		Status:         http.StatusOK,
		Error:          nil,
		UpdatedBalance: account.Balance,
	}, nil
}

func (s *AccountHandler) DeleteAccountByAccountNumber(ctx context.Context, req *proto.DeleteAccountByAccountNumberRequest) (*proto.DeleteAccountByAccountNumberResponse, error) {
	var (
		userID uuid.UUID
		err    error
	)

	if userID, err = uuid.FromBytes(req.UserId); err != nil {
		log.Printf("gRPC DeleteAccountByAccountNumber: Failed to convert account ID: %v\n", err)
		return nil, err
	}

	if err = s.service.DeleteAccountByAccountNumber(ctx, req.AccountNumber, userID); err != nil {
		log.Printf("gRPC DeleteAccountByAccountNumber: Failed to delete account: %v\n", err)
		return nil, err
	}
	return &proto.DeleteAccountByAccountNumberResponse{
		Status: http.StatusOK,
		Error:  nil,
	}, nil
}
