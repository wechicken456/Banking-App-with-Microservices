// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package sqlc

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type IdempotencyKey struct {
	KeyID           uuid.UUID    `json:"key_id"`
	Status          string       `json:"status"`
	ResponseMessage string       `json:"response_message"`
	CreatedAt       sql.NullTime `json:"created_at"`
	UpdatedAt       sql.NullTime `json:"updated_at"`
	ExpiredAt       sql.NullTime `json:"expired_at"`
}

type RefreshToken struct {
	ID        uuid.UUID    `json:"id"`
	UserID    uuid.UUID    `json:"user_id"`
	Token     string       `json:"token"`
	ExpiredAt time.Time    `json:"expired_at"`
	CreatedAt sql.NullTime `json:"created_at"`
}

type User struct {
	ID           uuid.UUID    `json:"id"`
	Email        string       `json:"email"`
	PasswordHash string       `json:"password_hash"`
	CreatedAt    sql.NullTime `json:"created_at"`
	UpdatedAt    sql.NullTime `json:"updated_at"`
}
