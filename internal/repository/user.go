package repository

import (
	"context"

	"p2p_wallet/internal/domain"
	"p2p_wallet/internal/service"
)

var _ service.Repo = (*repo)(nil)

type repo struct{}

func (r repo) Save(ctx context.Context, user *domain.User) (*domain.User, error) {
	//TODO implement me
	panic("implement me")
}

func (r repo) GetByLogin(ctx context.Context, login string) (*domain.User, error) {
	//TODO implement me
	panic("implement me")
}

func (r repo) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	//TODO implement me
	panic("implement me")
}

func New() *repo {
	return &repo{}
}
