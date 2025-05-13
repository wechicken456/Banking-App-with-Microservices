-- name: CreateTransaction :one
INSERT INTO transactions (id, idempotency_key, account_id, amount, transaction_type, status, transfer_id)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetTransactionByID :one
SELECT * FROM transactions WHERE id = $1;

-- name: GetTransactionByIdempotencyKey :one
SELECT * FROM transactions WHERE idempotency_key = $1;

-- name: GetTransactionsByAccountID :many
SELECT * FROM transactions WHERE account_id = $1;

-- name: GetTransactionByTransferID :many
SELECT * FROM transactions WHERE transfer_id = $1;

-- name: ListTransactions :many
SELECT * FROM transactions ORDER BY updated_at LIMIT $1;

-- name: UpdateTransactionStatus :exec
UPDATE transactions
SET status = sqlc.arg(status)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteTransactionByID :exec
DELETE FROM transactions
WHERE id = $1
RETURNING *;




