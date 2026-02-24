package httpserver

import (
	"context"
	"net/http"
	"time"

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

func accessLogMiddleware(log Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rec := &statusRecorder{
				ResponseWriter: w,
				status:         http.StatusOK,
			}

			start := time.Now()

			next.ServeHTTP(rec, r)

			log.Info(
				"http request",
				"req_id", requestIDFromContext(r.Context()),
				"method", r.Method,
				"path", r.URL.Path,
				"status", rec.status,
				"dur", time.Since(start),
				"bytes", rec.bytes,
			)
		})
	}
}

type ctxKey string

const requestIDKey ctxKey = "req_id"

func requestIDFromContext(ctx context.Context) string {
	v := ctx.Value(requestIDKey)
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

func requestIDMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rid := r.Header.Get("X-Request-Id")
			if rid == "" {
				rid = uuid.NewString()
			}

			ctx := context.WithValue(r.Context(), requestIDKey, rid)
			r = r.WithContext(ctx)

			w.Header().Set("X-Request-Id", rid)
			next.ServeHTTP(w, r)
		})
	}
}
