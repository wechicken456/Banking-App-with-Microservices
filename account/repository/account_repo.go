package repository

import (
	"account/db/sqlc"
	"account/model"
	"account/utils"
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type AccountRepository struct {
	queries *sqlc.Queries
	db      *sqlx.DB
}

// NewAccountRepository creates a new AccountRepository.
func NewAccountRepository(db *sqlx.DB) *AccountRepository {
	return &AccountRepository{queries: sqlc.New(db), db: db}
}

// WithTx returns a new AccountRepository that uses the provided transaction.
func (r *AccountRepository) WithTx(tx *sql.Tx) *AccountRepository {
	return &AccountRepository{
		queries: r.queries.WithTx(tx),
		db:      r.db,
	}
}

func convertToModelAccount(account sqlc.Account) *model.Account {
	return &model.Account{
		AccountID:     account.ID,
		UserID:        account.UserID,
		Balance:       account.Balance,
		AccountNumber: account.AccountNumber,
	}
}

func convertToModelTransaction(transaction sqlc.Transaction) *model.Transaction {
	return &model.Transaction{
		TransactionID:   transaction.ID,
		AccountID:       transaction.AccountID,
		IdempotencyKey:  transaction.IdempotencyKey,
		Amount:          transaction.Amount,
		TransactionType: transaction.TransactionType,
		Status:          transaction.Status,
		TransferID:      transaction.TransferID,
	}
}

func convertToCreateTransactionParams(transaction *model.Transaction) *sqlc.CreateTransactionParams {
	return &sqlc.CreateTransactionParams{
		ID:              transaction.TransactionID,
		IdempotencyKey:  transaction.IdempotencyKey,
		AccountID:       transaction.AccountID,
		Amount:          transaction.Amount,
		TransactionType: transaction.TransactionType,
		Status:          transaction.Status,
		TransferID:      transaction.TransferID,
	}
}

func convertToCreateAccountParams(account *model.Account) *sqlc.CreateAccountParams {
	return &sqlc.CreateAccountParams{
		ID:            account.AccountID,
		UserID:        account.UserID,
		Balance:       account.Balance,
		AccountNumber: account.AccountNumber,
	}
}

func (r *AccountRepository) CreateAccount(ctx context.Context, user *model.User) (*model.Account, error) {
	createdAccount, err := r.queries.CreateAccount(ctx, sqlc.CreateAccountParams{
		ID:            uuid.New(),
		UserID:        user.UserID,
		Balance:       user.Balance,
		AccountNumber: utils.RandomAccountNumber(),
	})
	if err != nil {
		return nil, err
	}
	return convertToModelAccount(createdAccount), nil
}

func (r *AccountRepository) GetAccountByAccountNumber(ctx context.Context, accountNumber int64) (*model.Account, error) {
	account, err := r.queries.GetAccountByAccountNumber(ctx, accountNumber)
	if err != nil {
		return nil, err
	}
	return convertToModelAccount(account), nil
}

func (r *AccountRepository) GetAccountByID(ctx context.Context, id uuid.UUID) (*model.Account, error) {
	account, err := r.queries.GetAccountByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return convertToModelAccount(account), nil
}

func (r *AccountRepository) GetAccountByUserID(ctx context.Context, userID uuid.UUID) ([]model.Account, error) {
	accounts, err := r.queries.GetAccountByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	modelAccounts := make([]model.Account, len(accounts))
	for i, account := range accounts {
		modelAccounts[i] = *convertToModelAccount(account)
	}
	return modelAccounts, nil
}

func (r *AccountRepository) AddToAccountBalance(ctx context.Context, accountNumber int64, amount int64) (*model.Account, error) {
	account, err := r.queries.AddToAccountBalance(ctx, sqlc.AddToAccountBalanceParams{
		AccountNumber: accountNumber,
		Amount:        amount,
	})
	if err != nil {
		return nil, err
	}
	return convertToModelAccount(account), nil
}

func (r *AccountRepository) DeleteAccountByAccountNumber(ctx context.Context, accountNumber int64) error {
	err := r.queries.DeleteAccountByAccountNumber(ctx, accountNumber)
	if err != nil {
		return err
	}
	return nil
}

func (r *AccountRepository) CreateTransaction(ctx context.Context, transaction *model.Transaction) (*model.Transaction, error) {
	createdTransaction, err := r.queries.CreateTransaction(ctx, sqlc.CreateTransactionParams{
		ID:              transaction.TransactionID,
		IdempotencyKey:  transaction.IdempotencyKey,
		AccountID:       transaction.AccountID,
		Amount:          transaction.Amount,
		TransactionType: transaction.TransactionType,
		Status:          transaction.Status,
		TransferID:      transaction.TransferID,
	})
	if err != nil {
		return nil, err
	}
	return convertToModelTransaction(createdTransaction), err
}

func (r *AccountRepository) GetTransactionByIdempotencyKey(ctx context.Context, idempotencyKey string) (*model.Transaction, error) {
	transaction, err := r.queries.GetTransactionByIdempotencyKey(ctx, idempotencyKey)
	if err != nil {
		return nil, err
	}
	return convertToModelTransaction(transaction), nil
}

func (r *AccountRepository) GetTransactionByID(ctx context.Context, id uuid.UUID) (*model.Transaction, error) {
	transaction, err := r.queries.GetTransactionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return convertToModelTransaction(transaction), nil
}

func (r *AccountRepository) GetTransactionsByAccountID(ctx context.Context, accountID uuid.UUID) ([]model.Transaction, error) {
	transactions, err := r.queries.GetTransactionsByAccountID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	modelTransactions := make([]model.Transaction, len(transactions))
	for i, transaction := range transactions {
		modelTransactions[i] = *convertToModelTransaction(transaction)
	}
	return modelTransactions, nil
}
