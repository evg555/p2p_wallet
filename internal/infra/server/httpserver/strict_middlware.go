package httpserver

import (
	"context"
	"net/http"

	"p2p_wallet/internal/api"
	"p2p_wallet/internal/shared/requestctx"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func TracingMiddleware() api.StrictMiddlewareFunc {
	tracer := otel.Tracer("p2p-wallet/httpserver/strict")

	return func(f api.StrictHandlerFunc, operationID string) api.StrictHandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, args interface{}) (interface{}, error) {
			ctx, span := tracer.Start(ctx, "handler."+operationID, trace.WithSpanKind(trace.SpanKindInternal))
			span.SetAttributes(attribute.String("handler.operation", operationID))
			if requestID := requestctx.RequestID(ctx); requestID != "" {
				span.SetAttributes(attribute.String("request.id", requestID))
			}

			resp, err := f(ctx, w, r, args)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}
			span.End()

			return resp, err
		}
	}
}

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
