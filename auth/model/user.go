package model

import (
	"errors"

	"github.com/google/uuid"
)

type User struct {
	UserID   uuid.UUID `json:"user_id"`
	Email    string    `json:"email"`
	Password string    `json:"password"`
}

var ErrInternalServer error = errors.New("internal server error")
var ErrUserAlreadyExists error = errors.New("user already exists")
