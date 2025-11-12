package service

import (
	"account/internal/cache"
	"account/model"
	"account/repository"
	"context"
	"database/sql"
	"encoding/json"
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

// userID is the ID of the user who initiated the request
func (s *AccountService) CreateAccount(ctx context.Context, user *model.User, idempotencyKey string, userID uuid.UUID) (*model.Account, error) {
	// Check if the user ID in the request matches the user ID in the context
	if userID != user.UserID {
		log.Printf("CreateAccount: User ID mismatch: %v != %v\n", userID, user.UserID)
		return nil, model.ErrNotAuthorized
	}

	var (
		attempt int
		backoff int
		res     *model.Account
		err     error
	)

	backoff = 2

	for attempt = range maxRetries {
		res, err = s.createAccountTx(ctx, user, idempotencyKey, userID)
		if err == nil {
			return res, nil
		}
		// Check for serialization failure (Postgres error code 40001: https://www.postgresql.org/docs/current/mvcc-serialization-failure-handling.html)
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "40001" {
			log.Printf("Serialization failure, retrying CreateAccount (attempt %d): %v", attempt+1, err)
			time.Sleep(time.Duration(backoff) * 100 * time.Millisecond) // Exponential backoff
			backoff *= 2
			continue
		}
		break // Non-retryable error
	}
	log.Printf("CreateAccount: Failed to create account after %d attempts: %v\n", attempt+1, err)
	return nil, err
}

// userID is the ID of the user who initiated the request
func (s *AccountService) createAccountTx(ctx context.Context, user *model.User, idempotencyKey string, userID uuid.UUID) (*model.Account, error) {
	var (
		tx             *sql.Tx
		err            error
		createdAccount *model.Account
	)

	// Start a transaction
	tx, err = s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		log.Printf("createAccountTx: Failed to begin transaction: %v\n", err)
		return nil, model.ErrInternalServer
	}
	defer tx.Rollback()

	// Create a new repository with the transaction
	txRepo := s.repo.WithTx(tx)

	// Try to insert idempotency key with status "PENDING".
	// the statement will block if another concurrent transactional already to inserts the same key, even if it hasn't committed yet.
	key, err := txRepo.GetOrClaimIdempotencyKey(ctx, &model.IdempotencyKey{
		KeyID:  idempotencyKey,
		UserID: userID,
		Status: "PENDING",
	})
	if err == nil {
		if key.Status != "PENDING" { // "PENDING" implies that we (the current transaction) is the first one to create the idempotency key. Otherwise, we blocked while another transaction inserted the same key.
			log.Printf("createAccountTx: idempotency key already exists: %v\n", err)
			cachedAccount := &model.Account{}
			err := json.Unmarshal([]byte(key.ResponseMessage), cachedAccount)
			if err != nil {
				log.Printf("createAccountTx: Failed to unmarshal account: %v\n", err)
				return nil, model.ErrInternalServer
			}
			return cachedAccount, nil
		}
	} else {
		log.Printf("createAccountTx: Failed to get idempotency key: %v\n", err)
		return nil, model.ErrInternalServer
	}

	createdAccount, err = txRepo.CreateAccount(ctx, user)
	if err != nil {
		log.Printf("createAccountTx: Failed to create account: %v\n", err)
		return nil, model.ErrInternalServer
	}

	// Update the idempotency key status
	key.Status = "COMPLETED"
	marshalled, err := json.Marshal(createdAccount)
	if err != nil {
		log.Printf("createAccountTx: Failed to marshal account: %v\n", err)
		return nil, model.ErrInternalServer
	}
	key.ResponseMessage = string(marshalled)

	if _, err = txRepo.UpdateIdempotencyKey(ctx, key); err != nil {
		log.Printf("createAccountTx: Failed to update idempotency key: %v\n", err)
		return nil, model.ErrInternalServer
	}

	if err = tx.Commit(); err != nil {
		log.Printf("createAccountTx: Failed to commit transaction: %v\n", err)
		return nil, model.ErrInternalServer
	}
	return createdAccount, nil
}

