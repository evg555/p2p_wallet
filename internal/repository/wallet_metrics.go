package repository

import (
	"context"
	"time"

	"p2p_wallet/internal/domain"

	"github.com/prometheus/client_golang/prometheus"
)

type metricsWalletRepo struct {
	next WalletRepo
}

func NewWalletRepoWithMetrics(next WalletRepo) WalletRepo {
	registerDBMetricsOnce.Do(func() {
		prometheus.MustRegister(dbQueryTotal, dbQueryDuration, dbQueryErrors)
	})

	return &metricsWalletRepo{next: next}
}

func (m *metricsWalletRepo) Save(ctx context.Context, wallet *domain.Wallet) (*domain.Wallet, error) {
	const op = "wallet.save"

	start := time.Now()
	res, err := m.next.Save(ctx, wallet)
	observeDB(op, start, err)

	return res, err
}

func (m *metricsWalletRepo) FindByUserID(ctx context.Context, userID domain.UserID) ([]*domain.Wallet, error) {
	const op = "wallet.find_by_user_id"

	start := time.Now()
	res, err := m.next.FindByUserID(ctx, userID)
	observeDB(op, start, err)

	return res, err
}

func (m *metricsWalletRepo) FindByID(ctx context.Context, id domain.WalletID) (*domain.Wallet, error) {
	const op = "wallet.find_by_id"

	start := time.Now()
	res, err := m.next.FindByID(ctx, id)
	observeDB(op, start, err)

	return res, err
}
