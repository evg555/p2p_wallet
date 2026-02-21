package service

import (
	"context"
	"fmt"
	"time"

	"p2p_wallet/internal/api"
	"p2p_wallet/internal/domain"
	"p2p_wallet/internal/errs"
)

var sessionTTL = 1 * time.Hour

type UserRepo interface {
	Save(ctx context.Context, user *domain.User) (*domain.User, error)
	GetByLogin(ctx context.Context, login string) (*domain.User, error)
	GetByID(ctx context.Context, id int64) (*domain.User, error)
}

type SessionRepo interface {
	Get(id int64) (domain.SessionID, error)
	Set(id int64, v domain.SessionID, ttl time.Duration)
	Delete(id int64)
}

type service struct {
	userRepo    UserRepo
	sessionRepo SessionRepo
}

func New(repo UserRepo, sessionRepo SessionRepo) *service {
	return &service{
		userRepo:    repo,
		sessionRepo: sessionRepo,
	}
}

func (s *service) Register(ctx context.Context, input api.RegisterRequest) (*domain.User, error) {
	user, err := domain.NewUser(input.Login, input.Password, input.Name, input.LastName)
	if err != nil {
		return nil, err
	}

	return s.userRepo.Save(ctx, user)
}

func (s *service) Login(ctx context.Context, input api.LoginRequest) (*domain.AuthResult, error) {
	user, err := s.userRepo.GetByLogin(ctx, input.Login)
	if err != nil {
		return nil, fmt.Errorf("userRepo: failed to get user by login: %w", err)
	}

	if user == nil {
		return nil, errs.ErrUserNotFound
	}

	if !user.CheckPassword(input.Password) {
		return nil, errs.ErrPasswordMismatch
	}

	newSessionID := domain.NewSessionID()
	s.sessionRepo.Set(user.ID, newSessionID, sessionTTL)

	return &domain.AuthResult{
		UserID:       user.ID,
		UserLogin:    user.Login,
		UserName:     user.FirstName,
		UserLastName: user.LastName,
		SessionID:    newSessionID,
	}, nil
}

func (s *service) Logout(ctx context.Context, id int64) error {
	existSessionID, err := s.sessionRepo.Get(id)
	if err != nil || existSessionID.IsEmpty() {
		return errs.ErrSessionNotFound
	}

	sessionID, ok := ctx.Value(domain.CtxKey(domain.SessionKey)).(string)
	if !ok || !existSessionID.Equal(sessionID) {
		return errs.ErrAccessDenied
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("userRepo: failed to get user by id: %w", err)
	}

	if user == nil {
		return errs.ErrUserNotFound
	}

	s.sessionRepo.Delete(user.ID)
	return nil
}
