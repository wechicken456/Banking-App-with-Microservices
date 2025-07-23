package repository

import (
	"account/db/initialize"
	"account/model"
	"account/utils"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) (*sqlx.DB, func()) {
	db := initialize.ConnectDB()
	return db, func() {
		err := db.Close()
		require.NoError(t, err)
	}
}

// create multiple concurrent goroutines that create different accounts.
func TestCreateAccount_Success(t *testing.T) {
	t.Parallel()

	db, teardown := setupTestDB(t)
	defer teardown()
	repo := NewAccountRepository(db)

	numTests := 10
	results := make(chan *model.Account, numTests)
	errChan := make(chan error, numTests)

	for i := 0; i < numTests; i++ {
		go func() {
			tx, err := db.BeginTx(context.Background(), nil)
			if err != nil {
				errChan <- err
				return
			}
			defer tx.Rollback()

			txRepo := repo.WithTx(tx)
			user := utils.RandomUser()
			account, err := txRepo.CreateAccount(context.Background(), user)
			if err != nil {
				errChan <- err
				return
			}
			if err := tx.Commit(); err != nil {
				errChan <- err
				return
			}
			results <- account
			errChan <- nil
		}()
	}

	createdAccounts := make([]*model.Account, 0, numTests)
	for i := 0; i < numTests; i++ {
		err := <-errChan
		require.NoError(t, err)
		account := <-results
		require.NotEmpty(t, account)
		require.NotEqual(t, uuid.Nil, account.AccountID)
		require.NotZero(t, account.AccountNumber)
		createdAccounts = append(createdAccounts, account)
	}

	// Cleanup the accounts we created to test
	for _, account := range createdAccounts {
		err := repo.DeleteAccountByAccountNumber(context.Background(), account.AccountNumber)
		require.NoError(t, err)
	}
}

// creating an account with the same account number should fail
func TestCreateAccount_Fail(t *testing.T) {
	t.Parallel()

	db, teardown := setupTestDB(t)
	defer teardown()
	repo := NewAccountRepository(db)

	account := utils.RandomAccount()

	_, err := repo.queries.CreateAccount(context.Background(), *convertToCreateAccountParams(account))
	require.NoError(t, err)

	// create the same account again

	_, err = repo.queries.CreateAccount(context.Background(), *convertToCreateAccountParams(account))
	require.Error(t, err)

	// Cleanup the account we created to test
	// Cleanup the accounts we created to test
	err = repo.DeleteAccountByAccountNumber(context.Background(), account.AccountNumber)
	require.NoError(t, err)
}

func TestGetAccountByAccountNumber_Success(t *testing.T) {
	t.Parallel()

	db, teardown := setupTestDB(t)
	defer teardown()

	repo := NewAccountRepository(db)
	tx, err := db.BeginTx(context.Background(), nil)
	require.NoError(t, err)
	defer tx.Rollback()
	txRepo := repo.WithTx(tx)

	var createdAccount *model.Account
	user := utils.RandomUser()
	createdAccount, err = txRepo.CreateAccount(context.Background(), user)
	require.NoError(t, err)

	retrievedAccount, err := txRepo.GetAccountByAccountNumber(context.Background(), createdAccount.AccountNumber)
	require.NoError(t, err)
	require.Equal(t, createdAccount, retrievedAccount)
}

func TestAddToAccountBalance_Success(t *testing.T) {
	t.Parallel()

	db, teardown := setupTestDB(t)
	defer teardown()
	repo := NewAccountRepository(db)

	var createdAccount *model.Account
	tx, err := db.BeginTx(context.Background(), nil)
	require.NoError(t, err)
	defer tx.Rollback()
	txRepo := repo.WithTx(tx)
	user := utils.RandomUser()
	createdAccount, err = txRepo.CreateAccount(context.Background(), user)
	require.NoError(t, err)
	err = tx.Commit()
	require.NoError(t, err)

	addAmount := int64(1000)
	numTests := 10
	errChan := make(chan error, numTests)
	results := make(chan *model.Account, numTests)

	// create concurrent goroutines that add to the account balance
	for i := 0; i < numTests; i++ {
		go func() {
			tx, err := db.BeginTx(context.Background(), nil)
			if err != nil {
				errChan <- err
				return
			}
			defer tx.Rollback()
			txRepo := repo.WithTx(tx)
			account, err := txRepo.AddToAccountBalance(context.Background(), createdAccount.AccountNumber, addAmount)
			if err != nil {
				errChan <- err
				return
			}
			if err := tx.Commit(); err != nil {
				errChan <- err
				return
			}
			results <- account
			errChan <- nil
		}()
	}

	for i := 0; i < numTests; i++ {
		err := <-errChan
		require.NoError(t, err)
		account := <-results
		require.NotEmpty(t, account)
	}

	// check if the balance is updated correctly
	finalAccount, err := repo.GetAccountByAccountNumber(context.Background(), createdAccount.AccountNumber)
	require.NoError(t, err)
	expectedBalance := createdAccount.Balance + (addAmount * int64(numTests))
	require.Equal(t, expectedBalance, finalAccount.Balance)

	// Cleanup the account we created to test
	err = repo.DeleteAccountByAccountNumber(context.Background(), createdAccount.AccountNumber)
	require.NoError(t, err)
}

