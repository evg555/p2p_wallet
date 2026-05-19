package repository

import (
	"context"
	"testing"

	"p2p_wallet/internal/domain"
	"p2p_wallet/internal/errs"

	"github.com/stretchr/testify/require"
)

func TestBalancePostgresRepoCreateTransaction(t *testing.T) {
	t.Run("creates ledger records and updates snapshots atomically", func(t *testing.T) {
		ctx, userRepo := newPostgresRepoForIntegration(t)
		walletRepo := &walletPostgresRepo{client: userRepo.client}
		repo := &balancePostgresRepo{client: userRepo.client}

		fromUser := saveUserForWalletTest(t, ctx, userRepo, "balance-create-user-from")
		toUser := saveUserForWalletTest(t, ctx, userRepo, "balance-create-user-to")
		fromWallet := saveWalletForBalanceTest(t, ctx, walletRepo, fromUser.ID, "USD")
		toWallet := saveWalletForBalanceTest(t, ctx, walletRepo, toUser.ID, "USD")

		setWalletSnapshotAmounts(t, ctx, userRepo, fromWallet.ID, 1000, 100)
		setWalletSnapshotAmounts(t, ctx, userRepo, toWallet.ID, 200, 0)

		fromWallet.TotalAmount = 1000
		fromWallet.HeldAmount = 100
		toWallet.TotalAmount = 200
		toWallet.HeldAmount = 0

		transaction, err := domain.NewTransaction("transfer-unique-1", 300, fromWallet, toWallet)
		require.NoError(t, err)

		saved, err := repo.CreateTransaction(ctx, transaction)
		require.NoError(t, err)
		require.NotNil(t, saved)
		require.Positive(t, saved.ID)
		require.Equal(t, transaction.IdempotencyKey, saved.IdempotencyKey)
		require.Len(t, saved.Entries, 2)
		require.Positive(t, saved.Entries[0].ID)
		require.Positive(t, saved.Entries[1].ID)

		fromTotal, fromHeld := getWalletSnapshotAmounts(t, ctx, userRepo, fromWallet.ID)
		toTotal, toHeld := getWalletSnapshotAmounts(t, ctx, userRepo, toWallet.ID)

		require.Equal(t, int64(700), fromTotal)
		require.Equal(t, int64(100), fromHeld)
		require.Equal(t, int64(500), toTotal)
		require.Equal(t, int64(0), toHeld)
	})

	t.Run("rechecks available money under lock", func(t *testing.T) {
		ctx, userRepo := newPostgresRepoForIntegration(t)
		walletRepo := &walletPostgresRepo{client: userRepo.client}
		repo := &balancePostgresRepo{client: userRepo.client}

		fromUser := saveUserForWalletTest(t, ctx, userRepo, "balance-race-user-from")
		toUser := saveUserForWalletTest(t, ctx, userRepo, "balance-race-user-to")
		fromWallet := saveWalletForBalanceTest(t, ctx, walletRepo, fromUser.ID, "USD")
		toWallet := saveWalletForBalanceTest(t, ctx, walletRepo, toUser.ID, "USD")

		fromWallet.TotalAmount = 1000
		fromWallet.HeldAmount = 0
		toWallet.TotalAmount = 0
		toWallet.HeldAmount = 0

		transaction, err := domain.NewTransaction("transfer-unique-2", 300, fromWallet, toWallet)
		require.NoError(t, err)

		setWalletSnapshotAmounts(t, ctx, userRepo, fromWallet.ID, 200, 0)
		setWalletSnapshotAmounts(t, ctx, userRepo, toWallet.ID, 0, 0)

		saved, err := repo.CreateTransaction(ctx, transaction)
		require.ErrorIs(t, err, errs.ErrNotEnoughMoney)
		require.Nil(t, saved)

		var transactionCount int
		err = userRepo.client.QueryRow(ctx, "SELECT COUNT(*) FROM ledger_transactions").Scan(&transactionCount)
		require.NoError(t, err)
		require.Zero(t, transactionCount)
	})

	t.Run("returns existing transaction for duplicate idempotency key", func(t *testing.T) {
		ctx, userRepo := newPostgresRepoForIntegration(t)
		walletRepo := &walletPostgresRepo{client: userRepo.client}
		repo := &balancePostgresRepo{client: userRepo.client}

		fromUser := saveUserForWalletTest(t, ctx, userRepo, "balance-idemp-user-from")
		toUser := saveUserForWalletTest(t, ctx, userRepo, "balance-idemp-user-to")
		fromWallet := saveWalletForBalanceTest(t, ctx, walletRepo, fromUser.ID, "USD")
		toWallet := saveWalletForBalanceTest(t, ctx, walletRepo, toUser.ID, "USD")

		setWalletSnapshotAmounts(t, ctx, userRepo, fromWallet.ID, 1000, 0)
		setWalletSnapshotAmounts(t, ctx, userRepo, toWallet.ID, 100, 0)

		fromWallet.TotalAmount = 1000
		toWallet.TotalAmount = 100

		firstTx, err := domain.NewTransaction("transfer-idemp-1", 300, fromWallet, toWallet)
		require.NoError(t, err)

		firstSaved, err := repo.CreateTransaction(ctx, firstTx)
		require.NoError(t, err)

		duplicateTx, err := domain.NewTransaction("transfer-idemp-1", 300, fromWallet, toWallet)
		require.NoError(t, err)

		duplicateSaved, err := repo.CreateTransaction(ctx, duplicateTx)
		require.NoError(t, err)
		require.NotNil(t, duplicateSaved)
		require.Equal(t, firstSaved.ID, duplicateSaved.ID)
		require.Equal(t, firstSaved.IdempotencyKey, duplicateSaved.IdempotencyKey)
		require.Equal(t, firstSaved.CreatedAt, duplicateSaved.CreatedAt)
		require.Len(t, duplicateSaved.Entries, 2)
		require.Equal(t, firstSaved.Entries, duplicateSaved.Entries)

		fromTotal, fromHeld := getWalletSnapshotAmounts(t, ctx, userRepo, fromWallet.ID)
		toTotal, toHeld := getWalletSnapshotAmounts(t, ctx, userRepo, toWallet.ID)
		require.Equal(t, int64(700), fromTotal)
		require.Equal(t, int64(0), fromHeld)
		require.Equal(t, int64(400), toTotal)
		require.Equal(t, int64(0), toHeld)

		var transactionCount int
		err = userRepo.client.QueryRow(ctx, "SELECT COUNT(*) FROM ledger_transactions").Scan(&transactionCount)
		require.NoError(t, err)
		require.Equal(t, 1, transactionCount)
	})

	t.Run("returns existing transaction before balance revalidation on repeated request", func(t *testing.T) {
		ctx, userRepo := newPostgresRepoForIntegration(t)
		walletRepo := &walletPostgresRepo{client: userRepo.client}
		repo := &balancePostgresRepo{client: userRepo.client}

		fromUser := saveUserForWalletTest(t, ctx, userRepo, "balance-repeat-user-from")
		toUser := saveUserForWalletTest(t, ctx, userRepo, "balance-repeat-user-to")
		fromWallet := saveWalletForBalanceTest(t, ctx, walletRepo, fromUser.ID, "USD")
		toWallet := saveWalletForBalanceTest(t, ctx, walletRepo, toUser.ID, "USD")

		setWalletSnapshotAmounts(t, ctx, userRepo, fromWallet.ID, 1000, 0)
		setWalletSnapshotAmounts(t, ctx, userRepo, toWallet.ID, 100, 0)

		fromWallet.TotalAmount = 1000
		toWallet.TotalAmount = 100

		firstTx, err := domain.NewTransaction("transfer-idemp-repeat-1", 300, fromWallet, toWallet)
		require.NoError(t, err)

		firstSaved, err := repo.CreateTransaction(ctx, firstTx)
		require.NoError(t, err)

		setWalletSnapshotAmounts(t, ctx, userRepo, fromWallet.ID, 0, 0)
		setWalletSnapshotAmounts(t, ctx, userRepo, toWallet.ID, 9999, 0)

		repeatedTx, err := domain.NewTransaction("transfer-idemp-repeat-1", 300, fromWallet, toWallet)
		require.NoError(t, err)

		repeatedSaved, err := repo.CreateTransaction(ctx, repeatedTx)
		require.NoError(t, err)
		require.NotNil(t, repeatedSaved)
		require.Equal(t, firstSaved.ID, repeatedSaved.ID)
		require.Equal(t, firstSaved.IdempotencyKey, repeatedSaved.IdempotencyKey)
		require.Equal(t, firstSaved.CreatedAt, repeatedSaved.CreatedAt)
		require.Equal(t, firstSaved.Entries, repeatedSaved.Entries)

		var transactionCount int
		err = userRepo.client.QueryRow(ctx, "SELECT COUNT(*) FROM ledger_transactions").Scan(&transactionCount)
		require.NoError(t, err)
		require.Equal(t, 1, transactionCount)
	})
}

