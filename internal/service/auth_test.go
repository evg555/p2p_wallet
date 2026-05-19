package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"p2p_wallet/internal/domain"
	"p2p_wallet/internal/domain/helpers/authctx"
	"p2p_wallet/internal/dto"
	"p2p_wallet/internal/errs"
	"p2p_wallet/internal/service/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type testLogger struct{}

func (l *testLogger) Info(_ string, _ ...any) {}
func (l *testLogger) Warn(_ string, _ ...any) {}

func TestRegister(t *testing.T) {
	ctx := context.Background()
	userRepo := mocks.NewMockUserRepo(t)
	sessionRepo := mocks.NewMockSessionRepo(t)
	logger := &testLogger{}
	svc := NewAuthService(logger, userRepo, sessionRepo)

	input := dto.RegisterInput{
		LoginInput: dto.LoginInput{
			Login:    "john",
			Password: "secret",
		},
		Name:     "John",
		LastName: "Doe",
	}

	userRepo.EXPECT().Save(ctx, mock.MatchedBy(func(u *domain.User) bool {
		return u != nil &&
			u.ID == 0 &&
			u.Login == input.Login &&
			u.CheckPassword(input.Password) &&
			u.FirstName == input.Name &&
			u.LastName == input.LastName
	})).RunAndReturn(func(_ context.Context, u *domain.User) (*domain.User, error) {
		return u, nil
	})

	got, err := svc.Register(ctx, input)
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, input.Login, got.Login)
	assert.True(t, got.CheckPassword(input.Password))
}

func TestLogin(t *testing.T) {
	t.Run("user repo error", func(t *testing.T) {
		ctx := context.Background()
		userRepo := mocks.NewMockUserRepo(t)
		sessionRepo := mocks.NewMockSessionRepo(t)
		logger := &testLogger{}
		svc := NewAuthService(logger, userRepo, sessionRepo)

		repoErr := errors.New("db down")
		userRepo.EXPECT().GetByLogin(ctx, "john").Return(nil, repoErr)

		got, err := svc.Login(ctx, dto.LoginInput{Login: "john", Password: "secret"})
		assert.Nil(t, got)
		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "failed to get user by login"))
		assert.ErrorIs(t, err, repoErr)
	})

	t.Run("user not found", func(t *testing.T) {
		ctx := context.Background()
		userRepo := mocks.NewMockUserRepo(t)
		sessionRepo := mocks.NewMockSessionRepo(t)
		logger := &testLogger{}
		svc := NewAuthService(logger, userRepo, sessionRepo)

		userRepo.EXPECT().GetByLogin(ctx, "john").Return(nil, errs.ErrUserNotFound)

		got, err := svc.Login(ctx, dto.LoginInput{Login: "john", Password: "secret"})
		assert.Nil(t, got)
		assert.ErrorIs(t, err, errs.ErrUserNotFound)
	})

	t.Run("password mismatch", func(t *testing.T) {
		ctx := context.Background()
		userRepo := mocks.NewMockUserRepo(t)
		sessionRepo := mocks.NewMockSessionRepo(t)
		logger := &testLogger{}
		svc := NewAuthService(logger, userRepo, sessionRepo)

		userRepo.EXPECT().GetByLogin(ctx, "john").Return(&domain.User{
			ID:       1,
			Login:    "john",
			Password: "another-secret",
		}, nil)

		got, err := svc.Login(ctx, dto.LoginInput{Login: "john", Password: "secret"})
		assert.Nil(t, got)
		assert.ErrorIs(t, err, errs.ErrPasswordMismatch)
	})

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		userRepo := mocks.NewMockUserRepo(t)
		sessionRepo := mocks.NewMockSessionRepo(t)
		logger := &testLogger{}
		svc := NewAuthService(logger, userRepo, sessionRepo)

		user, err := domain.NewUser("john", "secret", "John", "Doe")
		assert.NoError(t, err)

		user.ID = 1
		user.CreatedAt = time.Now()

		userRepo.EXPECT().GetByLogin(ctx, "john").Return(user, nil)
		sessionRepo.EXPECT().
			Set(
				mock.Anything,
				mock.MatchedBy(func(s *domain.Session) bool {
					return s != nil && s.UserID == user.ID
				}),
				sessionTTL,
			).
			Return()

		got, err := svc.Login(ctx, dto.LoginInput{Login: "john", Password: "secret"})
		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, domain.UserID(1), got.UserID)
		assert.Equal(t, "john", got.UserLogin)
		assert.NotEmpty(t, got.SessionID)
	})
}

func TestLogout(t *testing.T) {
	t.Run("session is empty", func(t *testing.T) {
		ctx := context.Background()
		userRepo := mocks.NewMockUserRepo(t)
		sessionRepo := mocks.NewMockSessionRepo(t)
		logger := &testLogger{}
		svc := NewAuthService(logger, userRepo, sessionRepo)

		sessionRepo.EXPECT().Delete(domain.SessionID("")).Return().Once()

		svc.Logout(ctx)
	})

	t.Run("success", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), authctx.CurrentSessionKey, domain.Session{ID: "sid"})
		userRepo := mocks.NewMockUserRepo(t)
		sessionRepo := mocks.NewMockSessionRepo(t)
		logger := &testLogger{}
		svc := NewAuthService(logger, userRepo, sessionRepo)

		sessionRepo.EXPECT().Delete(domain.SessionID("sid")).Return().Once()

		svc.Logout(ctx)
	})
}