// userID is the ID of the user who initiated the request
func (s *AccountService) GetAccount(ctx context.Context, accountID uuid.UUID, userID uuid.UUID) (*model.Account, error) {
	var err error

	cachedAcct, err := cache.Get(ctx, accountID)
	if err == nil {
		if cachedAcct.UserID != userID {
			return nil, model.ErrNotAuthorized
		}
		log.Printf("\n\nCache hit!\n\n")
		return cachedAcct, nil
	}

	if errors.Is(err, model.ErrCacheMiss) {
		log.Printf("Cache miss for accountID: %s", accountID)
	} else {
		log.Printf("Error hitting cache: %s", err)
	}

	account, err := s.repo.GetAccountByID(ctx, accountID)
	if err != nil {
		log.Printf("GetAccount: Failed to get accounts: %v\n", err)
		if err == sql.ErrNoRows {
			return nil, model.ErrInvalidArgument
		}
		return nil, model.ErrInternalServer
	}

	// Check if the user owns the account
	if account.UserID != userID {
		log.Printf("GetAccount: Unauthorized access attempt for account id %v by user %v\n", accountID, userID)
		return nil, model.ErrNotAuthorized
	}

	time.Sleep(50 * time.Millisecond) // sleep to simulate high network latency during benchmark

	err = cache.Set(ctx, account)
	log.Printf("\n\nCache miss. Populating cache: %v\n", err)
	if err != nil {
		log.Printf("Error populating cache: %v", err)
	}
	return account, nil
}

// userID is the ID of the user who initiated the request
func (s *AccountService) GetAccountsByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Account, error) {
	accounts, err := s.repo.GetAccountsByUserID(ctx, userID)
	if err != nil {
		log.Printf("GetAccountsByUserID: Failed to get accounts: %v\n", err)
		return nil, model.ErrInternalServer
	}
	return accounts, nil
}

func (s *AccountService) GetAccountByAccountNumber(ctx context.Context, accountNumber int32, userID uuid.UUID) (*model.Account, error) {
	account, err := s.repo.GetAccountByAccountNumber(ctx, accountNumber)
	if err != nil {
		log.Printf("GetAccountByAccountNumber: Failed to get account: %v\n", err)
		if err == sql.ErrNoRows {
			return nil, model.ErrInvalidArgument
		}
		return nil, model.ErrInternalServer
	}

	// Check if the user owns the account
	if account.UserID != userID {
		log.Printf("GetAccountByAccountNumber: Unauthorized access attempt for account number %v by user %v\n",
			accountNumber, userID)
		return nil, model.ErrNotAuthorized
	}

	return account, nil
}

// delete account by account number with exponential backoff retries
// userID is the ID of the user who initiated the request
func (s *AccountService) DeleteAccountByAccountNumber(ctx context.Context, accountNumber int32, idempotencyKey string, userID uuid.UUID) error {
	var (
		attempt int
		backoff int
		err     error
	)

	backoff = 2

	for attempt := range maxRetries {
		err = s.deleteAccountByAccountNumberTx(ctx, accountNumber, idempotencyKey, userID)
		if err == nil {
			return nil
		}
		// Check for serialization failure (Postgres error code 40001: https://www.postgresql.org/docs/current/mvcc-serialization-failure-handling.html)
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "40001" {
			log.Printf("Serialization failure, retrying DeleteAccountByAccountNumber (attempt %d): %v", attempt+1, err)
			time.Sleep(time.Duration(backoff) * 100 * time.Millisecond) // Exponential backoff
			backoff *= 2
			continue
		}
		break // Non-retryable error
	}
	log.Printf("DeleteAccountByAccountNumber: Failed to delete account after %d attempts: %v\n", attempt+1, err)
	return err
}

