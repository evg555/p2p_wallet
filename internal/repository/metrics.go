package repository

import (
	"errors"
	"sync"
	"time"

	"p2p_wallet/internal/errs"

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
