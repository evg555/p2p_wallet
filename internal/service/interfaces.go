package service

import (
	"context"

	"p2p_wallet/internal/domain"
)

type WalletRepo interface {
	Save(ctx context.Context, wallet *domain.Wallet) (*domain.Wallet, error)
	FindByUserID(ctx context.Context, userID domain.UserID) ([]*domain.Wallet, error)
	FindByID(ctx context.Context, id domain.WalletID) (*domain.Wallet, error)
}

type BalanceRepo interface {
	CreateTransaction(ctx context.Context, transaction *domain.Transaction) (*domain.Transaction, error)
}