// use serializable isolation level for the transaction
// userID is the ID of the user who initiated the request
func (s *AccountService) deleteAccountByAccountNumberTx(ctx context.Context, accountNumber int32, idempotencyKey string, userID uuid.UUID) error {
	var (
		tx      *sql.Tx
		err     error
		account *model.Account
	)

	// Start a transaction
	tx, err = s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		log.Printf("deleteAccountByAccountNumberTx: Failed to begin transaction: %v\n", err)
		return model.ErrInternalServer
	}
	defer tx.Rollback()

	// Create a new repository with the transaction
	txRepo := s.repo.WithTx(tx)

	// First check if account belongs to user
	account, err = txRepo.GetAccountByAccountNumber(ctx, accountNumber)
	if err != nil {
		log.Printf("deleteAccountByAccountNumberTx: Failed to get account: %v\n", err)
		if err == sql.ErrNoRows {
			return model.ErrInvalidArgument
		}
		return model.ErrInternalServer
	}

	// Check ownership
	if account.UserID != userID {
		log.Printf("deleteAccountByAccountNumberTx: Unauthorized deletion attempt for account number %v by user %v\n",
			accountNumber, userID)
		return model.ErrNotAuthorized
	}

	key, err := txRepo.GetOrClaimIdempotencyKey(ctx, &model.IdempotencyKey{
		KeyID:  idempotencyKey,
		UserID: userID,
		Status: "PENDING",
	})
	if err == nil {
		if key.Status != "PENDING" { // "PENDING" implies that we (the current transaction) is the first one to create the idempotency key. Otherwise, we blocked while another transaction inserted the same key.
			log.Printf("deleteAccountByAccountNumberTx: idempotency key already exists: %v\n", err)
			return nil
		}
	} else if err != sql.ErrNoRows {
		log.Printf("deleteAccountByAccountNumberTx: Failed to get idempotency key: %v\n", err)
		return model.ErrInternalServer
	}

	// Delete the account in the database
	err = txRepo.DeleteAccountByAccountNumber(ctx, accountNumber)
	if err != nil {
		log.Printf("deleteAccountByAccountNumberTx: Failed to delete account: %v\n", err)
		if err == sql.ErrNoRows {
			return model.ErrInvalidArgument
		}
		return model.ErrInternalServer
	}

	// Update the idempotency key
	key.Status = "COMPLETED"
	key.ResponseMessage = string("success")
	if _, err = txRepo.UpdateIdempotencyKey(ctx, key); err != nil {
		log.Printf("deleteAccountByAccountNumberTx: Failed to update idempotency key: %v\n", err)
		return model.ErrInternalServer
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Printf("deleteAccountByAccountNumberTx: Failed to commit transaction: %v\n", err)
		return model.ErrInternalServer
	}
	go cache.Invalidate(ctx, account.AccountID)
	return nil
}

// delete the idempotency key given its ID with exponential backoff retries
// should be used internally only so we don't need to check for ownership
func (s *AccountService) DeleteIdempotencyKeyByID(ctx context.Context, idempotencyKey string) error {
	var (
		attempt int
		backoff int
		err     error
	)

	backoff = 2

	for attempt := range maxRetries {
		err = s.deleteIdempotencyKeyByIDTx(ctx, idempotencyKey)
		if err == nil {
			return nil
		}
		// Check for serialization failure (Postgres error code 40001: https://www.postgresql.org/docs/current/mvcc-serialization-failure-handling.html)
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "40001" {
			log.Printf("Serialization failure, retrying DeleteIdempotencyKeyByID (attempt %d): %v", attempt+1, err)
			time.Sleep(time.Duration(backoff) * 100 * time.Millisecond) // Exponential backoff
			backoff *= 2
			continue
		}
		break // Non-retryable error
	}
	log.Printf("DeleteIdempotencyKeyByID: Failed to delete idempotency key after %d attempts: %v\n", attempt+1, err)
	return err
}

// use serializable isolation level for the transaction
// userID is the ID of the user who initiated the request
func (s *AccountService) deleteIdempotencyKeyByIDTx(ctx context.Context, idempotencyKey string) error {
	var (
		tx  *sql.Tx
		err error
	)

	// Start a transaction
	tx, err = s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		log.Printf("deleteIdempotencyKeyByIDTx: Failed to begin transaction: %v\n", err)
		return model.ErrInternalServer
	}
	defer tx.Rollback()

	// Create a new repository with the transaction
	txRepo := s.repo.WithTx(tx)

	err = txRepo.DeleteIdempotencyKeyByID(ctx, idempotencyKey)
	if err != nil {
		log.Printf("deleteIdempotencyKeyByIDTx: Failed to delete idempotency key: %v\n", err)
		return model.ErrInternalServer
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Printf("deleteIdempotencyKeyByIDTx: Failed to commit transaction: %v\n", err)
		return model.ErrInternalServer
	}
	return nil
}

