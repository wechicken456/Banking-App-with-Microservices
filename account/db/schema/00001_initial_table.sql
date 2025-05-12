-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id varchar NOT NULL,
    account_number bigint UNIQUE NOT NULL,
    balance bigint NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- CREATE INDEX to optimize queries on the account_number column since we'll probably be searching by account number frequently.
CREATE INDEX idx_account_number ON accounts (account_number);
CREATE INDEX idx_account_user_id ON accounts (user_id);

-- Automatically update the updated_at column for a row whenever that row is updated. 
CREATE OR REPLACE FUNCTION update_timestamp() 
RETURNS TRIGGER AS $$
BEGIN 
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_timestamp_accounts
BEFORE UPDATE ON accounts
FOR EACH ROW EXECUTE FUNCTION update_timestamp();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER trigger_update_timestamp ON accounts;
DROP FUNCTION update_timestamp();
DROP TABLE accounts;
-- +goose StatementEnd
