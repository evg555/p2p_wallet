package repository

import (
	"context"
	"errors"
	"fmt"

	"p2p_wallet/internal/domain"
	"p2p_wallet/internal/errs"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type walletPostgresRepo struct {
	client *pgxpool.Pool
}

func NewWalletPostgresRepo(client *pgxpool.Pool) *walletPostgresRepo {
	return &walletPostgresRepo{
		client: client,
	}
}

func (w *walletPostgresRepo) Save(ctx context.Context, wallet *domain.Wallet) (*domain.Wallet, error) {
	tx, err := w.client.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin wallet save transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	walletQuery, walletArgs, err := psql.Insert("wallets").
		Columns("currency", "status", "user_id").
		Values(wallet.Currency.String(), wallet.Status.String(), wallet.UserID).
		Suffix("RETURNING id, created_at, updated_at").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build save wallet query: %w", err)
	}

	savedWallet := &domain.Wallet{
		Currency:    wallet.Currency,
		Status:      wallet.Status,
		UserID:      wallet.UserID,
		HeldAmount:  wallet.HeldAmount,
		TotalAmount: wallet.TotalAmount,
	}

	err = tx.QueryRow(ctx, walletQuery, walletArgs...).Scan(
		&savedWallet.ID,
		&savedWallet.CreatedAt,
		&savedWallet.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, errs.ErrWalletAlreadyExist
		}

		return nil, fmt.Errorf("save wallet: %w", err)
	}

	snapshotQuery, snapshotArgs, err := psql.Insert("wallet_balance_snapshots").
		Columns("wallet_id", "currency").
		Values(savedWallet.ID, savedWallet.Currency.String()).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build save wallet snapshot query: %w", err)
	}

	if _, err = tx.Exec(ctx, snapshotQuery, snapshotArgs...); err != nil {
		return nil, fmt.Errorf("save wallet snapshot: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit wallet save transaction: %w", err)
	}

	return savedWallet, nil
}

func (w *walletPostgresRepo) FindByUserID(ctx context.Context, userID domain.UserID) ([]*domain.Wallet, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	query, args, err := psql.
		Select(
			"w.id",
			"w.currency",
			"w.status",
			"w.user_id",
			"s.total_amount",
			"s.held_amount",
			"w.created_at",
			"s.updated_at",
		).
		From("wallets w").
		Join("wallet_balance_snapshots s ON s.wallet_id = w.id").
		Where(sq.Eq{"w.user_id": userID}).
		OrderBy("w.id ASC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build find wallets by user id query: %w", err)
	}

	rows, err := w.client.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("find wallets by user id: %w", err)
	}
	defer rows.Close()

	wallets := make([]*domain.Wallet, 0)
	for rows.Next() {
		var (
			currency string
			status   string
			wallet   domain.Wallet
		)

		if err = rows.Scan(
			&wallet.ID,
			&currency,
			&status,
			&wallet.UserID,
			&wallet.TotalAmount,
			&wallet.HeldAmount,
			&wallet.CreatedAt,
			&wallet.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan wallet by user id: %w", err)
		}

		wallet.Currency, err = domain.NewCurrency(currency)
		if err != nil {
			return nil, fmt.Errorf("parse wallet currency: %w", err)
		}

		wallet.Status, err = newWalletStatus(status)
		if err != nil {
			return nil, fmt.Errorf("parse wallet status: %w", err)
		}

		wallets = append(wallets, &wallet)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate wallets by user id: %w", err)
	}

	return wallets, nil
}

func (w *walletPostgresRepo) FindByID(ctx context.Context, id domain.WalletID) (*domain.Wallet, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	query, args, err := psql.
		Select(
			"w.id",
			"w.currency",
			"w.status",
			"w.user_id",
			"s.total_amount",
			"s.held_amount",
			"w.created_at",
			"s.updated_at",
		).
		From("wallets w").
		Join("wallet_balance_snapshots s ON s.wallet_id = w.id").
		Where(sq.Eq{"w.id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build find wallet by id query: %w", err)
	}

	var (
		currency string
		status   string
		wallet   domain.Wallet
	)

	err = w.client.QueryRow(ctx, query, args...).Scan(
		&wallet.ID,
		&currency,
		&status,
		&wallet.UserID,
		&wallet.TotalAmount,
		&wallet.HeldAmount,
		&wallet.CreatedAt,
		&wallet.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errs.ErrWalletNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find wallet by id: %w", err)
	}

	wallet.Currency, err = domain.NewCurrency(currency)
	if err != nil {
		return nil, fmt.Errorf("parse wallet currency: %w", err)
	}

	wallet.Status, err = newWalletStatus(status)
	if err != nil {
		return nil, fmt.Errorf("parse wallet status: %w", err)
	}

	return &wallet, nil
}