// create transaction with exponential backoff retries
// userID is the ID of the user who initiated the request
func (s *AccountService) CreateTransaction(ctx context.Context, transaction *model.Transaction, idempotencyKey string, userID uuid.UUID) (*model.Transaction, error) {
	var (
		attempt int
		backoff int
		res     *model.Transaction
		err     error
	)

	backoff = 2

	for attempt = range maxRetries {
		res, err = s.createTransactionTx(ctx, transaction, idempotencyKey, userID)
		if err == nil {
			// invalidate cache
			go cache.Invalidate(ctx, transaction.AccountID)
			return res, nil
		}
		// Check for serialization failure (Postgres error code 40001: https://www.postgresql.org/docs/current/mvcc-serialization-failure-handling.html)
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "40001" {
			log.Printf("Serialization failure, retrying CreateTransaction (attempt %d): %v", attempt+1, err)
			time.Sleep(time.Duration(backoff) * 100 * time.Millisecond) // Exponential backoff
			backoff *= 2
			continue
		}
		break // Non-retryable error
	}
	log.Printf("CreateTransaction: Failed to create transaction after %d attempts: %v\n", attempt+1, err)
	return nil, err
}

// userID is the ID of the user who initiated the request
func (s *AccountService) createTransactionTx(ctx context.Context, transaction *model.Transaction, idempotencyKey string, userID uuid.UUID) (*model.Transaction, error) {
	var (
		tx                 *sql.Tx
		err                error
		account            *model.Account
		createdTransaction *model.Transaction
	)

	// Start a transaction
	tx, err = s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		log.Printf("createTransactionTx: Failed to begin transaction: %v\n", err)
		return nil, model.ErrInternalServer
	}
	defer tx.Rollback()

	// Create a new repository with the transaction
	txRepo := s.repo.WithTx(tx)

	// First check if account belongs to user
	account, err = txRepo.GetAccountByID(ctx, transaction.AccountID)
	if err != nil {
		log.Printf("createTransactionTx: Failed to get account: %v\n", err)
		if err == sql.ErrNoRows {
			return nil, model.ErrInvalidArgument
		}
		return nil, model.ErrInternalServer
	}

	// Check ownership
	if account.UserID != userID {
		log.Printf("createTransactionTx: Unauthorized balance modification attempt for account %v by user %v\n",
			account.AccountID, userID)
		return nil, model.ErrNotAuthorized
	}

	// Try to insert idempotency key with status "PENDING".
	// the statement will block if another concurrent transactional already to inserts the same key, even if it hasn't committed yet.
	key, err := txRepo.GetOrClaimIdempotencyKey(ctx, &model.IdempotencyKey{
		KeyID:  idempotencyKey,
		UserID: userID,
		Status: "PENDING",
	})
	if err == nil {
		if key.Status != "PENDING" { // "PENDING" implies that we (the current transaction) is the first one to create the idempotency key. Otherwise, we blocked while another transaction inserted the same key.
			log.Printf("createTransactionTx: idempotency key already exists: %v\n", key)
			cachedTransaction := &model.Transaction{}
			err := json.Unmarshal([]byte(key.ResponseMessage), cachedTransaction)
			if err != nil {
				log.Printf("createTransactionTx: Failed to unmarshal transaction: %v\n", err)
				return nil, model.ErrInternalServer
			}
			return cachedTransaction, nil
		}
	} else {
		log.Printf("createTransactionTx: Failed to get idempotency key: %v\n", err)
		return nil, model.ErrInternalServer
	}

	// Check if the transaction amount is valid
	if transaction.Amount == 0 {
		log.Printf("createTransactionTx: Invalid transaction amount = 0\n")
		return nil, model.ErrInvalidArgument
	}

	// Create the transaction in the database
	createdTransaction, err = txRepo.CreateTransaction(ctx, transaction)
	if err != nil {
		log.Printf("createTransactionTx: Failed to create transaction: %v\n", err)
		return nil, model.ErrInternalServer
	}

	// Update the account balance in the database
	_, err = txRepo.AddToAccountBalance(ctx, account.AccountNumber, transaction.Amount)
	if err != nil {
		log.Printf("createTransactionTx: Failed to update balance: %v\n", err)
		if err == sql.ErrNoRows {
			return nil, model.ErrInvalidArgument
		}
		return nil, model.ErrInternalServer
	}

	// Update the idempotency key status
	key.Status = "COMPLETED"
	marshalled, err := json.Marshal(createdTransaction)
	if err != nil {
		log.Printf("createTransactionTx: Failed to marshal transaction: %v\n", err)
		return nil, model.ErrInternalServer
	}
	key.ResponseMessage = string(marshalled)
	if _, err = txRepo.UpdateIdempotencyKey(ctx, key); err != nil {
		log.Printf("createTransactionTx: Failed to update idempotency key: %v\n", err)
		return nil, model.ErrInternalServer
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		log.Printf("createTransactionTx: Failed to commit transaction: %v\n", err)
		return nil, model.ErrInternalServer
	}

	return createdTransaction, nil
}

