package model

import "github.com/google/uuid"

type Transfer struct {
	TransferID     uuid.UUID `json:"transfer_id"`
	FromAccountID  uuid.UUID `json:"from_account_id"`
	ToAccountID    uuid.UUID `json:"to_account_id"`
	Amount         int64     `json:"amount"`
	IdempotencyKey string    `json:"idempotency_key"`
}
