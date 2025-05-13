-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS transfers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    idempotency_key VARCHAR(100) UNIQUE NOT NULL,
    from_account_id UUID NOT NULL,
    to_account_id UUID NOT NULL,
    amount BIGINT NOT NULL CHECK (amount > 0),
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'COMPLETED', 'FAILED', 'REVERSED')),
    created_at TIMESTAMPTZ DEFAULT NOw(),
    updated_at TIMESTAMPTZ DEFAULT NOw()
);

CREATE INDEX idx_transfer_from_account_id ON transfers (from_account_id);
CREATE INDEX idx_transfer_to_account_id ON transfers (to_account_id);
CREATE INDEX idx_transfer_idempotency_key ON transfers (idempotency_key);

CREATE TRIGGER trigger_update_timestamp_transfers
BEFORE UPDATE ON transfers
FOR EACH ROW EXECUTE FUNCTION update_timestamp();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER trigger_update_timestamp_transfers ON transfers;
DROP INDEX idx_transfer_from_account_id;
DROP INDEX idx_transfer_to_account_id;
DROP INDEX idx_transfer_idempotency_key;
DROP TABLE transfers;
-- +goose StatementEnd
