package repository

import (
	"auth/db/sqlc"
	"auth/model"
	"auth/utils"
	"context"
	"database/sql"
	"errors"

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

// WithTx returns a new AuthRepository that uses the provided transaction.
func (r *AuthRepository) WithTx(tx *sql.Tx) *AuthRepository {
	return &AuthRepository{
		queries: r.queries.WithTx(tx),
		db:      r.db,
	}
}

func convertToModelUser(user sqlc.User) *model.User {
	return &model.User{
		UserID:   user.ID,
		Email:    user.Email,
		Password: user.PasswordHash,
	}
}

func convertToCreateUserParams(user *model.User) sqlc.CreateUserParams {
	return sqlc.CreateUserParams{
		ID:           user.UserID,
		Email:        user.Email,
		PasswordHash: user.Password,
	}
}

func convertToModelRefreshTokenRepo(token sqlc.RefreshToken) *model.RefreshTokenRepo {
	return &model.RefreshTokenRepo{
		UserID:    token.UserID,
		TokenHash: token.Token,
		ExpiredAt: token.ExpiredAt,
	}
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
func (r *AuthRepository) CreateUserTx(ctx context.Context, user *model.User) (*model.User, error) {
	var createdUser sqlc.User

	err := r.execTx(ctx, func(q *sqlc.Queries) error {
		var err error

		// check if user already exists
		_, err = q.GetUserByEmail(ctx, user.Email)
		if err == nil {
			return errors.New("user already exists")
		}

		if errors.Is(err, sql.ErrNoRows) {
			return err
		}

		passwordHash, err := utils.HashPassword(user.Password)
		if err != nil {
			return err
		}

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
	if err != nil {
		return nil, err
	}

	modelUser := &model.User{
		Email:    createdUser.Email,
		Password: "", // Don't return the password
	}

	return modelUser, nil
}

func (r *AuthRepository) CreateUser(ctx context.Context, user *model.User) (*model.User, error) {
	createdUser, err := r.queries.CreateUser(ctx, convertToCreateUserParams(user))
	if err != nil {
		return nil, err
	}
	return convertToModelUser(createdUser), nil
}

func (r *AuthRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	user, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return convertToModelUser(user), nil
}

func (r *AuthRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	user, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return convertToModelUser(user), nil
}

func (r *AuthRepository) UpdateUser(ctx context.Context, user *model.User) (*model.User, error) {
	updatedUser, err := r.queries.UpdateUser(ctx, sqlc.UpdateUserParams{
		ID:           user.UserID,
		Email:        user.Email,
		PasswordHash: user.Password,
	})
	if err != nil {
		return nil, err
	}
	return convertToModelUser(updatedUser), nil
}

func (r *AuthRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	err := r.queries.DeleteUser(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

// business logic (invalid password, etc.) should be handled in the service layer
// get user by email
// if user exists, return user
func (r *AuthRepository) GetLoginPasswordHash(ctx context.Context, email string) (string, error) {
	user, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	return user.PasswordHash, nil
}

func (r *AuthRepository) CreateRefreshToken(ctx context.Context, token *model.RefreshTokenRepo) (*model.RefreshTokenRepo, error) {
	createdToken, err := r.queries.CreateRefreshToken(ctx, sqlc.CreateRefreshTokenParams{
		ID:        uuid.New(),
		UserID:    token.UserID,
		Token:     token.TokenHash,
		ExpiredAt: token.ExpiredAt,
	})
	if err != nil {
		return nil, err
	}
	return convertToModelRefreshTokenRepo(createdToken), nil
}

func (r *AuthRepository) GetRefreshToken(ctx context.Context, tokenString string) (*model.RefreshTokenRepo, error) {
	token, err := r.queries.GetRefreshToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}
	return convertToModelRefreshTokenRepo(token), nil
}

func (r *AuthRepository) GetOrClaimIdempotencyKey(ctx context.Context, idempotencyKey *model.IdempotencyKey) (*model.IdempotencyKey, error) {
	key, err := r.queries.GetOrClaimIdempotencyKey(ctx, idempotencyKey.KeyID)
	if err != nil {
		return nil, err
	}
	return &model.IdempotencyKey{
		KeyID:           key.KeyID,
		Status:          key.Status,
		ResponseMessage: key.ResponseMessage,
	}, nil
}

func (r *AuthRepository) UpdateIdempotencyKey(ctx context.Context, idempotencyKey *model.IdempotencyKey) (*model.IdempotencyKey, error) {
	_, err := r.queries.UpdateIdempotencyKey(ctx, sqlc.UpdateIdempotencyKeyParams{
		KeyID:           idempotencyKey.KeyID,
		Status:          idempotencyKey.Status,
		ResponseMessage: idempotencyKey.ResponseMessage,
	})
	if err != nil {
		return nil, err
	}
	return idempotencyKey, nil
}

func (r *AuthRepository) DeleteIdempotencyKeyByID(ctx context.Context, idempotencyKey string) error {
	err := r.queries.DeleteIdempotencyKeyByID(ctx, idempotencyKey)
	if err != nil {
		return err
	}
	return nil
}
