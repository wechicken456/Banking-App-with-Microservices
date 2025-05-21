package service

import (
	"account/model"
	"account/repository"
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

var maxRetries = 3

type AccountService struct {
	repo *repository.AccountRepository
	db   *sqlx.DB
}

// r and db should be created in the main function and passed to the service
// sqlx.DB object maintains a connection pool internally, and will attempt to connect when a connection is first needed.
func NewAccountService(r *repository.AccountRepository, db *sqlx.DB) *AccountService {
	return &AccountService{repo: r, db: db}
}

func (s *AccountService) CreateAccount(ctx context.Context, user *model.User) (*model.Account, error) {
	return s.repo.CreateAccount(ctx, user)
}

func (s *AccountService) GetAccount(ctx context.Context, accountID uuid.UUID, userID uuid.UUID) (*model.Account, error) {
	account, err := s.repo.GetAccountByID(ctx, accountID)
	if err != nil {
		log.Printf("gRPC GetAccount service: Failed to get accounts: %v\n", err)
		if err == sql.ErrNoRows {
			return nil, errors.New("account not found")
		}
		return nil, model.ErrInternalServer
	}

	// Check if the user owns the account
	if account.UserID != userID {
		log.Printf("gRPC GetAccount service: Unauthorized access attempt for account id %v by user %v\n", accountID, userID)
		return nil, model.ErrInternalServer
	}

	return account, nil
}

func (s *AccountService) GetAccountsByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Account, error) {
	accounts, err := s.repo.GetAccountsByUserID(ctx, userID)
	if err != nil {
		log.Printf("gRPC GetAccountsByUserID service: Failed to get accounts: %v\n", err)
		return nil, model.ErrInternalServer
	}
	return accounts, nil
}

func (s *AccountService) GetAccountByAccountNumber(ctx context.Context, accountNumber int64, userID uuid.UUID) (*model.Account, error) {
	account, err := s.repo.GetAccountByAccountNumber(ctx, accountNumber)
	if err != nil {
		log.Printf("gRPC GetAccountByAccountNumber service: Failed to get account: %v\n", err)
		if err == sql.ErrNoRows {
			return nil, errors.New("account not found")
		}
		return nil, model.ErrInternalServer
	}

	// Check if the user owns the account
	if account.UserID != userID {
		log.Printf("gRPC GetAccountByAccountNumber service: Unauthorized access attempt for account number %v by user %v\n",
			accountNumber, userID)
		return nil, model.ErrInternalServer
	}

	return account, nil
}

// delete account by account number with exponential backoff retries
func (s *AccountService) DeleteAccountByAccountNumber(ctx context.Context, accountNumber int64, userID uuid.UUID) error {
	var err error
	for attempt := 0; attempt < maxRetries; attempt++ {
		err = s.deleteAccountByAccountNumberTx(ctx, accountNumber, userID)
		if err == nil {
			return nil
		}
		// Check for serialization failure (Postgres error code 40001: https://www.postgresql.org/docs/current/mvcc-serialization-failure-handling.html)
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "40001" {
			log.Printf("Serialization failure, retrying DeleteAccountByAccountNumber (attempt %d): %v", attempt+1, err)
			time.Sleep(time.Duration(attempt+1) * 100 * time.Millisecond) // Exponential backoff
			continue
		}
		break // Non-retryable error
	}
	log.Printf("gRPC DeleteAccountByAccountNumber service: Failed to delete account after %d attempts: %v\n", maxRetries, err)
	return err
}

// use serializable isolation level for the transaction
func (s *AccountService) deleteAccountByAccountNumberTx(ctx context.Context, accountNumber int64, userID uuid.UUID) error {
	var (
		tx      *sql.Tx
		err     error
		account *model.Account
	)

	// Start a transaction
	tx, err = s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		log.Printf("gRPC DeleteAccountByAccountNumber service: Failed to begin transaction: %v\n", err)
		return model.ErrInternalServer
	}
	defer tx.Rollback()

	// Create a new repository with the transaction
	txRepo := s.repo.WithTx(tx)

	// First check if account belongs to user
	account, err = txRepo.GetAccountByAccountNumber(ctx, accountNumber)
	if err != nil {
		log.Printf("gRPC deleteAccountByAccountNumberTx service: Failed to get account: %v\n", err)
		if err == sql.ErrNoRows {
			return errors.New("account not found")
		}
		return model.ErrInternalServer
	}

	// Check ownership
	if account.UserID != userID {
		log.Printf("gRPC deleteAccountByAccountNumberTx service: Unauthorized deletion attempt for account number %v by user %v\n",
			accountNumber, userID)
		return model.ErrInternalServer
	}

	err = txRepo.DeleteAccountByAccountNumber(ctx, accountNumber)
	if err != nil {
		log.Printf("gRPC deleteAccountByAccountNumberTx service: Failed to delete account: %v\n", err)
		if err == sql.ErrNoRows {
			return errors.New("account not found")
		}
		return model.ErrInternalServer
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Printf("gRPC deleteAccountByAccountNumberTx service: Failed to commit transaction: %v\n", err)
		return model.ErrInternalServer
	}
	return nil
}

