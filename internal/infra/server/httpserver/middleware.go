package httpserver

import (
	"net/http"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func loggerMiddleware(log Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rec := &statusRecorder{
				ResponseWriter: w,
				status:         http.StatusOK,
			}

			next.ServeHTTP(rec, r)

			switch true {
			case rec.status < 400:
				log.Info(
					"http request",
					"method", r.Method,
					"path", r.URL.Path,
					"status", rec.status,
				)
			case 400 <= rec.status && rec.status < 500:
				log.Warn(
					"http request",
					"method", r.Method,
					"path", r.URL.Path,
					"status", rec.status,
				)
			default:
				log.Error(
					"http request",
					"method", r.Method,
					"path", r.URL.Path,
					"status", rec.status,
				)
			}
		})
	}
}
