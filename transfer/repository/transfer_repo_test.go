package repository

import (
	"context"
	"testing"
	"transfer/db/initialize"
	"transfer/model"
	"transfer/utils"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

// setupTestDB initializes a test database connection and returns a teardown function.
func setupTestDB(t *testing.T) (*sqlx.DB, func()) {
	db := initialize.ConnectDB()
	return db, func() {
		err := db.Close()
		require.NoError(t, err)
	}
}

// TestCreateTransfer_Success tests the successful creation of a single transfer.
func TestCreateTransfer_Success(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()
	repo := NewTransferRepository(db)

	// Create two accounts for the transfer
	account1 := utils.RandomAccount()
	account2 := utils.RandomAccount()

	// Create a transfer between the accounts
	transfer := &model.Transfer{
		TransferID:     uuid.New(),
		FromAccountID:  account1.AccountID,
		ToAccountID:    account2.AccountID,
		IdempotencyKey: uuid.NewString(),
		Amount:         100,
		Status:         "PENDING",
	}
	createdTransfer, err := repo.CreateTransfer(context.Background(), transfer)
	require.NoError(t, err)
	require.NotEmpty(t, createdTransfer)
	require.Equal(t, *transfer, *createdTransfer)
}

// Tests that creating a transfer with a duplicate idempotency key fails.
func TestCreateTransfer_Idempotency(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()
	repo := NewTransferRepository(db)

	// Create two accounts
	account1 := utils.RandomAccount()
	account2 := utils.RandomAccount()

	// Create a transfer
	idempotencyKey := uuid.NewString()
	transfer1 := &model.Transfer{
		TransferID:     uuid.New(),
		FromAccountID:  account1.AccountID,
		ToAccountID:    account2.AccountID,
		IdempotencyKey: idempotencyKey,
		Amount:         100,
		Status:         "PENDING",
	}
	_, err := repo.CreateTransfer(context.Background(), transfer1)
	require.NoError(t, err)

	// Attempt to create another transfer with the same idempotency key
	transfer2 := &model.Transfer{
		TransferID:     uuid.New(),
		FromAccountID:  account1.AccountID,
		ToAccountID:    account2.AccountID,
		IdempotencyKey: idempotencyKey,
		Amount:         200,
		Status:         "PENDING",
	}
	_, err = repo.CreateTransfer(context.Background(), transfer2)
	require.Error(t, err)
}

// Tests that creating multiple transfers concurrently doesn't give an error
func TestCreateMultipleTransfers_Success(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()
	repo := NewTransferRepository(db)

	// Create two accounts
	account1 := utils.RandomAccount()
	account2 := utils.RandomAccount()

	numTransfers := 10
	errChan := make(chan error, numTransfers)
	results := make(chan *model.Transfer, numTransfers)

	// Create transfers concurrently
	for i := 0; i < numTransfers; i++ {
		go func() {
			transfer := &model.Transfer{
				TransferID:     uuid.New(),
				FromAccountID:  account1.AccountID,
				ToAccountID:    account2.AccountID,
				IdempotencyKey: uuid.NewString(),
				Amount:         100,
				Status:         "PENDING",
			}
			createdTransfer, err := repo.CreateTransfer(context.Background(), transfer)
			if err != nil {
				errChan <- err
				return
			}
			results <- createdTransfer
			errChan <- nil
		}()
	}

	// Verify all transfers were created successfully
	for i := 0; i < numTransfers; i++ {
		err := <-errChan
		require.NoError(t, err)
		transfer := <-results
		require.NotEmpty(t, transfer)
	}
}

func TestGetTransferByID_Success(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()
	repo := NewTransferRepository(db)

	// Create accounts and a transfer
	account1 := utils.RandomAccount()
	account2 := utils.RandomAccount()

	transfer := &model.Transfer{
		TransferID:     uuid.New(),
		FromAccountID:  account1.AccountID,
		ToAccountID:    account2.AccountID,
		IdempotencyKey: uuid.NewString(),
		Amount:         100,
		Status:         "PENDING",
	}
	createdTransfer, err := repo.CreateTransfer(context.Background(), transfer)
	require.NoError(t, err)

	// Retrieve and verify the transfer
	retrievedTransfer, err := repo.GetTransferByID(context.Background(), createdTransfer.TransferID)
	require.NoError(t, err)
	require.Equal(t, createdTransfer, retrievedTransfer)
}
