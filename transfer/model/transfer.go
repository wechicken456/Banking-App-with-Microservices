package model

import "github.com/google/uuid"

type Transfer struct {
	TransferID     uuid.UUID `json:"transfer_id"`
	FromAccountID  uuid.UUID `json:"from_account_id"`
	ToAccountID    uuid.UUID `json:"to_account_id"`
	Amount         int64     `json:"amount"`
	IdempotencyKey string    `json:"idempotency_key"`
	Status         string    `json:"status"`
}

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
