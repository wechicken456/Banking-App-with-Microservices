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
