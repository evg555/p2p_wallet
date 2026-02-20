package service

import (
	"context"
	"fmt"
	"time"

	"p2p_wallet/internal/api"
	"p2p_wallet/internal/domain"
	"p2p_wallet/internal/errs"

	"github.com/google/uuid"
)

var sessionTTL = 1 * time.Hour

type UserRepo interface {
	Save(ctx context.Context, user *domain.User) (*domain.User, error)
	GetByLogin(ctx context.Context, login string) (*domain.User, error)
	GetByID(ctx context.Context, id int64) (*domain.User, error)
}

type SessionRepo interface {
	Get(id int64) (string, error)
	Set(id int64, v string, ttl time.Duration)
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
	user := &domain.User{
		Login:     input.Login,
		Password:  domain.EncodePassword(input.Password),
		FirstName: input.Name,
		LastName:  input.LastName,
		CreatedAt: time.Now().UTC(),
	}

	user.NewUserID()

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

	if !domain.CheckPassword(input.Password, user.Password) {
		return nil, errs.ErrPasswordMismatch
	}

	newSessionID, _ := uuid.NewV7()
	s.sessionRepo.Set(user.ID, newSessionID.String(), sessionTTL)

	return &domain.AuthResult{
		UserID:       user.ID,
		UserLogin:    user.Login,
		UserName:     user.FirstName,
		UserLastName: user.LastName,
		SessionID:    newSessionID.String(),
	}, nil
}

func (s *service) Logout(ctx context.Context, id int64) error {
	existSessionID, err := s.sessionRepo.Get(id)
	if err != nil || existSessionID == "" {
		return errs.ErrSessionNotFound
	}

	sessionID := ctx.Value(domain.CtxKey(domain.SessionKey)).(string)
	if sessionID != existSessionID {
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
