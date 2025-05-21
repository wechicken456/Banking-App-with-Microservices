-- name: CreateIdempotencyKey :one 
INSERT INTO idempotency_keys (key_id, user_id, status, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW())
RETURNING *;

-- name: GetIdempotencyKeyByID :one
SELECT * FROM idempotency_keys WHERE key_id = $1;

-- name: UpdateIdempotencyKeyByID :exec
UPDATE idempotency_keys
SET status = $1, response_code = $2, response_message = $3, updated_at = NOW()
WHERE key_id = $4;

-- name: DeleteIdempotencyKeyByID :exec
DELETE FROM idempotency_keys
WHERE key_id = $1;

-- name: DeleteIdempotencyKeysByUserID :exec
DELETE FROM idempotency_keys
WHERE user_id = $1;

-- name: DeleteIdempotencyKeysByExpiredAt :exec
DELETE FROM idempotency_keys
WHERE expired_at < NOW();


-- name: GetOrClaimIdempotencyKey :one
INSERT INTO idempotency_keys (
     -- Insert a new idempotency key. If concurrenct transactions already created the key, its status should be "COMPLETED" or "FAILED". 
     -- Else, we're the first to create it, and we set it to "PENDING".
     -- This statement will blockResponseMessage if there are concurrent transactions inserting the same row, even if they haven't been committed/rollbacked.
    key_id,
    user_id,
    status,            
    response_code,      
    response_message,     
    created_at,
    updated_at,
    expired_at
)
VALUES (
    $1,  
    $2,  -
    'PENDING',
    0, 
    "placeholder", 
    NOW(),
    NOW(),
    $3   -- expired_at (e.g., NOW() + interval '24 hours')
)
ON CONFLICT (idempotency_key_id, user_id)
DO UPDATE SET
    -- Only touch 'updated_at' if the existing status is 'PENDING', to signify this transaction is actively looking at it.
    -- If it's already 'COMPLETED' or 'FAILED', we don't want to modify it here; RETURNING * will give us its state.
    -- The 'status' is NOT changed here by the DO UPDATE clause if it was already terminal.
    -- If it was PENDING, it remains PENDING. If it was a new insert, it gets the $3 status.
    updated_at = CASE
                    WHEN idempotency_keys.status = 'PENDING' -- $3 is 'PENDING'
                        THEN NOW()
                    ELSE idempotency_keys.updated_at -- Keep existing updated_at if it was COMPLETED/FAILED or different pending
                 END
RETURNING *;

-- name: UpdateIdempotencyKey :one
UPDATE idempotency_keys
SET status = $1, response_code = $2, response_message = $3, updated_at = NOW(), expired_at = NOW() + interval '24 hours'
WHERE key_id = $4
RETURNING *;