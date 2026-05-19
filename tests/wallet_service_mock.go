package tests

import (
	"context"

	"p2p_wallet/internal/domain"
	"p2p_wallet/internal/dto"
)

type walletServiceMock struct {
	createWalletFn func(ctx context.Context, input dto.CreateWalletInput) (*domain.Wallet, error)
	listWalletsFn  func(ctx context.Context) ([]*domain.Wallet, error)
}

func (m *walletServiceMock) CreateWallet(ctx context.Context, input dto.CreateWalletInput) (*domain.Wallet, error) {
	return m.createWalletFn(ctx, input)
}

func (m *walletServiceMock) ListWallets(ctx context.Context) ([]*domain.Wallet, error) {
	return m.listWalletsFn(ctx)
}
