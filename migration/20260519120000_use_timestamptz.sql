-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
    ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC',
    ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE 'UTC',
    ALTER COLUMN created_at SET DEFAULT now();

ALTER TABLE wallets
    ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC',
    ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE 'UTC',
    ALTER COLUMN created_at SET DEFAULT now();

ALTER TABLE wallet_balance_snapshots
    ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE 'UTC',
    ALTER COLUMN updated_at SET DEFAULT now();

ALTER TABLE ledger_transactions
    ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC',
    ALTER COLUMN created_at SET DEFAULT now();

ALTER TABLE ledger_entries
    ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC',
    ALTER COLUMN created_at SET DEFAULT now();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE ledger_entries
    ALTER COLUMN created_at TYPE TIMESTAMP USING created_at AT TIME ZONE 'UTC',
    ALTER COLUMN created_at SET DEFAULT now();

ALTER TABLE ledger_transactions
    ALTER COLUMN created_at TYPE TIMESTAMP USING created_at AT TIME ZONE 'UTC',
    ALTER COLUMN created_at SET DEFAULT now();

ALTER TABLE wallet_balance_snapshots
    ALTER COLUMN updated_at TYPE TIMESTAMP USING updated_at AT TIME ZONE 'UTC',
    ALTER COLUMN updated_at SET DEFAULT now();

ALTER TABLE wallets
    ALTER COLUMN created_at TYPE TIMESTAMP USING created_at AT TIME ZONE 'UTC',
    ALTER COLUMN updated_at TYPE TIMESTAMP USING updated_at AT TIME ZONE 'UTC',
    ALTER COLUMN created_at SET DEFAULT now();

ALTER TABLE users
    ALTER COLUMN created_at TYPE TIMESTAMP USING created_at AT TIME ZONE 'UTC',
    ALTER COLUMN updated_at TYPE TIMESTAMP USING updated_at AT TIME ZONE 'UTC',
    ALTER COLUMN created_at SET DEFAULT now();
-- +goose StatementEnd
