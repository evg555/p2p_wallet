-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'ledger_transaction_type') THEN
        CREATE TYPE ledger_transaction_type AS ENUM (
            'topup',
            'transfer',
            'hold',
            'release',
            'capture',
            'withdrawal'
        );
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS ledger_transactions (
    id BIGSERIAL PRIMARY KEY,
    type ledger_transaction_type NOT NULL,
    reference_id VARCHAR(50) NULL,
    idemp_key VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    CONSTRAINT uq_ledger_transactions_idemp_key UNIQUE (idemp_key)
);

CREATE TABLE IF NOT EXISTS ledger_entries (
    id BIGSERIAL PRIMARY KEY,
    transaction_id BIGINT NOT NULL,
    wallet_id BIGINT NOT NULL,
    amount BIGINT NOT NULL DEFAULT 0,
    currency wallet_currency NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    CONSTRAINT fk_ledger_entries_transaction_id
        FOREIGN KEY (transaction_id) REFERENCES ledger_transactions(id) ON DELETE CASCADE,
    CONSTRAINT fk_ledger_entries_wallet_id
        FOREIGN KEY (wallet_id) REFERENCES wallet_balance_snapshots(wallet_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS ledger_entries;
DROP TABLE IF EXISTS ledger_transactions;

DROP TYPE IF EXISTS ledger_transaction_type;
-- +goose StatementEnd
