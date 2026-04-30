package repository

import (
	"context"
	"errors"
	"fmt"

	"p2p_wallet/internal/domain"
	"p2p_wallet/internal/errs"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type balancePostgresRepo struct {
	client *pgxpool.Pool
}

func NewBalancePostgresRepo(client *pgxpool.Pool) *balancePostgresRepo {
	return &balancePostgresRepo{
		client: client,
	}
}

func (b *balancePostgresRepo) CreateTransaction(ctx context.Context, transaction *domain.Transaction) (*domain.Transaction, error) {
	tx, err := b.client.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin balance repo transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	query, args, err := psql.Insert("ledger_transactions").
		Columns("type", "status", "idemp_key").
		Values(domain.TypeTransfer, transaction.Status, transaction.IdempotencyKey, transaction.CreatedAt).
		Suffix("RETURNING id, created_at").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build insert transaction query: %w", err)
	}

	savedTransaction := &domain.Transaction{
		IdempotencyKey: transaction.IdempotencyKey,
		Status:         transaction.Status,
		Entries:        transaction.Entries,
	}

	err = tx.QueryRow(ctx, query, args...).Scan(&savedTransaction.ID, transaction.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, errs.ErrTransactionAlreadyCreated
		}

		return nil, fmt.Errorf("insert transaction: %w", err)
	}

	entryFrom := savedTransaction.Entries[0]
	query, args, err = psql.Insert("ledger_entries").
		Columns("transaction_id", "wallet_id", "amount", "currency").
		Values(savedTransaction.ID, entryFrom.WalletID, entryFrom.Money.Amount(), entryFrom.Money.Currency()).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build insert transaction query: %w", err)
	}

	savedEntryFrom := &domain.Entry{
		WalletID: entryFrom.WalletID,
		Money:    entryFrom.Money,
	}

	err = tx.QueryRow(ctx, query, args...).Scan(&savedEntryFrom.ID)
	if err != nil {
		return nil, fmt.Errorf("insert entry: %w", err)
	}

	entryTo := savedTransaction.Entries[1]
	query, args, err = psql.Insert("ledger_entries").
		Columns("transaction_id", "wallet_id", "amount", "currency").
		Values(savedTransaction.ID, entryTo.WalletID, entryTo.Money.Amount(), entryTo.Money.Currency()).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build insert transaction query: %w", err)
	}

	savedEntryTo := &domain.Entry{
		WalletID: entryTo.WalletID,
		Money:    entryTo.Money,
	}

	err = tx.QueryRow(ctx, query, args...).Scan(&savedEntryTo.ID)
	if err != nil {
		return nil, fmt.Errorf("insert entry: %w", err)
	}

	savedTransaction.Entries = []domain.Entry{*savedEntryFrom, *savedEntryTo}

	_, err = tx.Exec(
		ctx,
		"UPDATE wallets SET total_amount = total_amount - $1 WHERE id = $2",
		savedEntryFrom.Money.Amount(), savedEntryFrom.WalletID,
	)
	if err != nil {
		return nil, fmt.Errorf("update wallet: %w", err)
	}

	_, err = tx.Exec(
		ctx,
		"UPDATE wallets SET total_amount = total_amount + $1 WHERE id = $2",
		savedEntryTo.Money.Amount(), savedEntryTo.WalletID,
	)
	if err != nil {
		return nil, fmt.Errorf("update wallet: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit balance repo transaction: %w", err)
	}

	return savedTransaction, nil
}
