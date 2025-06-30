package repository

import (
	"auth/db/initialize"
	"auth/db/sqlc"
	"auth/utils"
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

var (
	testRepo *AuthRepository
	testDB   *sqlx.DB
)

func setupTestDB() func(t *testing.T) {
	testDB = initialize.ConnectDB()
	testRepo = NewAuthRepository(testDB)

	// Return a teardown function that will be run after each test
	return func(t *testing.T) {
		// Close the database connection
		err := testDB.Close()
		require.NoError(t, err)
	}
}

func randomCreateUserParams() (sqlc.CreateUserParams, error) {
	var randomEmail string

	passwordHash, err := utils.HashPassword(utils.RandomString(10))
	if err != nil {
		return sqlc.CreateUserParams{}, err
	}
	for { // make sure we don't create an alreay existing account
		randomEmail = utils.RandomEmail()
		_, err := testRepo.queries.GetUserByEmail(context.Background(), randomEmail)
		if err != nil {
			break
		}
	}
	return sqlc.CreateUserParams{
		ID:           uuid.New(),
		Email:        randomEmail,
		PasswordHash: passwordHash,
	}, nil
}

func TestCreateUser_Success(t *testing.T) {
	teardown := setupTestDB()
	defer teardown(t)

	// create a transaction so we can rollback after testing.
	tx, err := testDB.BeginTx(context.Background(), nil)
	q := testRepo.queries.WithTx(tx)
	require.NoError(t, err)
	defer tx.Rollback()

	// Create a new user
	createUserArg, err := randomCreateUserParams()
	require.NoError(t, err)
	res, err := q.CreateUser(context.Background(), createUserArg)
	require.NoError(t, err)
	require.NotEmpty(t, res)
	require.NotEmpty(t, res.ID)
	require.NotEmpty(t, res.CreatedAt)
	require.NotEmpty(t, res.UpdatedAt)
	require.Equal(t, res.Email, createUserArg.Email)
	require.Equal(t, res.PasswordHash, createUserArg.PasswordHash)
	fmt.Println("Passed TestCreateUser_Success")
	// tx.Commit() // for testing that it does commit if this line runs.
}
