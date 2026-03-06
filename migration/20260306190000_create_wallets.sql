-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'wallet_currency') THEN
        CREATE TYPE wallet_currency AS ENUM ('USD', 'EUR');
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'wallet_status') THEN
        CREATE TYPE wallet_status AS ENUM ('active', 'blocked');
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS wallets (
    id BIGSERIAL PRIMARY KEY,
    currency wallet_currency NOT NULL,
    status wallet_status NOT NULL DEFAULT 'active',
    user_id BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NULL,
    CONSTRAINT fk_wallets_user_id FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT uq_wallet_user_currency UNIQUE (user_id, currency)
);

CREATE TABLE IF NOT EXISTS wallet_balance_snapshots (
    wallet_id BIGINT PRIMARY KEY,
    currency wallet_currency NOT NULL,
    held_amount BIGINT NOT NULL DEFAULT 0,
    total_amount BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    CONSTRAINT fk_wallet_balance_snapshots_wallet_id
        FOREIGN KEY (wallet_id) REFERENCES wallets(id) ON DELETE CASCADE,
    CONSTRAINT chk_balance_nonnegative
        CHECK (held_amount >= 0 AND total_amount >= 0 AND held_amount <= total_amount)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS wallet_balance_snapshots;
DROP TABLE IF EXISTS wallets;

DROP TYPE IF EXISTS wallet_status;
DROP TYPE IF EXISTS wallet_currency;
-- +goose StatementEnd