func (s *AccountService) CreateTransaction(ctx context.Context, transaction *model.Transaction, userID uuid.UUID) (*model.Transaction, error) {
	var (
		tx                 *sql.Tx
		err                error
		account            *model.Account
		createdTransaction *model.Transaction
	)

	// Start a transaction
	tx, err = s.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("gRPC CreateTransaction service: Failed to begin transaction: %v\n", err)
		return nil, model.ErrInternalServer
	}
	defer tx.Rollback()

	// Create a new repository with the transaction
	txRepo := s.repo.WithTx(tx)

	// First check if account belongs to user
	account, err = txRepo.GetAccountByID(ctx, transaction.AccountID)
	if err != nil {
		log.Printf("gRPC AddToAccountBalance service: Failed to get account: %v\n", err)
		if err == sql.ErrNoRows {
			return nil, errors.New("account not found")
		}
		return nil, model.ErrInternalServer
	}

	// Check ownership
	if account.UserID != userID {
		log.Printf("gRPC AddToAccountBalance service: Unauthorized balance modification attempt for account %v by user %v\n",
			account.AccountID, userID)
		return nil, model.ErrInternalServer
	}

	// Check if the transaction amount is valid
	if transaction.Amount == 0 {
		log.Printf("gRPC CreateTransaction service: Invalid transaction amount = 0\n")
		return nil, errors.New("invalid transaction amount")
	}

	// Create the transaction in the database
	createdTransaction, err = txRepo.CreateTransaction(ctx, transaction)
	if err != nil {
		log.Printf("gRPC CreateTransaction service: Failed to create transaction: %v\n", err)
		return nil, model.ErrInternalServer
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		log.Printf("gRPC CreateTransaction service: Failed to commit transaction: %v\n", err)
		return nil, model.ErrInternalServer
	}
	return createdTransaction, nil
}

// I think AddToAccountBalance is clearer than UpdateAccountBalance cause Update can mean "set" it to this amount instead of adding/substracting to it
// AddToAccountBalance adds the given amount (could be negative) to the account balance and retries on serialization failure
// with exponential backoff. It returns the updated account.
func (s *AccountService) AddToAccountBalance(ctx context.Context, accountNumber int64, amount int64, userID uuid.UUID) (*model.Account, error) {
	var (
		err error
		res *model.Account
	)
	for attempt := 0; attempt < maxRetries; attempt++ {
		res, err = s.addToAccountBalanceTx(ctx, accountNumber, amount, userID)
		if err == nil { // success
			return res, nil
		}
		// Check for serialization failure (Postgres error code 40001: https://www.postgresql.org/docs/current/mvcc-serialization-failure-handling.html)
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "40001" {
			log.Printf("Serialization failure, retrying AddToAccountBalance (attempt %d): %v", attempt+1, err)
			time.Sleep(time.Duration(attempt+1) * 100 * time.Millisecond) // Exponential backoff
			continue
		}
		break // Non-retryable error
	}
	log.Printf("gRPC AddToAccountBalance service: Failed to add to account balance after %d attempts: %v\n", maxRetries, err)
	return nil, err
}

