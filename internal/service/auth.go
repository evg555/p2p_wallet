package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"p2p_wallet/internal/domain"
	"p2p_wallet/internal/domain/helpers/authctx"
	"p2p_wallet/internal/dto"
	"p2p_wallet/internal/errs"
	"p2p_wallet/internal/shared/requestctx"
)

var sessionTTL = 1 * time.Hour

type UserRepo interface {
	Save(ctx context.Context, user *domain.User) (*domain.User, error)
	GetByLogin(ctx context.Context, login string) (*domain.User, error)
	GetByID(ctx context.Context, id domain.UserID) (*domain.User, error)
}

type SessionRepo interface {
	Get(key domain.SessionID) (*domain.Session, error)
	Set(key domain.SessionID, v *domain.Session, ttl time.Duration)
	Delete(key domain.SessionID)
}

type Logger interface {
	Info(msg string, keysAndValues ...any)
	Warn(msg string, keysAndValues ...any)
}

type authService struct {
	log         Logger
	userRepo    UserRepo
	sessionRepo SessionRepo
}

func NewAuthService(log Logger, repo UserRepo, sessionRepo SessionRepo) *authService {
	return &authService{
		log:         log,
		userRepo:    repo,
		sessionRepo: sessionRepo,
	}
}

func (s *authService) Register(ctx context.Context, input dto.RegisterInput) (*domain.User, error) {
	user, err := domain.NewUser(input.Login, input.Password, input.Name, input.LastName)
	if err != nil {
		s.log.Warn("user register failed", withReqID(ctx, "login", input.Login, "error", err.Error())...)
		return nil, err
	}

	savedUser, err := s.userRepo.Save(ctx, user)
	if err != nil {
		if errors.Is(err, errs.ErrUserAlreadyExist) {
			s.log.Warn("user register failed", withReqID(ctx, "login", input.Login, "error", err.Error())...)
		}
		return nil, err
	}

	s.log.Info("user register succeeded", withReqID(ctx, "user_id", savedUser.ID, "login", savedUser.Login)...)
	return savedUser, nil
}

func (s *authService) Login(ctx context.Context, input dto.LoginInput) (*domain.AuthResult, error) {
	user, err := s.userRepo.GetByLogin(ctx, input.Login)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			s.log.Warn("user login failed", withReqID(ctx, "login", input.Login, "error", err.Error())...)
		}

		return nil, fmt.Errorf("userRepo: failed to get user by login: %w", err)
	}

	if !user.CheckPassword(input.Password) {
		s.log.Warn("user login failed: password mismatch", withReqID(ctx, "login", input.Login, "user_id", user.ID)...)
		return nil, errs.ErrPasswordMismatch
	}

	newSession := domain.NewSession(user.ID, sessionTTL)
	s.sessionRepo.Set(newSession.SessionID(), newSession, sessionTTL)
	s.log.Info("user login succeeded", withReqID(ctx, "user_id", user.ID, "login", user.Login)...)

	return &domain.AuthResult{
		UserID:       user.ID,
		UserLogin:    user.Login,
		UserName:     user.FirstName,
		UserLastName: user.LastName,
		SessionID:    newSession.SessionID(),
	}, nil
}

func (s *authService) Logout(ctx context.Context) {
	session := authctx.Session(ctx)
	user := authctx.CurrentUser(ctx)

	s.sessionRepo.Delete(session.SessionID())
	s.log.Info("user logout succeeded", withReqID(ctx, "user_id", user.ID, "login", user.Login)...)
}

func withReqID(ctx context.Context, keysAndValues ...any) []any {
	requestID := requestctx.RequestID(ctx)
	if requestID == "" {
		return keysAndValues
	}

	out := make([]any, 0, len(keysAndValues)+2)
	out = append(out, keysAndValues...)
	out = append(out, "req_id", requestID)

	return out
}
