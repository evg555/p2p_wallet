package tests

import (
	"context"

	"p2p_wallet/internal/api"
	"p2p_wallet/internal/domain"
)

type serviceMock struct {
	registerFn func(ctx context.Context, input *api.RegisterRequest) (*domain.User, error)
	loginFn    func(ctx context.Context, input *api.LoginRequest) (*domain.AuthResult, error)
	logoutFn   func(ctx context.Context, id int64) error
}

func (m *serviceMock) Register(ctx context.Context, input *api.RegisterRequest) (*domain.User, error) {
	return m.registerFn(ctx, input)
}

func (m *serviceMock) Login(_ context.Context, input *api.LoginRequest) (*domain.AuthResult, error) {
	return m.loginFn(context.Background(), input)
}

func (m *serviceMock) Logout(ctx context.Context, id int64) error {
	return m.logoutFn(ctx, id)
}
