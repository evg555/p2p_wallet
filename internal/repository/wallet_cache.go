package repository

import (
	"context"
	"fmt"
	"time"

	"p2p_wallet/internal/domain"
	"p2p_wallet/internal/infra/cache"
)

const walletListTTL = time.Hour

type cachedWalletRepo struct {
	next  WalletRepo
	cache Cache
}

func NewWalletRepoWithCache(next WalletRepo) WalletRepo {
	return &cachedWalletRepo{
		next:  next,
		cache: cache.NewCache(cacheSize),
	}
}

func (r *cachedWalletRepo) Save(ctx context.Context, wallet *domain.Wallet) (*domain.Wallet, error) {
	savedWallet, err := r.next.Save(ctx, wallet)
	if err != nil {
		return nil, err
	}

	r.cache.Delete(walletListCacheKey(savedWallet.UserID))

	return savedWallet, nil
}

func (r *cachedWalletRepo) FindByUserID(ctx context.Context, userID domain.UserID) ([]*domain.Wallet, error) {
	val, ok := r.cache.Get(walletListCacheKey(userID))
	if ok {
		wallets, ok := val.([]*domain.Wallet)
		if !ok {
			return nil, fmt.Errorf("wallet list cache: %w", errWrongType)
		}

		return wallets, nil
	}

	wallets, err := r.next.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	r.cache.SetWithTTL(walletListCacheKey(userID), wallets, walletListTTL)

	return wallets, nil
}
func (r *cachedWalletRepo) FindByID(ctx context.Context, id domain.WalletID) (*domain.Wallet, error) {
	panic("implement me")
}

func walletListCacheKey(userID domain.UserID) string {
	return fmt.Sprintf("wallets:user:%d", userID)
}
