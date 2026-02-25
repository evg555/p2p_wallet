package httpserver

import (
	"context"
	"net/http"

	"p2p_wallet/internal/api"

	"github.com/google/uuid"
)

type ctxKey string

const requestIDKey ctxKey = "req_id"

func RequestIDMiddleware(f api.StrictHandlerFunc, operationID string) api.StrictHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, args interface{}) (interface{}, error) {
		requestID := uuid.New().String()
		ctx = context.WithValue(ctx, requestIDKey, requestID)
		r = r.WithContext(context.WithValue(r.Context(), requestIDKey, requestID))
		w.Header().Set("X-Request-ID", requestID)
		return f(ctx, w, r, args)
	}
}

func ErrorLoggingMiddleware(log Logger) api.StrictMiddlewareFunc {
	return func(f api.StrictHandlerFunc, operationID string) api.StrictHandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, args interface{}) (interface{}, error) {
			resp, err := f(ctx, w, r, args)
			if err != nil {
				requestID, _ := ctx.Value(requestIDKey).(string)
				log.Error("handler error",
					"operation", operationID,
					"request_id", requestID,
					"error", err.Error(),
				)
			}
			return resp, err
		}
	}
}
