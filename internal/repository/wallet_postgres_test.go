package repository

import (
	"context"
	"testing"

	"p2p_wallet/internal/domain"
	"p2p_wallet/internal/errs"

	"github.com/stretchr/testify/require"
)

func TestWalletPostgresRepoSave(t *testing.T) {
	ctx, userRepo := newPostgresRepoForIntegration(t)
	repo := &walletPostgresRepo{client: userRepo.client}

	user := saveUserForWalletTest(t, ctx, userRepo, "wallet-save-user")

	input, err := domain.NewWallet(user.ID, "USD")
	require.NoError(t, err)

	saved, err := repo.Save(ctx, input)
	require.NoError(t, err)
	require.NotNil(t, saved)
	require.Positive(t, saved.ID)
	require.Equal(t, input.Currency, saved.Currency)
	require.Equal(t, input.Status, saved.Status)
	require.Equal(t, input.UserID, saved.UserID)
	require.Zero(t, saved.TotalAmount)
	require.Zero(t, saved.HeldAmount)
	require.False(t, saved.CreatedAt.IsZero())
	require.Nil(t, saved.UpdatedAt)

	duplicateInput, err := domain.NewWallet(user.ID, "USD")
	require.NoError(t, err)

	duplicateSaved, err := repo.Save(ctx, duplicateInput)
	require.ErrorIs(t, err, errs.ErrWalletAlreadyExist)
	require.Nil(t, duplicateSaved)
}

func TestWalletPostgresRepoFindByUserID(t *testing.T) {
	t.Run("wallets exist", func(t *testing.T) {
		ctx, userRepo := newPostgresRepoForIntegration(t)
		repo := &walletPostgresRepo{client: userRepo.client}

		user := saveUserForWalletTest(t, ctx, userRepo, "wallet-find-user")

		usdWallet, err := domain.NewWallet(user.ID, "USD")
		require.NoError(t, err)

		eurWallet, err := domain.NewWallet(user.ID, "EUR")
		require.NoError(t, err)

		savedUSD, err := repo.Save(ctx, usdWallet)
		require.NoError(t, err)

		savedEUR, err := repo.Save(ctx, eurWallet)
		require.NoError(t, err)

		got, err := repo.FindByUserID(ctx, user.ID)
		require.NoError(t, err)
		require.Len(t, got, 2)

		require.Equal(t, savedUSD.ID, got[0].ID)
		require.Equal(t, savedUSD.Currency, got[0].Currency)
		require.Equal(t, savedUSD.Status, got[0].Status)
		require.Equal(t, savedUSD.UserID, got[0].UserID)
		require.Zero(t, got[0].TotalAmount)
		require.Zero(t, got[0].HeldAmount)
		require.False(t, got[0].CreatedAt.IsZero())
		require.NotNil(t, got[0].UpdatedAt)

		require.Equal(t, savedEUR.ID, got[1].ID)
		require.Equal(t, savedEUR.Currency, got[1].Currency)
		require.Equal(t, savedEUR.Status, got[1].Status)
		require.Equal(t, savedEUR.UserID, got[1].UserID)
		require.Zero(t, got[1].TotalAmount)
		require.Zero(t, got[1].HeldAmount)
		require.False(t, got[1].CreatedAt.IsZero())
		require.NotNil(t, got[1].UpdatedAt)
	})

	t.Run("wallets do not exist", func(t *testing.T) {
		ctx, userRepo := newPostgresRepoForIntegration(t)
		repo := &walletPostgresRepo{client: userRepo.client}

		got, err := repo.FindByUserID(ctx, domain.UserID(9999999999))
		require.NoError(t, err)
		require.Empty(t, got)
	})
}

func saveUserForWalletTest(
	t *testing.T,
	ctx context.Context,
	repo *userPostgresRepo,
	login string,
) *domain.User {
	t.Helper()

	user, err := domain.NewUser(login, "secret", "John", "Doe")
	require.NoError(t, err)

	saved, err := repo.Save(ctx, user)
	require.NoError(t, err)

	return saved
}
