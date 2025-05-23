// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: accounts.sql

package sqlc

import (
	"context"

	"github.com/google/uuid"
)

const addToAccountBalance = `-- name: AddToAccountBalance :one
UPDATE accounts
SET balance = balance + $1
WHERE account_number = $2
RETURNING id, user_id, account_number, balance, created_at, updated_at
`

type AddToAccountBalanceParams struct {
	Amount        int64 `json:"amount"`
	AccountNumber int64 `json:"account_number"`
}

func (q *Queries) AddToAccountBalance(ctx context.Context, arg AddToAccountBalanceParams) (Account, error) {
	row := q.db.QueryRowContext(ctx, addToAccountBalance, arg.Amount, arg.AccountNumber)
	var i Account
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.AccountNumber,
		&i.Balance,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const createAccount = `-- name: CreateAccount :one
INSERT INTO accounts (id, account_number, user_id, balance)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, account_number, balance, created_at, updated_at
`

type CreateAccountParams struct {
	ID            uuid.UUID `json:"id"`
	AccountNumber int64     `json:"account_number"`
	UserID        uuid.UUID `json:"user_id"`
	Balance       int64     `json:"balance"`
}

func (q *Queries) CreateAccount(ctx context.Context, arg CreateAccountParams) (Account, error) {
	row := q.db.QueryRowContext(ctx, createAccount,
		arg.ID,
		arg.AccountNumber,
		arg.UserID,
		arg.Balance,
	)
	var i Account
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.AccountNumber,
		&i.Balance,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteAccountByAccountNumber = `-- name: DeleteAccountByAccountNumber :exec
DELETE FROM accounts
WHERE account_number = $1
RETURNING id, user_id, account_number, balance, created_at, updated_at
`

func (q *Queries) DeleteAccountByAccountNumber(ctx context.Context, accountNumber int64) error {
	_, err := q.db.ExecContext(ctx, deleteAccountByAccountNumber, accountNumber)
	return err
}

const getAccountByAccountNumber = `-- name: GetAccountByAccountNumber :one
SELECT id, user_id, account_number, balance, created_at, updated_at FROM accounts WHERE account_number = $1
`

func (q *Queries) GetAccountByAccountNumber(ctx context.Context, accountNumber int64) (Account, error) {
	row := q.db.QueryRowContext(ctx, getAccountByAccountNumber, accountNumber)
	var i Account
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.AccountNumber,
		&i.Balance,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getAccountByID = `-- name: GetAccountByID :one
SELECT id, user_id, account_number, balance, created_at, updated_at FROM accounts WHERE id = $1
`

func (q *Queries) GetAccountByID(ctx context.Context, id uuid.UUID) (Account, error) {
	row := q.db.QueryRowContext(ctx, getAccountByID, id)
	var i Account
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.AccountNumber,
		&i.Balance,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getAccountsByUserID = `-- name: GetAccountsByUserID :many
SELECT id, user_id, account_number, balance, created_at, updated_at FROM accounts WHERE user_id = $1
`

func (q *Queries) GetAccountsByUserID(ctx context.Context, userID uuid.UUID) ([]Account, error) {
	rows, err := q.db.QueryContext(ctx, getAccountsByUserID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Account
	for rows.Next() {
		var i Account
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.AccountNumber,
			&i.Balance,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listAccounts = `-- name: ListAccounts :many
SELECT id, user_id, account_number, balance, created_at, updated_at FROM accounts ORDER BY id LIMIT $1
`

func (q *Queries) ListAccounts(ctx context.Context, limit int32) ([]Account, error) {
	rows, err := q.db.QueryContext(ctx, listAccounts, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Account
	for rows.Next() {
		var i Account
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.AccountNumber,
			&i.Balance,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
