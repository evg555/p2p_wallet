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

	session, _ := ctx.Value(CurrentSessionKey).(domain.Session)
	return session
}
