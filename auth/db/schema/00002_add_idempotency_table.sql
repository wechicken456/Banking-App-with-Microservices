-- +goose Up
-- +goose StatementBegin
CREATE TABLE idempotency_keys (
    key_id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),

    status VARCHAR(30) NOT NULL CHECK (status IN ('PENDING', 'COMPLETED', 'FAILED')),
    response_message TEXT NOT NULL,          -- gRPC response message

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    expired_at TIMESTAMPTZ DEFAULT NOW() + INTERVAL '24 hour'

);

CREATE INDEX idx_idempotency_key_id ON idempotency_keys (key_id);
CREATE INDEX idx_idempotency_key_expired_at ON idempotency_keys (expired_at);

-- Automatically update the updated_at column for a row whenever that row is updated. 
CREATE OR REPLACE FUNCTION update_timestamp() 
RETURNS TRIGGER AS $$
BEGIN 
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


CREATE TRIGGER trigger_update_timestamp_idempotency_keys
BEFORE UPDATE ON idempotency_keys
FOR EACH ROW EXECUTE FUNCTION update_timestamp();


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER trigger_update_timestamp_idempotency_keys ON idempotency_keys;
DROP INDEX idx_idempotency_key_id;
DROP INDEX idx_idempotency_key_expired_at;
DROP TABLE idempotency_keys;


-- +goose StatementEnd
