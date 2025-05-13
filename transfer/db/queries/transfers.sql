-- name: CreateTransfer :one
INSERT INTO transfers (id, idempotency_key, from_account_id, to_account_id, amount, status)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetTransfersByFromID :many
SELECT * FROM transfers WHERE from_account_id = $1;

-- name: GetTransfersByToID :many
SELECT * FROM transfers WHERE to_account_id = $1;

-- name: GetTransferByID :one
SELECT * FROM transfers WHERE id = $1;

-- name: GetTransferByIdempotencyKey :one
SELECT * FROM transfers WHERE idempotency_key = $1;

-- name: ListTransfers :many
SELECT * FROM transfers ORDER BY updated_at LIMIT $1;

-- name: UpdateTransferStatus :exec
UPDATE transfers
SET status = sqlc.arg(status)
WHERE id = sqlc.arg(id)
RETURNING *;

