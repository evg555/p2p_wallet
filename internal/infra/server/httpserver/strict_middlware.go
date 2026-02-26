package httpserver

import (
	"context"
	"net/http"

	"p2p_wallet/internal/api"
	"p2p_wallet/internal/shared/requestctx"
)

func ErrorLoggingMiddleware(log Logger) api.StrictMiddlewareFunc {
	return func(f api.StrictHandlerFunc, operationID string) api.StrictHandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, args interface{}) (interface{}, error) {
			resp, err := f(ctx, w, r, args)
			if err != nil {
				requestID := requestctx.RequestID(ctx)
				log.Error("handler error",
					"operation", operationID,
					"req_id", requestID,
					"error", err.Error(),
				)
			}
			return resp, err
		}
	}
}
