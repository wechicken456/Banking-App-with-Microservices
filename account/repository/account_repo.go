package repository

import (
	"account/db/sqlc"
	"account/model"
	"context"

	"account/utils"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type AccountRepository struct {
	queries *sqlc.Queries
	db      *sqlx.DB
}

func NewAccountRepository(db *sqlx.DB) *AccountRepository {
	return &AccountRepository{queries: sqlc.New(db), db: db}
}

func (r *AccountRepository) CreateAccount(ctx context.Context, user *model.User) (*sqlc.Account, error) {
	createdAccount, err := r.queries.CreateAccount(ctx, sqlc.CreateAccountParams{
		ID:            uuid.New(),
		UserID:        user.UserID,
		Balance:       user.Balance,
		AccountNumber: utils.RandomAccountNumber(),
	})
	if err != nil {
		return nil, err
	}
	return &createdAccount, nil
}

func (r *AccountRepository) GetAccountByAccountNumber(ctx context.Context, accountNumber int64) (*sqlc.Account, error) {
	account, err := r.queries.GetAccountByAccountNumber(ctx, accountNumber)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *AccountRepository) GetAccountByID(ctx context.Context, id uuid.UUID) (*sqlc.Account, error) {
	account, err := r.queries.GetAccountByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *AccountRepository) GetAccountByUserID(ctx context.Context, userID uuid.UUID) ([]sqlc.Account, error) {
	accounts, err := r.queries.GetAccountByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

// amount could be negative or positive
func (r *AccountRepository) AddToAccountBalance(ctx context.Context, accountNumber int64, amount int64) (*sqlc.Account, error) {
	account, err := r.queries.AddToAccountBalance(ctx, sqlc.AddToAccountBalanceParams{
		AccountNumber: accountNumber,
		Amount:        amount,
	})
	if err != nil {
		return nil, err
	}
	return &account, nil
}
