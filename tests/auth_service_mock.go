package tests

import (
	"context"

	"p2p_wallet/internal/domain"
	"p2p_wallet/internal/dto"
)

type authServiceMock struct {
	registerFn func(ctx context.Context, input dto.RegisterInput) (*domain.User, error)
	loginFn    func(ctx context.Context, input dto.LoginInput) (*domain.AuthResult, error)
	logoutFn   func(ctx context.Context, id int64) error
}

func (m *authServiceMock) Register(ctx context.Context, input dto.RegisterInput) (*domain.User, error) {
	return m.registerFn(ctx, input)
}

func (m *authServiceMock) Login(ctx context.Context, input dto.LoginInput) (*domain.AuthResult, error) {
	return m.loginFn(ctx, input)
}

func (m *authServiceMock) Logout(ctx context.Context, id int64) error {
	return m.logoutFn(ctx, id)
}
