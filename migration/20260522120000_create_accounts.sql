-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'account_type') THEN
        CREATE TYPE account_type AS ENUM (
            'available',
            'held',
            'external_in',
            'external_out',
            'fees'
        );
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS accounts (
    id BIGSERIAL PRIMARY KEY,
    wallet_id BIGINT NULL,
    account_type account_type NOT NULL DEFAULT 'available',
    currency wallet_currency NOT NULL,
    status wallet_status NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NULL,
    CONSTRAINT fk_accounts_wallet_id
        FOREIGN KEY (wallet_id) REFERENCES wallets(id) ON DELETE CASCADE,
    CONSTRAINT chk_accounts_wallet_scope
        CHECK (
            (wallet_id IS NOT NULL AND account_type IN ('available', 'held'))
            OR
            (wallet_id IS NULL AND account_type IN ('external_in', 'external_out', 'fees'))
        )
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_accounts_wallet_type_currency
    ON accounts (wallet_id, account_type, currency)
    WHERE wallet_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_accounts_system_type_currency
    ON accounts (account_type, currency)
    WHERE wallet_id IS NULL;

INSERT INTO accounts (wallet_id, account_type, currency, status, created_at, updated_at)
SELECT
    w.id,
    'available',
    w.currency,
    w.status,
    w.created_at,
    w.updated_at
FROM wallets w;

INSERT INTO accounts (wallet_id, account_type, currency, status, created_at, updated_at)
SELECT
    w.id,
    'held',
    w.currency,
    w.status,
    w.created_at,
    w.updated_at
FROM wallets w;

INSERT INTO accounts (wallet_id, account_type, currency, status)
SELECT NULL, system_accounts.account_type, c.currency, 'active'
FROM (SELECT DISTINCT currency FROM wallets) c
CROSS JOIN (
    VALUES
        ('external_in'::account_type),
        ('external_out'::account_type),
        ('fees'::account_type)
) AS system_accounts(account_type);

ALTER TABLE ledger_entries
    ADD COLUMN account_id BIGINT NULL;

UPDATE ledger_entries le
SET account_id = a.id
FROM accounts a
WHERE a.wallet_id = le.wallet_id
  AND a.account_type = 'available'
  AND a.currency = le.currency;

ALTER TABLE ledger_entries
    ALTER COLUMN account_id SET NOT NULL,
    ADD CONSTRAINT fk_ledger_entries_account_id
        FOREIGN KEY (account_id) REFERENCES accounts(id),
    DROP CONSTRAINT fk_ledger_entries_wallet_id,
    DROP COLUMN wallet_id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE ledger_entries
    ADD COLUMN wallet_id BIGINT NULL;

UPDATE ledger_entries le
SET wallet_id = a.wallet_id
FROM accounts a
WHERE a.id = le.account_id;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM ledger_entries le
        JOIN accounts a ON a.id = le.account_id
        WHERE a.wallet_id IS NULL
    ) THEN
        RAISE EXCEPTION 'cannot rollback accounts migration: ledger_entries contains system account references';
    END IF;
END $$;

ALTER TABLE ledger_entries
    ALTER COLUMN wallet_id SET NOT NULL,
    ADD CONSTRAINT fk_ledger_entries_wallet_id
        FOREIGN KEY (wallet_id) REFERENCES wallet_balance_snapshots(wallet_id),
    DROP CONSTRAINT fk_ledger_entries_account_id,
    DROP COLUMN account_id;

DROP INDEX IF EXISTS uq_accounts_system_type_currency;
DROP INDEX IF EXISTS uq_accounts_wallet_type_currency;
DROP TABLE IF EXISTS accounts;

DROP TYPE IF EXISTS account_type;
-- +goose StatementEnd
