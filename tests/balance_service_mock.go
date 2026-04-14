package tests

import (
	"context"

	"p2p_wallet/internal/domain"
	"p2p_wallet/internal/dto"
)

type balanceServiceMock struct {
	transferFn func(ctx context.Context, input dto.TransferBalanceInput) (*domain.Transaction, error)
}

func (m *balanceServiceMock) Transfer(ctx context.Context, input dto.TransferBalanceInput) (*domain.Transaction, error) {
	return m.transferFn(ctx, input)
}
