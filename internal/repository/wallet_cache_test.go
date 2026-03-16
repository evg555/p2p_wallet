package repository

import (
	"context"
	"errors"
	"testing"

	"p2p_wallet/internal/domain"
	"p2p_wallet/internal/infra/cache"

	"github.com/stretchr/testify/require"
)

func TestCachedWalletRepoFindByUserID(t *testing.T) {
	t.Run("cache miss then hit", func(t *testing.T) {
		next := &stubWalletRepo{
			findByUserIDResult: []*domain.Wallet{
				{ID: 1, UserID: 42, Currency: domain.CurrencyEUR, Status: domain.StatusActive},
			},
		}
		repo := &cachedWalletRepo{
			next:  next,
			cache: cache.NewCache(10),
		}

		gotFirst, err := repo.FindByUserID(context.Background(), domain.UserID(42))
		require.NoError(t, err)
		require.Len(t, gotFirst, 1)
		require.Equal(t, 1, next.findByUserIDCalls)

		gotSecond, err := repo.FindByUserID(context.Background(), domain.UserID(42))
		require.NoError(t, err)
		require.Equal(t, gotFirst, gotSecond)
		require.Equal(t, 1, next.findByUserIDCalls)
	})

	t.Run("different users use different cache keys", func(t *testing.T) {
		next := &stubWalletRepo{
			findByUserIDResults: map[domain.UserID][]*domain.Wallet{
				1: {{ID: 1, UserID: 1, Currency: domain.CurrencyEUR, Status: domain.StatusActive}},
				2: {{ID: 2, UserID: 2, Currency: domain.CurrencyUSD, Status: domain.StatusBlocked}},
			},
		}
		repo := &cachedWalletRepo{
			next:  next,
			cache: cache.NewCache(10),
		}

		gotFirst, err := repo.FindByUserID(context.Background(), domain.UserID(1))
		require.NoError(t, err)
		require.Len(t, gotFirst, 1)

		gotSecond, err := repo.FindByUserID(context.Background(), domain.UserID(2))
		require.NoError(t, err)
		require.Len(t, gotSecond, 1)

		_, err = repo.FindByUserID(context.Background(), domain.UserID(1))
		require.NoError(t, err)

		require.Equal(t, 2, next.findByUserIDCalls)
	})

	t.Run("find error is not cached", func(t *testing.T) {
		next := &stubWalletRepo{
			findByUserIDErr: errors.New("db fail"),
		}
		repo := &cachedWalletRepo{
			next:  next,
			cache: cache.NewCache(10),
		}

		got, err := repo.FindByUserID(context.Background(), domain.UserID(42))
		require.Nil(t, got)
		require.EqualError(t, err, "db fail")

		got, err = repo.FindByUserID(context.Background(), domain.UserID(42))
		require.Nil(t, got)
		require.EqualError(t, err, "db fail")
		require.Equal(t, 2, next.findByUserIDCalls)
	})
}

func TestCachedWalletRepoSave(t *testing.T) {
	t.Run("save invalidates cached wallets for user", func(t *testing.T) {
		userID := domain.UserID(42)
		next := &stubWalletRepo{
			findByUserIDResults: map[domain.UserID][]*domain.Wallet{
				userID: {{ID: 1, UserID: userID, Currency: domain.CurrencyEUR, Status: domain.StatusActive}},
			},
			saveResult: &domain.Wallet{ID: 2, UserID: userID, Currency: domain.CurrencyUSD, Status: domain.StatusActive},
		}
		repo := &cachedWalletRepo{
			next:  next,
			cache: cache.NewCache(10),
		}

		_, err := repo.FindByUserID(context.Background(), userID)
		require.NoError(t, err)
		require.Equal(t, 1, next.findByUserIDCalls)

		_, err = repo.Save(context.Background(), &domain.Wallet{UserID: userID, Currency: domain.CurrencyUSD})
		require.NoError(t, err)
		require.Equal(t, 1, next.saveCalls)

		_, err = repo.FindByUserID(context.Background(), userID)
		require.NoError(t, err)
		require.Equal(t, 2, next.findByUserIDCalls)
	})

	t.Run("save error does not invalidate cache", func(t *testing.T) {
		userID := domain.UserID(42)
		next := &stubWalletRepo{
			findByUserIDResults: map[domain.UserID][]*domain.Wallet{
				userID: {{ID: 1, UserID: userID, Currency: domain.CurrencyEUR, Status: domain.StatusActive}},
			},
			saveErr: errors.New("insert fail"),
		}
		repo := &cachedWalletRepo{
			next:  next,
			cache: cache.NewCache(10),
		}

		_, err := repo.FindByUserID(context.Background(), userID)
		require.NoError(t, err)
		require.Equal(t, 1, next.findByUserIDCalls)

		saved, err := repo.Save(context.Background(), &domain.Wallet{UserID: userID, Currency: domain.CurrencyUSD})
		require.Nil(t, saved)
		require.EqualError(t, err, "insert fail")

		_, err = repo.FindByUserID(context.Background(), userID)
		require.NoError(t, err)
		require.Equal(t, 1, next.findByUserIDCalls)
	})
}

type stubWalletRepo struct {
	findByUserIDCalls   int
	saveCalls           int
	findByUserIDResult  []*domain.Wallet
	findByUserIDResults map[domain.UserID][]*domain.Wallet
	findByUserIDErr     error
	saveResult          *domain.Wallet
	saveErr             error
}

func (s *stubWalletRepo) Save(_ context.Context, _ *domain.Wallet) (*domain.Wallet, error) {
	s.saveCalls++
	if s.saveErr != nil {
		return nil, s.saveErr
	}

	return s.saveResult, nil
}

func (s *stubWalletRepo) FindByUserID(_ context.Context, userID domain.UserID) ([]*domain.Wallet, error) {
	s.findByUserIDCalls++
	if s.findByUserIDErr != nil {
		return nil, s.findByUserIDErr
	}

	if s.findByUserIDResults != nil {
		return s.findByUserIDResults[userID], nil
	}

	return s.findByUserIDResult, nil
}
