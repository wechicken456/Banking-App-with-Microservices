package model

import "github.com/google/uuid"

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
	IdempotencyKey  string        `json:"idempotency_key"`
	Amount          int64         `json:"amount"`
	TransactionType string        `json:"transaction_type"`
	Status          string        `json:"status"`
	TransferID      uuid.NullUUID `json:"transfer_id"`
}