func TestDeleteAccountByAccountNumber_Success(t *testing.T) {
	t.Parallel()

	db, teardown := setupTestDB(t)
	defer teardown()
	repo := NewAccountRepository(db)

	var createdAccount *model.Account
	tx, err := db.BeginTx(context.Background(), nil)
	require.NoError(t, err)
	defer tx.Rollback()
	txRepo := repo.WithTx(tx)

	user := utils.RandomUser()
	createdAccount, err = txRepo.CreateAccount(context.Background(), user)
	require.NoError(t, err)

	err = txRepo.queries.DeleteAccountByAccountNumber(context.Background(), int64(createdAccount.AccountNumber))
	require.NoError(t, err)

	res, err := repo.GetAccountByAccountNumber(context.Background(), createdAccount.AccountNumber)
	require.Error(t, err)
	require.Nil(t, res)
}

// Create 1 single transaction
func TestCreateTransaction_Success(t *testing.T) {
	t.Parallel()

	db, teardown := setupTestDB(t)
	defer teardown()
	repo := NewAccountRepository(db)

	var createdAccount *model.Account
	tx, err := db.BeginTx(context.Background(), nil)
	require.NoError(t, err)
	defer tx.Rollback()
	txRepo := repo.WithTx(tx)

	user := utils.RandomUser()
	createdAccount, err = txRepo.CreateAccount(context.Background(), user)
	require.NoError(t, err)

	transaction := utils.RandomTransaction()
	transaction.AccountID = createdAccount.AccountID
	createdTransaction, err := txRepo.CreateTransaction(context.Background(), transaction)
	require.NoError(t, err)

	require.NotEmpty(t, createdTransaction)
	require.Equal(t, transaction.AccountID, createdTransaction.AccountID)
}

// Multiple concurrent goroutines that create different transactions should update the account balance correctly
func TestCreateMultipleTransactions_Success(t *testing.T) {
	t.Parallel()

	db, teardown := setupTestDB(t)
	defer teardown()
	repo := NewAccountRepository(db)

	var createdAccount *model.Account
	tx, err := db.BeginTx(context.Background(), nil)
	require.NoError(t, err)
	defer tx.Rollback()
	txRepo := repo.WithTx(tx)
	user := utils.RandomUser()
	createdAccount, err = txRepo.CreateAccount(context.Background(), user)
	require.NoError(t, err)
	err = tx.Commit()
	require.NoError(t, err)

	balance := createdAccount.Balance
	numTransactions := 50
	errChan := make(chan error, numTransactions)
	results := make(chan *model.Transaction, numTransactions)

	for i := 0; i < numTransactions; i++ {
		go func() {
			tx, err := db.BeginTx(context.Background(), nil)
			if err != nil {
				errChan <- err
				return
			}
			defer tx.Rollback()
			txRepo := repo.WithTx(tx)
			transaction := utils.RandomTransaction()
			transaction.AccountID = createdAccount.AccountID
			result, err := txRepo.CreateTransaction(context.Background(), transaction)
			if err != nil {
				errChan <- err
				return
			}
			_, err = txRepo.AddToAccountBalance(context.Background(), createdAccount.AccountNumber, transaction.Amount)
			if err != nil {
				errChan <- err
				return
			}
			if err := tx.Commit(); err != nil {
				errChan <- err
				return
			}
			results <- result
			errChan <- nil
		}()
	}

	createdTransactions := make([]*model.Transaction, 0, numTransactions)
	for i := 0; i < numTransactions; i++ {
		err := <-errChan
		require.NoError(t, err)
		result := <-results
		require.NotEmpty(t, result)
		balance += result.Amount
		createdTransactions = append(createdTransactions, result)
	}

	// check if the balance is updated correctly
	finalAccount, err := repo.GetAccountByAccountNumber(context.Background(), createdAccount.AccountNumber)
	require.NoError(t, err)
	require.Equal(t, balance, finalAccount.Balance)

	// Cleanup the account and transactions we created to test
	tx, err = db.BeginTx(context.Background(), nil)
	require.NoError(t, err)
	defer tx.Rollback()
	txRepo = repo.WithTx(tx)
	for _, tx := range createdTransactions {
		err := txRepo.queries.DeleteTransactionByID(context.Background(), tx.TransactionID)
		if err != nil {
			return
		}
	}
	err = txRepo.queries.DeleteAccountByAccountNumber(context.Background(), int64(createdAccount.AccountNumber))
	require.NoError(t, err)
	err = tx.Commit()
	require.NoError(t, err)
}

func TestCreateTransactionWithTransferID_Success(t *testing.T) {
	t.Parallel()

	db, teardown := setupTestDB(t)
	defer teardown()
	repo := NewAccountRepository(db)

	var createdAccount *model.Account
	tx, err := db.BeginTx(context.Background(), nil)
	require.NoError(t, err)
	defer tx.Rollback()
	txRepo := repo.WithTx(tx)
	user := utils.RandomUser()
	createdAccount, err = txRepo.CreateAccount(context.Background(), user)
	require.NoError(t, err)

	transferID := uuid.New()
	transaction := &model.Transaction{
		TransactionID:   uuid.New(),
		AccountID:       createdAccount.AccountID,
		Amount:          100,
		TransactionType: "TRANSFER_CREDIT",
		Status:          "COMPLETED",
		TransferID:      uuid.NullUUID{UUID: transferID, Valid: true},
	}

	var createdTransaction *model.Transaction
	createdTransaction, err = txRepo.CreateTransaction(context.Background(), transaction)
	require.NoError(t, err)
	require.NotEmpty(t, createdTransaction)
	require.Equal(t, transaction, createdTransaction)
}
