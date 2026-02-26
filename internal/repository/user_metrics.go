package repository

import (
	"context"
	"errors"
	"sync"
	"time"

	"p2p_wallet/internal/domain"
	"p2p_wallet/internal/errs"
	"p2p_wallet/internal/service"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	registerDBMetricsOnce sync.Once

	dbQueryTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_query_total",
			Help: "Total number of DB operations by status.",
		},
		[]string{"op", "status"},
	)

	dbQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "DB operation latency in seconds.",
			Buckets: []float64{0.001, 0.003, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
		},
		[]string{"op"},
	)

	dbQueryErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_query_errors_total",
			Help: "Total number of DB operation errors by class.",
		},
		[]string{"op", "error_class"},
	)
)

type metricsUserRepo struct {
	next service.UserRepo
}

func NewUserRepoWithMetrics(next service.UserRepo) service.UserRepo {
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

func (m *metricsUserRepo) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	const op = "user.get_by_id"

	start := time.Now()
	res, err := m.next.GetByID(ctx, id)
	observeDB(op, start, err)

	return res, err
}

func observeDB(op string, started time.Time, err error) {
	status := "ok"
	if err != nil {
		status = "error"
		dbQueryErrors.WithLabelValues(op, errorClass(err)).Inc()
	}

	dbQueryTotal.WithLabelValues(op, status).Inc()
	dbQueryDuration.WithLabelValues(op).Observe(time.Since(started).Seconds())
}

func errorClass(err error) string {
	switch {
	case errors.Is(err, errs.ErrUserNotFound):
		return "user_not_found"
	case errors.Is(err, errs.ErrUserAlreadyExist):
		return "user_already_exist"
	default:
		return "unknown"
	}
}
