package httpserver

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

const (
	healthPath = "/health"
	readyPath  = "/ready"
	metricPath = "/metric"

	readinessTimeout = 2 * time.Second
)

func registerProbeEndpoints(r chi.Router, checker readinessChecker) {
	r.Get(healthPath, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Get(readyPath, func(w http.ResponseWriter, r *http.Request) {
		if checker == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), readinessTimeout)
		defer cancel()

		if err := checker.Ping(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusOK)
	})
}

func isProbePath(path string) bool {
	return path == healthPath || path == readyPath
}

func isObservabilityExcludedPath(path string) bool {
	return isProbePath(path) || path == metricPath
}
