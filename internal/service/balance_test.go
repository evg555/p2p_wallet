package service

import (
	"context"
	"testing"

	"p2p_wallet/internal/domain"
	"p2p_wallet/internal/domain/helpers/authctx"
	"p2p_wallet/internal/dto"
	"p2p_wallet/internal/errs"
	"p2p_wallet/internal/service/mocks"

	"github.com/stretchr/testify/assert"
)

type balanceRepoStub struct {
	transaction *domain.Transaction
	err         error
	gotTx       *domain.Transaction
}

func (b *balanceRepoStub) CreateTransaction(_ context.Context, tx *domain.Transaction) (*domain.Transaction, error) {
	b.gotTx = tx
	return b.transaction, b.err
}

func TestTransfer(t *testing.T) {
	t.Run("from wallet belongs to another user", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), authctx.CurrentUserKey, domain.User{ID: 1})
		walletRepo := mocks.NewMockWalletRepo(t)
		logger := &testLogger{}
		svc := NewBalanceService(logger, walletRepo, &balanceRepoStub{})

		walletRepo.EXPECT().FindByID(ctx, domain.WalletID(10)).Return(&domain.Wallet{
			ID:          domain.WalletID(10),
			UserID:      domain.UserID(2),
			Currency:    domain.CurrencyUSD,
			TotalAmount: 500,
		}, nil)

		got, err := svc.Transfer(ctx, dto.TransferBalanceInput{
			IdempotencyKey: "transfer-123",
			FromWalletID:   10,
			ToWalletID:     20,
			Amount:         100,
		})

		assert.Nil(t, got)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrWalletMismatch)
	})

	t.Run("delegates balance validation to repository", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), authctx.CurrentUserKey, domain.User{ID: 1})
		walletRepo := mocks.NewMockWalletRepo(t)
		logger := &testLogger{}
		balanceRepo := &balanceRepoStub{err: errs.ErrNotEnoughMoney}
		svc := NewBalanceService(logger, walletRepo, balanceRepo)

		walletRepo.EXPECT().FindByID(ctx, domain.WalletID(10)).Return(&domain.Wallet{
			ID:          domain.WalletID(10),
			UserID:      domain.UserID(1),
			Currency:    domain.CurrencyUSD,
			TotalAmount: 50,
			HeldAmount:  0,
		}, nil)
		walletRepo.EXPECT().FindByID(ctx, domain.WalletID(20)).Return(&domain.Wallet{
			ID:          domain.WalletID(20),
			UserID:      domain.UserID(2),
			Currency:    domain.CurrencyUSD,
			TotalAmount: 0,
			HeldAmount:  0,
		}, nil)

		got, err := svc.Transfer(ctx, dto.TransferBalanceInput{
			IdempotencyKey: "transfer-123",
			FromWalletID:   10,
			ToWalletID:     20,
			Amount:         100,
		})

		assert.Nil(t, got)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrNotEnoughMoney)
		assert.NotNil(t, balanceRepo.gotTx)
		assert.Equal(t, "transfer-123", balanceRepo.gotTx.IdempotencyKey)
		assert.Len(t, balanceRepo.gotTx.Entries, 2)
	})
}
