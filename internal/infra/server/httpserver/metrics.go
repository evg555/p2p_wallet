package httpserver

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type httpMetrics struct {
	registry      *prometheus.Registry
	requestsTotal *prometheus.CounterVec
	errorsTotal   *prometheus.CounterVec
	latency       *prometheus.HistogramVec
}

func newHTTPMetrics() *httpMetrics {
	registry := prometheus.NewRegistry()

	requestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status_code"},
	)

	errorsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_errors_total",
			Help: "Total number of HTTP requests with 4xx/5xx statuses.",
		},
		[]string{"method", "path", "status_code"},
	)

	latency := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds.",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path"},
	)

	registry.MustRegister(requestsTotal, errorsTotal, latency)

	return &httpMetrics{
		registry:      registry,
		requestsTotal: requestsTotal,
		errorsTotal:   errorsTotal,
		latency:       latency,
	}
}

func (m *httpMetrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

func (m *httpMetrics) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metric" {
			next.ServeHTTP(w, r)
			return
		}

		rec := &statusRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		start := time.Now()
		next.ServeHTTP(rec, r)

		path := routePattern(r)
		code := statusCode(rec.status)
		durationSec := time.Since(start).Seconds()

		m.requestsTotal.WithLabelValues(r.Method, path, code).Inc()
		m.latency.WithLabelValues(r.Method, path).Observe(durationSec)

		if code == "4xx" || code == "5xx" {
			m.errorsTotal.WithLabelValues(r.Method, path, code).Inc()
		}
	})
}

func routePattern(r *http.Request) string {
	if rctx := chi.RouteContext(r.Context()); rctx != nil {
		p := rctx.RoutePattern()
		if p != "" {
			return p
		}
	}
	return r.URL.Path
}

func statusCode(status int) string {
	class := status / 100
	if class < 1 || class > 5 {
		return "other"
	}
	return strconv.Itoa(class) + "xx"
}
