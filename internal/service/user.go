package service

import (
	"context"

	"p2p_wallet/internal/api"
	"p2p_wallet/internal/domain"
)

type Repo interface {
	Register()
	Login()
	Logout()
	GetUser()
}

type service struct {
	repo Repo
}

func New(repo Repo) *service {
	return &service{repo: repo}
}

func (s *service) Register(ctx context.Context, input api.RegisterRequest) (domain.User, error) {
	panic("implement me")
}

func (s *service) Login(ctx context.Context, input api.LoginRequest) (domain.User, error) {
	panic("implement me")
}
func (s *service) Logout(ctx context.Context, id int64) error {
	panic("implement me")
}

func (s *service) GetUser(ctx context.Context, id int64) (domain.User, error) {
	panic("implement me")
}