// // I think AddToAccountBalance is clearer than UpdateAccountBalance cause Update can mean "set" it to this amount instead of adding/substracting to it
// // AddToAccountBalance adds the given amount (could be negative) to the account balance and retries on serialization failure
// // with exponential backoff. It returns the updated account.
// // userID is the ID of the user who initiated the request
// func (s *AccountService) AddToAccountBalance(ctx context.Context, accountNumber int32, amount int64, idempotencyKey string, userID uuid.UUID) (*model.Account, error) {
// 	var (
// 		err error
// 		res *model.Account
// 	)
// 	for attempt := 0; attempt < maxRetries; attempt++ {
// 		res, err = s.addToAccountBalanceTx(ctx, accountNumber, amount, userID)
// 		if err == nil { // success
// 			return res, nil
// 		}
// 		// Check for serialization failure (Postgres error code 40001: https://www.postgresql.org/docs/current/mvcc-serialization-failure-handling.html)
// 		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "40001" {
// 			log.Printf("Serialization failure, retrying AddToAccountBalance (attempt %d): %v", attempt+1, err)
// 			time.Sleep(time.Duration(attempt+1) * 100 * time.Millisecond) // Exponential backoff
// 			continue
// 		}
// 		break // Non-retryable error
// 	}
// 	log.Printf("AddToAccountBalance: Failed to add to account balance after %d attempts: %v\n", maxRetries, err)
// 	return nil, err
// }

// // userID is the ID of the user who initiated the request
// func (s *AccountService) addToAccountBalanceTx(ctx context.Context, accountNumber int32, amount int64, idempotencyKey string, userID uuid.UUID) (*model.Account, error) {
// 	var (
// 		tx             *sql.Tx
// 		err            error
// 		account        *model.Account
// 		updatedAccount *model.Account
// 	)

// 	// Start a transaction
// 	tx, err = s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
// 	if err != nil {
// 		log.Printf("addToAccountBalanceTx: Failed to begin transaction: %v\n", err)
// 		return nil, model.ErrInternalServer
// 	}
// 	defer tx.Rollback()

// 	// Create a new repository with the transaction
// 	txRepo := s.repo.WithTx(tx)

// 	// First check if account belongs to user
// 	account, err = txRepo.GetAccountByAccountNumber(ctx, accountNumber)
// 	if err != nil {
// 		log.Printf("addToAccountBalanceTx: Failed to get account: %v\n", err)
// 		if err == sql.ErrNoRows {
// 			return nil, errors.New("account not found")
// 		}
// 		return nil, model.ErrInternalServer
// 	}

// 	// Check ownership
// 	if account.UserID != userID {
// 		log.Printf("addToAccountBalanceTx: Unauthorized balance modification attempt for account %v by user %v\n",
// 			accountNumber, userID)
// 		return nil, model.ErrInternalServer
// 	}

// 	// Check Idempotency key
// 	key, err := txRepo.GetIdempotencyKey(ctx, idempotencyKey)
// 	if err != nil {
// 		log.Printf("addToAccountBalanceTx: Failed to get idempotency key: %v\n", err)
// 		return nil, model.ErrInternalServer
// 	}
// 	if key != nil {
// 		log.Printf("addToAccountBalanceTx: Idempotency key %v already exists\n", idempotencyKey)
// 		return nil, model.ErrIdempotencyKeyExists
// 	}