func (s *AccountService) addToAccountBalanceTx(ctx context.Context, accountNumber int64, amount int64, userID uuid.UUID) (*model.Account, error) {
	var (
		tx             *sql.Tx
		err            error
		account        *model.Account
		updatedAccount *model.Account
	)

	// Start a transaction
	tx, err = s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		log.Printf("gRPC addToAccountBalanceTx service: Failed to begin transaction: %v\n", err)
		return nil, model.ErrInternalServer
	}
	defer tx.Rollback()

	// Create a new repository with the transaction
	txRepo := s.repo.WithTx(tx)

	// First check if account belongs to user
	account, err = txRepo.GetAccountByAccountNumber(ctx, accountNumber)
	if err != nil {
		log.Printf("gRPC addToAccountBalanceTx service: Failed to get account: %v\n", err)
		if err == sql.ErrNoRows {
			return nil, errors.New("account not found")
		}
		return nil, model.ErrInternalServer
	}

	// Check ownership
	if account.UserID != userID {
		log.Printf("gRPC addToAccountBalanceTx service: Unauthorized balance modification attempt for account %v by user %v\n",
			accountNumber, userID)
		return nil, model.ErrInternalServer
	}

	// Update the account balance in the database
	updatedAccount, err = txRepo.AddToAccountBalance(ctx, accountNumber, amount)
	if err != nil {
		log.Printf("gRPC addToAccountBalanceTx service: Failed to update balance: %v\n", err)
		if err == sql.ErrNoRows {
			return nil, errors.New("account not found")
		}
		return nil, model.ErrInternalServer
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		log.Printf("gRPC addToAccountBalanceTx service: Failed to commit transaction: %v\n", err)
		return nil, model.ErrInternalServer
	}
	return updatedAccount, nil
}

// we want repeatable read since we need to make sure that this account still exists after checking if the user is the owner
// since we're only reading data, we don't have to retry as per the Postgres documentation:
// "Note that only updating transactions might need to be retried; read-only transactions will never have serialization conflicts."
// https://www.postgresql.org/docs/current/transaction-iso.html#XACT-REPEATABLE-READ
func (s *AccountService) GetTransactionsByAccountID(ctx context.Context, accountID uuid.UUID, userID uuid.UUID) ([]*model.Transaction, error) {
	var (
		tx      *sql.Tx
		err     error
		account *model.Account
	)

	// Start a transaction
	tx, err = s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if err != nil {
		log.Printf("gRPC AddToAccountBalance service: Failed to begin transaction: %v\n", err)
		return nil, model.ErrInternalServer
	}
	defer tx.Rollback()

	// Create a new repository with the transaction
	txRepo := s.repo.WithTx(tx)

	// First verify user owns the account
	account, err = txRepo.GetAccountByID(ctx, accountID)
	if err != nil {
		log.Printf("gRPC GetTransactionsByAccountID service: Failed to get account: %v\n", err)
		if err == sql.ErrNoRows {
			return nil, errors.New("account not found")
		}
		return nil, model.ErrInternalServer
	}

	// Check ownership
	if account.UserID != userID {
		log.Printf("gRPC GetTransactionsByAccountID service: Unauthorized access attempt for account %v by user %v\n",
			accountID, userID)
		return nil, model.ErrInternalServer
	}

	transactions, err := txRepo.GetTransactionsByAccountID(ctx, accountID)
	if err != nil {
		log.Printf("gRPC GetTransactionsByAccountID service: Failed to get transactions: %v\n", err)
		if err == sql.ErrNoRows {
			return nil, errors.New("transaction not found")
		}
		return nil, model.ErrInternalServer
	}
	return transactions, nil
}

// Check if an account exists and belongs to the given user
func (s *AccountService) ValidateAccountNumber(ctx context.Context, accountNumber int64, userID uuid.UUID) (bool, error) {
	account, err := s.repo.GetAccountByAccountNumber(ctx, accountNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		log.Printf("gRPC ValidateAccount service: Failed to validate account: %v\n", err)
		return false, model.ErrInternalServer
	}

	// Check ownership
	if account.UserID != userID {
		return false, nil
	}

	return true, nil
}

// Function to check if account has sufficient balance and belongs to the user
func (s *AccountService) HasSufficientBalance(ctx context.Context, accountNumber int64, amount int64, userID uuid.UUID) (bool, error) {
	account, err := s.repo.GetAccountByAccountNumber(ctx, accountNumber)
	if err != nil {
		log.Printf("gRPC HasSufficientBalance service: Failed to get account: %v\n", err)
		if err == sql.ErrNoRows {
			return false, errors.New("account not found")
		}
		return false, model.ErrInternalServer
	}

	// Check ownership
	if account.UserID != userID {
		log.Printf("gRPC HasSufficientBalance service: Unauthorized balance check attempt for account %v by user %v\n",
			accountNumber, userID)
		return false, model.ErrInternalServer
	}

	return account.Balance >= amount, nil
}
