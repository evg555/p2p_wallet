package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"p2p_wallet/internal/domain"
	"p2p_wallet/internal/dto"
	"p2p_wallet/internal/errs"
	"p2p_wallet/internal/service/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateWallet(t *testing.T) {
	t.Run("user repo error", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), domain.CtxKey(domain.SessionKey), "sid")
		userRepo := mocks.NewMockUserRepo(t)
		walletRepo := mocks.NewMockWalletRepo(t)
		sessionRepo := mocks.NewMockSessionRepo(t)
		logger := &testLogger{}
		svc := NewWalletService(logger, userRepo, walletRepo, sessionRepo)

		repoErr := errors.New("db down")
		userRepo.EXPECT().GetByID(ctx, domain.UserID(1)).Return(nil, repoErr)

		got, err := svc.CreateWallet(ctx, dto.CreateWalletInput{UserID: 1, Currency: "EUR"})
		assert.Nil(t, got)
		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "failed to get user by id"))
		assert.ErrorIs(t, err, repoErr)
	})

	t.Run("session not found in context", func(t *testing.T) {
		ctx := context.Background()
		userRepo := mocks.NewMockUserRepo(t)
		walletRepo := mocks.NewMockWalletRepo(t)
		sessionRepo := mocks.NewMockSessionRepo(t)
		logger := &testLogger{}
		svc := NewWalletService(logger, userRepo, walletRepo, sessionRepo)

		userRepo.EXPECT().GetByID(ctx, domain.UserID(1)).Return(&domain.User{ID: 1}, nil)

		got, err := svc.CreateWallet(ctx, dto.CreateWalletInput{UserID: 1, Currency: "EUR"})
		assert.Nil(t, got)
		assert.ErrorIs(t, err, errs.ErrSessionNotFound)
	})

	t.Run("invalid currency", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), domain.CtxKey(domain.SessionKey), "sid")
		userRepo := mocks.NewMockUserRepo(t)
		walletRepo := mocks.NewMockWalletRepo(t)
		sessionRepo := mocks.NewMockSessionRepo(t)
		logger := &testLogger{}
		svc := NewWalletService(logger, userRepo, walletRepo, sessionRepo)

		userRepo.EXPECT().GetByID(ctx, domain.UserID(1)).Return(&domain.User{ID: 1}, nil)
		sessionRepo.EXPECT().Get(domain.SessionID("sid")).Return(&domain.Session{
			ID:     domain.SessionID("sid"),
			UserID: domain.UserID(1),
		}, nil)

		got, err := svc.CreateWallet(ctx, dto.CreateWalletInput{UserID: 1, Currency: "RUB"})
		assert.Nil(t, got)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "wallet: unknown currency RUB")
	})

	t.Run("wallet already exists", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), domain.CtxKey(domain.SessionKey), "sid")
		userRepo := mocks.NewMockUserRepo(t)
		walletRepo := mocks.NewMockWalletRepo(t)
		sessionRepo := mocks.NewMockSessionRepo(t)
		logger := &testLogger{}
		svc := NewWalletService(logger, userRepo, walletRepo, sessionRepo)

		userRepo.EXPECT().GetByID(ctx, domain.UserID(1)).Return(&domain.User{ID: 1}, nil)
		sessionRepo.EXPECT().Get(domain.SessionID("sid")).Return(&domain.Session{
			ID:     domain.SessionID("sid"),
			UserID: domain.UserID(1),
		}, nil)
		walletRepo.EXPECT().Save(ctx, mock.AnythingOfType("*domain.Wallet")).Return(nil, errs.ErrWalletAlreadyExist)

		got, err := svc.CreateWallet(ctx, dto.CreateWalletInput{UserID: 1, Currency: "EUR"})
		assert.Nil(t, got)
		assert.ErrorIs(t, err, errs.ErrWalletAlreadyExist)
	})

	t.Run("success", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), domain.CtxKey(domain.SessionKey), "sid")
		userRepo := mocks.NewMockUserRepo(t)
		walletRepo := mocks.NewMockWalletRepo(t)
		sessionRepo := mocks.NewMockSessionRepo(t)
		logger := &testLogger{}
		svc := NewWalletService(logger, userRepo, walletRepo, sessionRepo)

		userRepo.EXPECT().GetByID(ctx, domain.UserID(1)).Return(&domain.User{ID: 1}, nil)
		sessionRepo.EXPECT().Get(domain.SessionID("sid")).Return(&domain.Session{
			ID:     domain.SessionID("sid"),
			UserID: domain.UserID(1),
		}, nil)
		walletRepo.EXPECT().Save(ctx, mock.MatchedBy(func(w *domain.Wallet) bool {
			return w != nil &&
				w.ID == 0 &&
				w.UserID == domain.UserID(1) &&
				w.Currency == domain.CurrencyEUR &&
				w.Status == domain.StatusActive &&
				w.TotalAmount == 0 &&
				w.HeldAmount == 0 &&
				!w.CreatedAt.IsZero() &&
				w.UpdatedAt == nil
		})).RunAndReturn(func(_ context.Context, w *domain.Wallet) (*domain.Wallet, error) {
			w.ID = domain.WalletID(10)
			return w, nil
		})

		got, err := svc.CreateWallet(ctx, dto.CreateWalletInput{UserID: 1, Currency: "EUR"})
		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, domain.WalletID(10), got.ID)
		assert.Equal(t, domain.UserID(1), got.UserID)
		assert.Equal(t, domain.CurrencyEUR, got.Currency)
		assert.Equal(t, domain.StatusActive, got.Status)
	})
}