// 	// Update the account balance in the database
// 	updatedAccount, err = txRepo.AddToAccountBalance(ctx, accountNumber, amount)
// 	if err != nil {
// 		log.Printf("addToAccountBalanceTx: Failed to update balance: %v\n", err)
// 		if err == sql.ErrNoRows {
// 			return nil, errors.New("account not found")
// 		}
// 		return nil, model.ErrInternalServer
// 	}

// 	// Commit transaction
// 	if err = tx.Commit(); err != nil {
// 		log.Printf("addToAccountBalanceTx: Failed to commit transaction: %v\n", err)
// 		return nil, model.ErrInternalServer
// 	}
// 	return updatedAccount, nil
// }

// we want repeatable read since we need to make sure that this account still exists after checking if the user is the owner
// since we're only reading data, we don't have to retry as per the Postgres documentation:
// "Note that only updating transactions might need to be retried; read-only transactions will never have serialization conflicts."
// https://www.postgresql.org/docs/current/transaction-iso.html#XACT-REPEATABLE-READ
// userID is the ID of the user who initiated the request
func (s *AccountService) GetTransactionsByAccountID(ctx context.Context, accountID uuid.UUID, userID uuid.UUID) ([]*model.Transaction, error) {
	var (
		tx      *sql.Tx
		err     error
		account *model.Account
	)

	// Start a transaction
	tx, err = s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if err != nil {
		log.Printf("GetTransactionsByAccountID: Failed to begin transaction: %v\n", err)
		return nil, model.ErrInternalServer
	}
	defer tx.Rollback()

	// Create a new repository with the transaction
	txRepo := s.repo.WithTx(tx)

	// First verify user owns the account
	account, err = txRepo.GetAccountByID(ctx, accountID)
	if err != nil {
		log.Printf("GetTransactionsByAccountID: Failed to get account: %v\n", err)
		if err == sql.ErrNoRows {
			return nil, model.ErrInvalidArgument
		}
		return nil, model.ErrInternalServer
	}

	// Check ownership
	if account.UserID != userID {
		log.Printf("GetTransactionsByAccountID: Unauthorized access attempt for account %v by user %v\n",
			accountID, userID)
		return nil, model.ErrNotAuthorized
	}

	transactions, err := txRepo.GetTransactionsByAccountID(ctx, accountID)
	if err != nil {
		log.Printf("GetTransactionsByAccountID: Failed to get transactions: %v\n", err)
		if err == sql.ErrNoRows {
			return nil, model.ErrInvalidArgument
		}
		return nil, model.ErrInternalServer
	}
	return transactions, nil
}

// Check if an account exists and belongs to the given user
// userID is the ID of the user who initiated the request
func (s *AccountService) ValidateAccountNumber(ctx context.Context, accountNumber int32, userID uuid.UUID) (bool, error) {
	account, err := s.repo.GetAccountByAccountNumber(ctx, accountNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		log.Printf("ValidateAccountNumber: Failed to validate account: %v\n", err)
		return false, model.ErrInternalServer
	}

	// Check ownership
	if account.UserID != userID {
		return false, nil
	}

	return true, nil
}

// Function to check if account has sufficient balance and belongs to the user
// userID is the ID of the user who initiated the request
func (s *AccountService) HasSufficientBalance(ctx context.Context, accountNumber int32, amount int64, userID uuid.UUID) (bool, error) {
	account, err := s.repo.GetAccountByAccountNumber(ctx, accountNumber)
	if err != nil {
		log.Printf("HasSufficientBalance: Failed to get account: %v\n", err)
		if err == sql.ErrNoRows {
			return false, model.ErrInvalidArgument
		}
		return false, model.ErrInternalServer
	}

	// Check ownership
	if account.UserID != userID {
		log.Printf("HasSufficientBalance: Unauthorized balance check attempt for account %v by user %v\n",
			accountNumber, userID)
		return false, model.ErrNotAuthorized
	}

	return account.Balance >= amount, nil
}
