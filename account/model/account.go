package model

import (
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type User struct {
	UserID  uuid.UUID `json:"user_id"`
	Balance int64     `json:"balance"`
}

type Account struct {
	AccountID     uuid.UUID `json:"account_id"`
	UserID        uuid.UUID `json:"user_id"`
	Balance       int64     `json:"balance"`
	AccountNumber int32     `json:"account_number"`
}

type Transaction struct {
	TransactionID   uuid.UUID     `json:"transaction_id"`
	AccountID       uuid.UUID     `json:"account_id"`
	Amount          int64         `json:"amount"`
	TransactionType string        `json:"transaction_type"` // DEPOSIT, WITHDRAWAL, TRANSFER_DEBIT, TRANSFER_CREDIT
	Status          string        `json:"status"`
	TransferID      uuid.NullUUID `json:"transfer_id"`
}

type IdempotencyKey struct {
	KeyID  string    `json:"key_id"`
	UserID uuid.UUID `json:"user_id"`

	Status          string `json:"status"`
	ResponseMessage string `json:"response_body"`
}

var (
	ErrInternalServer   error = status.Error(codes.Internal, "internal server error")
	ErrInvalidArgument  error = status.Error(codes.InvalidArgument, "invalid argument")
	ErrNotAuthorized    error = status.Error(codes.PermissionDenied, "not authorized")
	ErrNotAuthenticated error = status.Error(codes.Unauthenticated, "not authenticated")
)
