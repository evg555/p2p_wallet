package authctx

import (
	"context"

	"p2p_wallet/internal/domain"
)

const CurrentSessionKey ctxKey = "session"

func Session(ctx context.Context) domain.Session {
	if ctx == nil {
		return domain.Session{}
	}

	if session, ok := ctx.Value(CurrentSessionKey).(domain.Session); ok {
		return session
	}

	if session, ok := ctx.Value(CurrentSessionKey).(*domain.Session); ok && session != nil {
		return *session
	}

	return domain.Session{}
}