func saveWalletForBalanceTest(
	t *testing.T,
	ctx context.Context,
	repo *walletPostgresRepo,
	userID domain.UserID,
	currency string,
) *domain.Wallet {
	t.Helper()

	wallet, err := domain.NewWallet(userID, currency)
	require.NoError(t, err)

	saved, err := repo.Save(ctx, wallet)
	require.NoError(t, err)

	return saved
}

func setWalletSnapshotAmounts(
	t *testing.T,
	ctx context.Context,
	userRepo *userPostgresRepo,
	walletID domain.WalletID,
	totalAmount int64,
	heldAmount int64,
) {
	t.Helper()

	_, err := userRepo.client.Exec(
		ctx,
		"UPDATE wallet_balance_snapshots SET total_amount = $1, held_amount = $2, updated_at = now() WHERE wallet_id = $3",
		totalAmount, heldAmount, walletID,
	)
	require.NoError(t, err)
}

func getWalletSnapshotAmounts(
	t *testing.T,
	ctx context.Context,
	userRepo *userPostgresRepo,
	walletID domain.WalletID,
) (int64, int64) {
	t.Helper()

	var totalAmount int64
	var heldAmount int64

	err := userRepo.client.QueryRow(
		ctx,
		"SELECT total_amount, held_amount FROM wallet_balance_snapshots WHERE wallet_id = $1",
		walletID,
	).Scan(&totalAmount, &heldAmount)
	require.NoError(t, err)

	return totalAmount, heldAmount
}
