package service

import (
	"account/db/initialize"
	"account/model"
	"account/repository"
	"account/utils"
	"context"
	"os"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

var db *sqlx.DB = nil
var service *AccountService = nil

// main will run before all tests
func TestMain(m *testing.M) {
	err := godotenv.Load("../.env")
	if err != nil {
		panic("Error loading .env file")
	}

	db = initialize.ConnectDB()
	repo := repository.NewAccountRepository(db)
	service = NewAccountService(repo, db)

	// m.Run() runs all the tests. We can add cleanup code before os.Exit if we want
	os.Exit(m.Run())
}

// create multiple concurrent goroutines that create different accounts with different idempotency keys
func TestCreateAccount_Success(t *testing.T) {

	numTests := 10
	results := make(chan *model.Account, numTests)
	errChan := make(chan error, numTests)
	keyChan := make(chan uuid.UUID, numTests)

	for i := 0; i < numTests; i++ {
		go func() {

			key := utils.RandomIdempotencyKey()
			user := utils.RandomUser()
			account, err := service.CreateAccount(context.Background(), user, key, user.UserID)

			if err != nil {
				errChan <- err
				return
			}
			results <- account
			keyChan <- key
			errChan <- nil
		}()
	}

	for i := 0; i < numTests; i++ {
		err := <-errChan
		require.NoError(t, err)
		account := <-results
		key := <-keyChan
		require.NotEmpty(t, account)
		require.NotEqual(t, uuid.Nil, account.AccountID)
		require.NotZero(t, account.AccountNumber)

		// Cleanup the accounts we created to test
		_key := utils.RandomIdempotencyKey()
		err = service.DeleteAccountByAccountNumber(context.Background(), account.AccountNumber, _key, account.UserID)
		require.NoError(t, err)
		err = service.DeleteIdempotencyKeyByID(context.Background(), _key)
		require.NoError(t, err)
		err = service.DeleteIdempotencyKeyByID(context.Background(), key)
		require.NoError(t, err)

	}
}

func TestGetAccountByAccountNumber_Success(t *testing.T) {
	user := utils.RandomUser()
	key := utils.RandomIdempotencyKey()
	createdAccount, err := service.CreateAccount(context.Background(), user, key, user.UserID)
	require.NoError(t, err)

	// delete the idempotency key we used to create the account
	err = service.DeleteIdempotencyKeyByID(context.Background(), key)
	require.NoError(t, err)

	retrievedAccount, err := service.GetAccountByAccountNumber(context.Background(), createdAccount.AccountNumber, user.UserID)
	require.NoError(t, err)
	require.Equal(t, createdAccount, retrievedAccount)

	// Cleanup the account we created to test
	key = utils.RandomIdempotencyKey()
	err = service.DeleteAccountByAccountNumber(context.Background(), createdAccount.AccountNumber, key, user.UserID)
	require.NoError(t, err)

	err = service.DeleteIdempotencyKeyByID(context.Background(), key)
	require.NoError(t, err)
}

func TestDeleteAccountByAccountNumber_Success(t *testing.T) {

	user := utils.RandomUser()
	key := utils.RandomIdempotencyKey()

	createdAccount, err := service.CreateAccount(context.Background(), user, key, user.UserID)
	require.NoError(t, err)

	// delete the idempotency key we used to create the account
	err = service.DeleteIdempotencyKeyByID(context.Background(), key)
	require.NoError(t, err)

	key = utils.RandomIdempotencyKey()
	err = service.DeleteAccountByAccountNumber(context.Background(), createdAccount.AccountNumber, key, user.UserID)
	require.NoError(t, err)

	err = service.DeleteIdempotencyKeyByID(context.Background(), key)
	require.NoError(t, err)

	res, err := service.GetAccountByAccountNumber(context.Background(), createdAccount.AccountNumber, user.UserID)
	require.Error(t, err)
	require.Nil(t, res)

}

// Create 1 single transaction
func TestCreateTransaction_Success(t *testing.T) {

	key := utils.RandomIdempotencyKey()
	user := utils.RandomUser()
	createdAccount, err := service.CreateAccount(context.Background(), user, key, user.UserID)
	require.NoError(t, err)

	// delete the idempotency key we used to create the account
	err = service.DeleteIdempotencyKeyByID(context.Background(), key)
	require.NoError(t, err)

	transaction := utils.RandomTransaction()
	transaction.AccountID = createdAccount.AccountID
	key = utils.RandomIdempotencyKey()
	createdTransaction, err := service.CreateTransaction(context.Background(), transaction, key, user.UserID)
	require.NoError(t, err)
	require.NotEmpty(t, createdTransaction)
	require.Equal(t, transaction.AccountID, createdTransaction.AccountID)

	err = service.DeleteIdempotencyKeyByID(context.Background(), key)
	require.NoError(t, err)

	// Cleanup the account and transaction we created to test
	// the transaction table uses ON DELETE CASCADE so deleting an account automatically deletes all transactions associated with it
	key = utils.RandomIdempotencyKey()
	err = service.DeleteAccountByAccountNumber(context.Background(), createdAccount.AccountNumber, key, user.UserID)
	require.NoError(t, err)

	err = service.DeleteIdempotencyKeyByID(context.Background(), key)
	require.NoError(t, err)
}

// Multiple concurrent goroutines that create different transactions (different idempotency keys) to the same account should update the account balance correctly
func TestCreateMultipleTransactions_Success(t *testing.T) {

	user := utils.RandomUser()
	key := utils.RandomIdempotencyKey()
	createdAccount, err := service.CreateAccount(context.Background(), user, key, user.UserID)
	require.NoError(t, err)

	// delete the idempotency key we used to create the account
	err = service.DeleteIdempotencyKeyByID(context.Background(), key)
	require.NoError(t, err)

	balance := createdAccount.Balance
	numTransactions := 30
	errChan := make(chan error, numTransactions)
	results := make(chan *model.Transaction, numTransactions)
	keyChan := make(chan uuid.UUID, numTransactions)

	for i := 0; i < numTransactions; i++ {
		go func() {
			key := utils.RandomIdempotencyKey()
			transaction := utils.RandomTransaction()
			transaction.AccountID = createdAccount.AccountID
			result, err := service.CreateTransaction(context.Background(), transaction, key, user.UserID)
			if err != nil {
				errChan <- err
				return
			}
			results <- result
			errChan <- nil
			keyChan <- key
		}()
	}

	for i := 0; i < numTransactions; i++ {
		err := <-errChan
		require.NoError(t, err)
		result := <-results
		require.NotEmpty(t, result)
		balance += result.Amount

		key := <-keyChan
		err = service.DeleteIdempotencyKeyByID(context.Background(), key)
		require.NoError(t, err)
	}

	// check if the balance is updated correctly
	finalAccount, err := service.GetAccountByAccountNumber(context.Background(), createdAccount.AccountNumber, user.UserID)
	require.NoError(t, err)
	require.Equal(t, balance, finalAccount.Balance)

	// Cleanup the account and transactions we created to test
	// the transaction table uses ON DELETE CASCADE so deleting an account automatically deletes all transactions associated with it
	key = utils.RandomIdempotencyKey()
	err = service.DeleteAccountByAccountNumber(context.Background(), createdAccount.AccountNumber, key, user.UserID)
	require.NoError(t, err)
	err = service.DeleteIdempotencyKeyByID(context.Background(), key)
	require.NoError(t, err)
}

// Multiple concurrent goroutines that create the same transactions (same idempotency keys) should update the account balance only once
func TestCreateMultipleTransactionsIdempotencyKey_Success(t *testing.T) {

	user := utils.RandomUser()
	key := utils.RandomIdempotencyKey()
	createdAccount, err := service.CreateAccount(context.Background(), user, key, user.UserID)
	require.NoError(t, err)

	// delete the idempotency key we used to create the account
	err = service.DeleteIdempotencyKeyByID(context.Background(), key)
	require.NoError(t, err)

	balance := createdAccount.Balance
	numTransactions := 30

	key = utils.RandomIdempotencyKey()
	transaction := utils.RandomTransaction()
	transaction.AccountID = createdAccount.AccountID

	expectedBalance := balance + transaction.Amount

	var wg sync.WaitGroup

	for i := 0; i < numTransactions; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			service.CreateTransaction(context.Background(), transaction, key, user.UserID)
		}()
	}

	wg.Wait()

	// delete the idempotency key we used to create the transactions
	err = service.DeleteIdempotencyKeyByID(context.Background(), key)
	require.NoError(t, err)

	// check if the balance is updated correctly
	finalAccount, err := service.GetAccountByAccountNumber(context.Background(), createdAccount.AccountNumber, user.UserID)
	require.NoError(t, err)
	require.Equal(t, expectedBalance, finalAccount.Balance)

	// Cleanup the account and transaction we created to test
	// the transaction table uses ON DELETE CASCADE so deleting an account automatically deletes all transactions associated with it
	key = utils.RandomIdempotencyKey()
	err = service.DeleteAccountByAccountNumber(context.Background(), createdAccount.AccountNumber, key, user.UserID)
	require.NoError(t, err)

	err = service.DeleteIdempotencyKeyByID(context.Background(), key)
	require.NoError(t, err)
}
