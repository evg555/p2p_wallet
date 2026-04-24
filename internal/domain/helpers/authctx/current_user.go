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

	currentUser, _ := ctx.Value(CurrentUserKey).(domain.User)
	return currentUser
}
