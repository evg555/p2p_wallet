package handler

import (
	"context"

	"p2p_wallet/internal/api"
	"p2p_wallet/internal/domain"
	"p2p_wallet/internal/dto"
)

type WalletService interface {
	CreateWallet(ctx context.Context, input dto.CreateWalletInput) (*domain.Wallet, error)
	ListWallets(ctx context.Context, userID int64) ([]*domain.Wallet, error)
}

func (h *handler) CreateWallet(ctx context.Context, req api.CreateWalletRequestObject) (api.CreateWalletResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

func (h *handler) ListUserWallets(ctx context.Context, req api.ListUserWalletsRequestObject) (api.ListUserWalletsResponseObject, error) {
	//TODO implement me
	panic("implement me")
}
