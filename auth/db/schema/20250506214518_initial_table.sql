-- +goose Up
-- +goose StatementBegin
CREATE TABLE "users" (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email varchar(255) NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);


CREATE INDEX ON "users" ("email");

CREATE INDEX ON "refresh_tokens" ("user_id");

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE "refresh_tokens";
DROP TABLE "accounts";
-- +goose StatementEnd
