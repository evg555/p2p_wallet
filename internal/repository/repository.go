package repository

import (
	"context"

	"p2p_wallet/internal/domain"
)

type UserRepo interface {
	Save(ctx context.Context, user *domain.User) (*domain.User, error)
	GetByLogin(ctx context.Context, login string) (*domain.User, error)
	GetByID(ctx context.Context, id int64) (*domain.User, error)
}

type WalletRepo interface {
	Save(ctx context.Context, wallet *domain.Wallet) (*domain.Wallet, error)
	FindByUserID(ctx context.Context, userID domain.UserID) ([]*domain.Wallet, error)
}
