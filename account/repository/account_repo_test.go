package repository

import (
	"account/db/initialize"
	"account/db/sqlc"
	"account/model"
	"account/utils"
	"context"
	"database/sql"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

var testRepo *AccountRepository
var testDB *sqlx.DB

func setupTestDB() func(t *testing.T) {
	godotenv.Load("../.env")
	testDB = initialize.ConnectDB()
	testRepo = NewAccountRepository(testDB)

	// Return a teardown function that will be run after each test
	return func(t *testing.T) {
		// Close the database connection
		err := testDB.Close()
		require.NoError(t, err)
	}
}

func TestCreateAccount_Success(t *testing.T) {
	var account *model.Account

	teardown := setupTestDB()
	defer teardown(t)

	numTests := 10
	errChan := make(chan error)
	results := make(chan sqlc.Account)
	txChan := make(chan *sql.Tx)
	for i := 0; i < numTests; i++ {
		go func() {
			tx, err := testDB.BeginTx(context.Background(), nil)
			q := testRepo.queries.WithTx(tx)
			require.NoError(t, err)

			account = utils.RandomAccount()
			res, err := q.CreateAccount(context.Background(), sqlc.CreateAccountParams{
				ID:            account.AccountID,
				AccountNumber: account.AccountNumber,
				UserID:        account.UserID,
				Balance:       account.Balance,
			})

			errChan <- err
			results <- res
			txChan <- tx
		}()
	}

	for i := 0; i < numTests; i++ {
		err := <-errChan
		require.NoError(t, err)
		res := <-results
		require.NotEmpty(t, res)
		tx := <-txChan
		tx.Rollback()
	}
}

// Same account number should fail
func TestCreateAccount_Fail(t *testing.T) {
	var account *model.Account

	teardown := setupTestDB()
	defer teardown(t)

	tx, err := testDB.BeginTx(context.Background(), nil)
	require.NoError(t, err)
	defer tx.Rollback()
	q := testRepo.queries.WithTx(tx)
	require.NoError(t, err)

	account = utils.RandomAccount()
	_, err = q.CreateAccount(context.Background(), sqlc.CreateAccountParams{
		ID:            account.AccountID,
		AccountNumber: account.AccountNumber,
		UserID:        account.UserID,
		Balance:       account.Balance,
	})
	require.NoError(t, err)
	_, err = q.CreateAccount(context.Background(), sqlc.CreateAccountParams{
		ID:            account.AccountID,
		AccountNumber: account.AccountNumber,
		UserID:        account.UserID,
		Balance:       account.Balance,
	})
	require.Error(t, err)

}

func TestGetAccountByAccountNumber_Success(t *testing.T) {
	teardown := setupTestDB()
	defer teardown(t)

	tx, err := testDB.BeginTx(context.Background(), nil)
	require.NoError(t, err)
	defer tx.Rollback()
	q := testRepo.queries.WithTx(tx)

	account := utils.RandomAccount()
	res1, err := q.CreateAccount(context.Background(), sqlc.CreateAccountParams{
		ID:            account.AccountID,
		AccountNumber: account.AccountNumber,
		UserID:        account.UserID,
		Balance:       account.Balance,
	})
	require.NoError(t, err)

	res2, err := q.GetAccountByAccountNumber(context.Background(), account.AccountNumber)
	require.NoError(t, err)
	require.NotEmpty(t, res2)
	require.Equal(t, res1, res2)
}

// verify that concurrent transactions that adds the same amount to an account updates the balance correctly
func TestAddToAccountBalance_Success(t *testing.T) {
	teardown := setupTestDB()
	defer teardown(t)

	createdAccountTx, err := testDB.BeginTx(context.Background(), nil)
	require.NoError(t, err)
	defer createdAccountTx.Rollback()
	q := testRepo.queries.WithTx(createdAccountTx)

	account := utils.RandomAccount()
	_, err = q.CreateAccount(context.Background(), sqlc.CreateAccountParams{
		ID:            account.AccountID,
		AccountNumber: account.AccountNumber,
		UserID:        account.UserID,
		Balance:       account.Balance,
	})
	createdAccountTx.Commit()
	require.NoError(t, err)

	addAmount := int64(1000)
	numTests := 10
	errChan := make(chan error)
	results := make(chan sqlc.Account)
	for i := 0; i < numTests; i++ {
		go func() {
			updateTx, err := testDB.BeginTx(context.Background(), nil)
			require.NoError(t, err)
			q := testRepo.queries.WithTx(updateTx)

			res2, err := q.AddToAccountBalance(context.Background(), sqlc.AddToAccountBalanceParams{
				AccountNumber: account.AccountNumber,
				Amount:        addAmount,
			})
			if err != nil {
				updateTx.Rollback()
			} else {
				updateTx.Commit()
			}
			errChan <- err
			results <- res2
		}()
	}

	// check for errors
	for i := 0; i < numTests; i++ {
		err := <-errChan
		require.NoError(t, err)
		res := <-results
		require.NotEmpty(t, res)
	}

	// check if the balance is updated correctly
	res, err := testRepo.GetAccountByAccountNumber(context.Background(), account.AccountNumber)
	require.NoError(t, err)
	require.NotEmpty(t, res)
	require.Equal(t, res.Balance, account.Balance+addAmount*int64(numTests))

	// delete the account that we created to test
	err = testRepo.queries.DeleteAccountByAccountNumber(context.Background(), account.AccountNumber)
	require.NoError(t, err)
}
