-- name: CreateAccount :one
INSERT INTO accounts (id, account_number, user_id, balance)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetAccountByID :one
SELECT * FROM accounts WHERE id = $1;

-- name: GetAccountByUserID :many
SELECT * FROM accounts WHERE user_id = $1;

-- name: GetAccountByAccountNumber :one
SELECT * FROM accounts WHERE account_number = $1;

-- name: ListAccounts :many
SELECT * FROM accounts ORDER BY id LIMIT $1;


-- name: AddToAccountBalance :one
UPDATE accounts
SET balance = balance + sqlc.arg(amount)
WHERE account_number = sqlc.arg(account_number)
RETURNING *;
