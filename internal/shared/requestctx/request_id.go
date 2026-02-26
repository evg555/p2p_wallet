package requestctx

import "context"

type ctxKey string

const requestIDKey ctxKey = "req_id"

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

func RequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	requestID, _ := ctx.Value(requestIDKey).(string)
	return requestID
}
