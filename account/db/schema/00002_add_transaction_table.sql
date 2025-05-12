-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID REFERENCES accounts(id) ON DELETE CASCADE,
    transaction_type VARCHAR(30) NOT NULL CHECK (transaction_type IN ('DEPOSIT', 'WITHDRAWAL', 'TRANSFER_DEBIT', 'TRANSFER_CREDIT')),
    amount BIGINT NOT NULL CHECK (amount > 0),
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
DROP TRIGGER trigger_update_timestamp_transactions ON transactions;
DROP TABLE transactions;
DROP INDEX idx_transaction_account_id;
DROP INDEX idx_transaction_transfer_id;
-- +goose StatementEnd
