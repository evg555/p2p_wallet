package httpserver

import (
	"context"
	"net/http"
	"time"

	"p2p_wallet/internal/domain"
	"p2p_wallet/internal/shared/requestctx"

	"github.com/google/uuid"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	n, err := r.ResponseWriter.Write(b)
	r.bytes += n
	return n, err
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		ctx := requestctx.WithRequestID(r.Context(), requestID)
		r = r.WithContext(ctx)
		w.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(w, r)
	})
}

func AccessLogMiddleware(log Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rec := &statusRecorder{
				ResponseWriter: w,
				status:         http.StatusOK,
			}

			start := time.Now()

			next.ServeHTTP(rec, r)

			requestID := requestctx.RequestID(r.Context())

			log.Info(
				"http request",
				"req_id", requestID,
				"method", r.Method,
				"path", r.URL.Path,
				"status", rec.status,
				"dur", time.Since(start).String(),
				"bytes", rec.bytes,
			)
		})
	}
}

func SessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(domain.SessionKey)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), domain.CtxKey(domain.SessionKey), cookie.Value)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
