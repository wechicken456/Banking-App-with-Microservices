-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (id, user_id, token ,expires_at)
VALUES ($1,$2,$3,$4)
RETURNING *;

-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens WHERE token = $1;

-- name: DeleteRefreshToken :exec
DELETE FROM refresh_tokens WHERE token = $1;
