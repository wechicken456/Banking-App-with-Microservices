-- +goose Up
-- +goose StatementBegin
CREATE TABLE idempotency_keys (
    key_id UUID NOT NULL DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,          -- user ID of the user who initiated the request 

    status VARCHAR(30) NOT NULL CHECK (status IN ('PENDING', 'COMPLETED', 'FAILED')),
    response_message TEXT NOT NULL,          -- gRPC response message

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    expired_at TIMESTAMPTZ DEFAULT NOW() + INTERVAL '24 hour',

    PRIMARY KEY (key_id, user_id) -- composite primary key to enable authorization as well
);

CREATE INDEX idx_idempotency_key_id ON idempotency_keys (key_id);
CREATE INDEX idx_idempotency_key_user_id ON idempotency_keys (user_id);
CREATE INDEX idx_idempotency_key_expired_at ON idempotency_keys (expired_at);

CREATE TRIGGER trigger_update_timestamp_idempotency_keys
BEFORE UPDATE ON idempotency_keys
FOR EACH ROW EXECUTE FUNCTION update_timestamp();


CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID REFERENCES accounts(id) ON DELETE CASCADE NOT NULL,
    transaction_type VARCHAR(30) NOT NULL CHECK (transaction_type IN ('CREDIT', 'DEBIT', 'TRANSFER_DEBIT', 'TRANSFER_CREDIT')),
    amount BIGINT NOT NULL CHECK (amount <> 0),
    status VARCHAR(30) NOT NULL CHECK (status IN ('PENDING', 'COMPLETED', 'FAILED', 'REVERSED')),
    transfer_id UUID,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_transaction_account_id ON transactions (account_id);
CREATE INDEX idx_transaction_transfer_id ON transactions (transfer_id);

-- Automatically update the updated_at column for a row whenever that row is updated.
CREATE TRIGGER trigger_update_timestamp_transactions
BEFORE UPDATE ON transactions
FOR EACH ROW EXECUTE FUNCTION update_timestamp();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER trigger_update_timestamp_idempotency_keys ON idempotency_keys;
DROP INDEX idx_idempotency_key_id;
DROP INDEX idx_idempotency_key_user_id;
DROP INDEX idx_idempotency_key_expired_at;

DROP TABLE idempotency_keys;
DROP TRIGGER trigger_update_timestamp_transactions ON transactions;
DROP INDEX idx_transaction_account_id;
DROP INDEX idx_transaction_transfer_id;
DROP TABLE transactions;
-- +goose StatementEnd
