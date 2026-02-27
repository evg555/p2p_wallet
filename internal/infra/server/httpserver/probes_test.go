package httpserver

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

type readinessCheckerStub struct {
	pingFn func(ctx context.Context) error
}

func (s readinessCheckerStub) Ping(ctx context.Context) error {
	return s.pingFn(ctx)
}

func TestRegisterProbeEndpoints_HealthAlwaysOK(t *testing.T) {
	t.Parallel()

	router := chi.NewRouter()
	registerProbeEndpoints(router, readinessCheckerStub{
		pingFn: func(_ context.Context) error {
			return errors.New("db down")
		},
	})

	req := httptest.NewRequest(http.MethodGet, healthPath, nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestRegisterProbeEndpoints_ReadyOKWhenPingSuccess(t *testing.T) {
	t.Parallel()

	router := chi.NewRouter()
	registerProbeEndpoints(router, readinessCheckerStub{
		pingFn: func(_ context.Context) error {
			return nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, readyPath, nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestRegisterProbeEndpoints_ReadyServiceUnavailableWhenPingFails(t *testing.T) {
	t.Parallel()

	router := chi.NewRouter()
	registerProbeEndpoints(router, readinessCheckerStub{
		pingFn: func(_ context.Context) error {
			return errors.New("db down")
		},
	})

	req := httptest.NewRequest(http.MethodGet, readyPath, nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusServiceUnavailable, rec.Code)
}
