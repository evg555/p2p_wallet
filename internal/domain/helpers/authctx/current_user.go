package authctx

import (
	"context"

	"p2p_wallet/internal/domain"
)

type ctxKey string

const CurrentUserKey ctxKey = "current_user"

func CurrentUser(ctx context.Context) domain.User {
	if ctx == nil {
		return domain.User{}
	}

	if currentUser, ok := ctx.Value(CurrentUserKey).(domain.User); ok {
		return currentUser
	}

	if currentUser, ok := ctx.Value(CurrentUserKey).(*domain.User); ok && currentUser != nil {
		return *currentUser
	}

	return domain.User{}
}
