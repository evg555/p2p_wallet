package repository

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"p2p_wallet/internal/domain"
	"p2p_wallet/internal/errs"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
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

	snapshots, err := lockWalletSnapshots(ctx, tx, transaction.Entries)
	if err != nil {
		return nil, fmt.Errorf("lock wallet snapshots: %w", err)
	}

	if err = validateLockedSnapshots(transaction.Entries, snapshots); err != nil {
		return nil, err
	}

	query, args, err := psql.Insert("ledger_transactions").
		Columns("type", "reference_id", "idemp_key", "created_at").
		Values(domain.TypeTransfer, nil, transaction.IdempotencyKey, transaction.CreatedAt).
		Suffix("RETURNING id, created_at").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build insert transaction query: %w", err)
	}

	savedTransaction := &domain.Transaction{
		IdempotencyKey: transaction.IdempotencyKey,
		Entries:        transaction.Entries,
	}

	err = tx.QueryRow(ctx, query, args...).Scan(&savedTransaction.ID, &savedTransaction.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
				return nil, fmt.Errorf("rollback duplicate transaction insert: %w", rollbackErr)
			}

			return b.getTransactionByIdempotencyKey(ctx, transaction.IdempotencyKey)
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
		"UPDATE wallet_balance_snapshots SET total_amount = total_amount + $1, updated_at = now() WHERE wallet_id = $2",
		savedEntryFrom.Money.Amount(), savedEntryFrom.WalletID,
	)
	if err != nil {
		return nil, fmt.Errorf("update wallet: %w", err)
	}

	_, err = tx.Exec(
		ctx,
		"UPDATE wallet_balance_snapshots SET total_amount = total_amount + $1, updated_at = now() WHERE wallet_id = $2",
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

func (b *balancePostgresRepo) getTransactionByIdempotencyKey(ctx context.Context, idempotencyKey string) (*domain.Transaction, error) {
	rows, err := b.client.Query(
		ctx,
		`SELECT lt.id, lt.created_at, le.id, le.wallet_id, le.amount, le.currency
		FROM ledger_transactions lt
		JOIN ledger_entries le ON le.transaction_id = lt.id
		WHERE lt.idemp_key = $1
		ORDER BY le.id ASC`,
		idempotencyKey,
	)
	if err != nil {
		return nil, fmt.Errorf("select transaction by idempotency key: %w", err)
	}
	defer rows.Close()

	var transaction *domain.Transaction
	for rows.Next() {
		var (
			entryID       int64
			walletID      int64
			amount        int64
			currency      string
			transactionID int64
			createdAt     time.Time
		)

		if err = rows.Scan(&transactionID, &createdAt, &entryID, &walletID, &amount, &currency); err != nil {
			return nil, fmt.Errorf("scan transaction by idempotency key: %w", err)
		}

		if transaction == nil {
			transaction = &domain.Transaction{
				ID:             transactionID,
				IdempotencyKey: idempotencyKey,
				CreatedAt:      createdAt,
			}
		}

		entryCurrency, currencyErr := domain.NewCurrency(currency)
		if currencyErr != nil {
			return nil, fmt.Errorf("parse transaction currency: %w", currencyErr)
		}

		transaction.Entries = append(transaction.Entries, domain.Entry{
			ID:       entryID,
			WalletID: domain.WalletID(walletID),
			Money:    domain.NewMoney(amount, entryCurrency),
		})
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate transaction by idempotency key: %w", err)
	}

	if transaction == nil {
		return nil, errs.ErrTransactionAlreadyCreated
	}

	return transaction, nil
}

type walletSnapshot struct {
	WalletID    domain.WalletID
	TotalAmount int64
	HeldAmount  int64
}

func lockWalletSnapshots(ctx context.Context, tx pgx.Tx, entries []domain.Entry) (map[domain.WalletID]walletSnapshot, error) {
	walletIDs := make([]int64, 0, len(entries))
	seen := make(map[domain.WalletID]struct{}, len(entries))

	for _, entry := range entries {
		if _, ok := seen[entry.WalletID]; ok {
			continue
		}
		seen[entry.WalletID] = struct{}{}
		walletIDs = append(walletIDs, entry.WalletID.Int64())
	}

	sort.Slice(walletIDs, func(i, j int) bool {
		return walletIDs[i] < walletIDs[j]
	})

	rows, err := tx.Query(
		ctx,
		`SELECT wallet_id, total_amount, held_amount
		FROM wallet_balance_snapshots
		WHERE wallet_id = ANY($1)
		ORDER BY wallet_id ASC
		FOR UPDATE`,
		walletIDs,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	snapshots := make(map[domain.WalletID]walletSnapshot, len(walletIDs))
	for rows.Next() {
		var snapshot walletSnapshot
		if err = rows.Scan(&snapshot.WalletID, &snapshot.TotalAmount, &snapshot.HeldAmount); err != nil {
			return nil, err
		}
		snapshots[snapshot.WalletID] = snapshot
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if _, ok := snapshots[entry.WalletID]; !ok {
			return nil, errs.ErrWalletNotFound
		}
	}

	return snapshots, nil
}

func validateLockedSnapshots(entries []domain.Entry, snapshots map[domain.WalletID]walletSnapshot) error {
	for _, entry := range entries {
		if entry.Money.Amount() >= 0 {
			continue
		}

		snapshot := snapshots[entry.WalletID]
		if snapshot.TotalAmount-snapshot.HeldAmount < -entry.Money.Amount() {
			return errs.ErrNotEnoughMoney
		}
	}

	return nil
}
