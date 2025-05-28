package model

import (
	"errors"

	"github.com/google/uuid"
)

type User struct {
	UserID  uuid.UUID `json:"user_id"`
	Balance int64     `json:"balance"`
}

type Account struct {
	AccountID     uuid.UUID `json:"account_id"`
	UserID        uuid.UUID `json:"user_id"`
	Balance       int64     `json:"balance"`
	AccountNumber int64     `json:"account_number"`
}

type Transaction struct {
	TransactionID   uuid.UUID     `json:"transaction_id"`
	AccountID       uuid.UUID     `json:"account_id"`
	Amount          int64         `json:"amount"`
	TransactionType string        `json:"transaction_type"`
	Status          string        `json:"status"`
	TransferID      uuid.NullUUID `json:"transfer_id"`
}

type IdempotencyKey struct {
	KeyID  uuid.UUID `json:"key_id"`
	UserID uuid.UUID `json:"user_id"`

	Status          string `json:"status"`
	ResponseMessage string `json:"response_body"`
}

var (
	ErrInternalServer       error = errors.New("internal server error")
	ErrIdempotencyKeyExists error = errors.New("idempotency key already exists")
	ErrUserIDMismatch       error = errors.New("user id mismatch")
)
