package repository

import (
	"context"
	"errors"
	"time"

	"p2p_wallet/internal/domain"
	"p2p_wallet/internal/infra"
	"p2p_wallet/internal/service"
)

var _ service.Repo = (*repo)(nil)

type Client interface {
	Get(id int64) (any, error)
	GetAll() ([]any, error)
	Set(id int64, v any, ttl time.Duration)
}
type repo struct {
	client Client
}

func (r repo) Save(ctx context.Context, user *domain.User) (*domain.User, error) {
	existUser, err := r.client.Get(user.ID)
	if err != nil {
		return nil, err
	}

	if existUser != nil {
		return nil, errors.New("user already exists")
	}

	r.client.Set(user.ID, user, 0)
	return user, nil
}

func (r repo) GetByLogin(ctx context.Context, login string) (*domain.User, error) {
	vals, err := r.client.GetAll()
	if err != nil {
		return nil, err
	}

	for _, v := range vals {
		user := v.(*domain.User)
		if user.Login == login {
			return user, nil
		}
	}

	return nil, nil
}

func (r repo) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	val, err := r.client.Get(id)
	if err != nil {
		return nil, err
	}

	if val == nil {
		return nil, nil
	}

	return val.(*domain.User), nil
}

func New() *repo {
	return &repo{
		client: infra.NewCache(),
	}
}
