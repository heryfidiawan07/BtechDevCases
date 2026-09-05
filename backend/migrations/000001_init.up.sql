CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    balance NUMERIC(18, 2) NOT NULL DEFAULT 0 CHECK (balance >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE transfers (
    id BIGSERIAL PRIMARY KEY,
    from_user_id BIGINT NOT NULL REFERENCES users (id),
    to_user_id BIGINT NOT NULL REFERENCES users (id),
    amount NUMERIC(18, 2) NOT NULL CHECK (amount > 0),
    notes TEXT NOT NULL DEFAULT '',
    idempotency_key TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT transfers_from_idempotency_key UNIQUE (from_user_id, idempotency_key)
);

CREATE INDEX idx_transfers_from_created_at ON transfers (from_user_id, created_at DESC);
CREATE INDEX idx_transfers_to_created_at ON transfers (to_user_id, created_at DESC);
