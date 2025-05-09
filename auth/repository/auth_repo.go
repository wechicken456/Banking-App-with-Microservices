package repository

import (
	"auth/db/sqlc"
	"auth/model"
	"auth/utils"
	"database/sql"
	"errors"

	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type AuthRepository struct {
	queries *sqlc.Queries
	db      *sqlx.DB
}

func NewAuthRepository(db *sqlx.DB) *AuthRepository {
	return &AuthRepository{queries: sqlc.New(db), db: db}
}

// wrap a function in a transaction and execute it
func (r *AuthRepository) execTx(ctx context.Context, fn func(*sqlc.Queries) error) error {
	tx, err := r.db.BeginTx(ctx, nil) // create a transaction
	if err != nil {
		return err
	}
	defer tx.Rollback()
	q := r.queries.WithTx(tx) // return a new queries object with the transaction

	if err := fn(q); err != nil { // execute the function with the transaction
		return err
	}

	return tx.Commit()
}

// create user in a transaction: check if user already exists, if not create user
func (r *AuthRepository) CreateUserTx(ctx context.Context, user *model.User) (*sqlc.User, error) {

	var createdUser sqlc.User

	err := r.execTx(ctx, func(q *sqlc.Queries) error {
		var err error

		// check if user already exists
		_, err = q.GetUserByEmail(ctx, user.Email)
		if err == nil {
			return errors.New("user already exists")
		}
		// check if the error is due to no rows found
		if errors.Is(err, sql.ErrNoRows) {
			return err
		}

		passwordHash, err := utils.HashPassword(user.Password)
		if err != nil {
			return err
		}

		// create user
		createdUser, err = q.CreateUser(ctx, sqlc.CreateUserParams{
			ID:           uuid.New(),
			Email:        user.Email,
			PasswordHash: passwordHash,
		})
		if err != nil {
			return err
		}
		return nil
	})
	return &createdUser, err
}

// business logic (invalid password, etc.) should be handled in the service layer
// get user by email
// if user exists, return user
func (r *AuthRepository) UserLogin(ctx context.Context, email string) (*sqlc.User, error) {
    user, err := r.queries.GetUserByEmail(ctx, email)
    if err != nil {
        return nil, err
    }
    return &user, nil
}