func TestListWallets(t *testing.T) {
	t.Run("user not found", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), domain.CtxKey(domain.SessionKey), "sid")
		userRepo := mocks.NewMockUserRepo(t)
		walletRepo := mocks.NewMockWalletRepo(t)
		sessionRepo := mocks.NewMockSessionRepo(t)
		logger := &testLogger{}
		svc := NewWalletService(logger, userRepo, walletRepo, sessionRepo)

		userRepo.EXPECT().GetByID(ctx, domain.UserID(1)).Return(nil, errs.ErrUserNotFound)

		got, err := svc.ListWallets(ctx, 1)
		assert.Nil(t, got)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrUserNotFound)
	})

	t.Run("wallet repo error", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), domain.CtxKey(domain.SessionKey), "sid")
		userRepo := mocks.NewMockUserRepo(t)
		walletRepo := mocks.NewMockWalletRepo(t)
		sessionRepo := mocks.NewMockSessionRepo(t)
		logger := &testLogger{}
		svc := NewWalletService(logger, userRepo, walletRepo, sessionRepo)

		userRepo.EXPECT().GetByID(ctx, domain.UserID(1)).Return(&domain.User{ID: 1}, nil)
		sessionRepo.EXPECT().Get(domain.SessionID("sid")).Return(&domain.Session{
			ID:     domain.SessionID("sid"),
			UserID: domain.UserID(1),
		}, nil)

		repoErr := errors.New("db fail")
		walletRepo.EXPECT().FindByUserID(ctx, domain.UserID(1)).Return(nil, repoErr)

		got, err := svc.ListWallets(ctx, 1)
		assert.Nil(t, got)
		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "failed to get wallets by user id"))
		assert.ErrorIs(t, err, repoErr)
	})

	t.Run("success", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), domain.CtxKey(domain.SessionKey), "sid")
		userRepo := mocks.NewMockUserRepo(t)
		walletRepo := mocks.NewMockWalletRepo(t)
		sessionRepo := mocks.NewMockSessionRepo(t)
		logger := &testLogger{}
		svc := NewWalletService(logger, userRepo, walletRepo, sessionRepo)

		userRepo.EXPECT().GetByID(ctx, domain.UserID(1)).Return(&domain.User{ID: 1}, nil)
		sessionRepo.EXPECT().Get(domain.SessionID("sid")).Return(&domain.Session{
			ID:     domain.SessionID("sid"),
			UserID: domain.UserID(1),
		}, nil)

		expected := []*domain.Wallet{
			{ID: 1, UserID: 1, Currency: domain.CurrencyEUR, Status: domain.StatusActive},
			{ID: 2, UserID: 1, Currency: domain.CurrencyUSD, Status: domain.StatusBlocked},
		}
		walletRepo.EXPECT().FindByUserID(ctx, domain.UserID(1)).Return(expected, nil)

		got, err := svc.ListWallets(ctx, 1)
		assert.NoError(t, err)
		assert.Equal(t, expected, got)
	})
}
