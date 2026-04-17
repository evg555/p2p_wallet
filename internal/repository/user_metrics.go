package repository

import (
	"context"
	"time"

	"p2p_wallet/internal/domain"

	"github.com/prometheus/client_golang/prometheus"
)

type metricsUserRepo struct {
	next UserRepo
}

func NewUserRepoWithMetrics(next UserRepo) UserRepo {
	registerDBMetricsOnce.Do(func() {
		prometheus.MustRegister(dbQueryTotal, dbQueryDuration, dbQueryErrors)
	})

	return &metricsUserRepo{next: next}
}

func (m *metricsUserRepo) Save(ctx context.Context, user *domain.User) (*domain.User, error) {
	const op = "user.save"

	start := time.Now()
	res, err := m.next.Save(ctx, user)
	observeDB(op, start, err)

	return res, err
}

func (m *metricsUserRepo) GetByLogin(ctx context.Context, login string) (*domain.User, error) {
	const op = "user.get_by_login"

	start := time.Now()
	res, err := m.next.GetByLogin(ctx, login)
	observeDB(op, start, err)

	return res, err
}

func (m *metricsUserRepo) GetByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	const op = "user.get_by_id"

	start := time.Now()
	res, err := m.next.GetByID(ctx, id)
	observeDB(op, start, err)

	return res, err
}
