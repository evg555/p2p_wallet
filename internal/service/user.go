package service

import (
	"context"
	"fmt"

	"p2p_wallet/internal/api"
	"p2p_wallet/internal/domain"
	"p2p_wallet/internal/errs"
)

type Repo interface {
	Save(ctx context.Context, user *domain.User) (*domain.User, error)
	GetByLogin(ctx context.Context, login string) (*domain.User, error)
	GetByID(ctx context.Context, id int64) (*domain.User, error)
}

type service struct {
	repo Repo
}

func New(repo Repo) *service {
	return &service{repo: repo}
}

func (s *service) Register(ctx context.Context, input api.RegisterRequest) (*domain.User, error) {
	user := &domain.User{
		Login:     input.Login,
		Password:  domain.EncodePassword(input.Password),
		FirstName: input.Name,
		LastName:  input.LastName,
	}

	user.NewUserID()

	return s.repo.Save(ctx, user)
}

func (s *service) Login(ctx context.Context, input api.LoginRequest) (*domain.User, error) {
	user, err := s.repo.GetByLogin(ctx, input.Login)
	if err != nil {
		return nil, fmt.Errorf("repo: failed to get user by login: %w", err)
	}

	if user == nil {
		return nil, errs.ErrUserNotFound
	}

	if domain.CheckPassword(user.Password, input.Password) {
		return nil, errs.ErrPasswordMismatch
	}

	return user, nil
}

func (s *service) Logout(ctx context.Context, id int64) error {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("repo: failed to get user by id: %w", err)
	}

	if user == nil {
		return errs.ErrUserNotFound
	}

	return nil
}

func (s *service) GetUser(ctx context.Context, id int64) (*domain.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("repo: failed to get user by id: %w", err)
	}

	if user == nil {
		return nil, errs.ErrUserNotFound
	}

	return user, nil
}
