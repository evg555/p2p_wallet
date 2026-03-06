package tests

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"p2p_wallet/internal/api"
	"p2p_wallet/internal/domain"
	"p2p_wallet/internal/dto"
	"p2p_wallet/internal/errs"
	"p2p_wallet/internal/handler"

	"github.com/stretchr/testify/require"
)

func TestRegisterUser_Contract_Created(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, time.February, 25, 10, 0, 0, 0, time.UTC)
	var gotInput dto.RegisterInput

	svc := &authServiceMock{
		registerFn: func(_ context.Context, input dto.RegisterInput) (*domain.User, error) {
			gotInput = input
			return &domain.User{
				ID:        101,
				Login:     "john",
				FirstName: "John",
				LastName:  "Doe",
				CreatedAt: createdAt,
			}, nil
		},
	}
	server := newTestServer(svc)

	req := httptest.NewRequest(http.MethodPost, "/users/register", strings.NewReader(`{
		"login":"john",
		"password":"secret",
		"name":"John",
		"last_name":"Doe"
	}`))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
	require.NotNil(t, gotInput)
	require.Equal(t, "john", gotInput.Login)
	require.Equal(t, "secret", gotInput.Password)
	require.Equal(t, "John", gotInput.Name)
	require.Equal(t, "Doe", gotInput.LastName)

	var resp api.UserResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	require.Equal(t, int64(101), resp.Id)
	require.Equal(t, "john", resp.Login)
	require.Equal(t, "John", resp.Name)
	require.Equal(t, "Doe", resp.LastName)
	require.Equal(t, createdAt, resp.CreatedAt)
}

func TestRegisterUser_Contract_Conflict(t *testing.T) {
	t.Parallel()

	svc := &authServiceMock{
		registerFn: func(_ context.Context, _ dto.RegisterInput) (*domain.User, error) {
			return nil, errs.ErrUserAlreadyExist
		},
	}
	server := newTestServer(svc)

	req := httptest.NewRequest(http.MethodPost, "/users/register", strings.NewReader(`{
		"login":"john",
		"password":"secret",
		"name":"John",
		"last_name":"Doe"
	}`))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	require.Equal(t, http.StatusConflict, rec.Code)
	require.Contains(t, rec.Body.String(), "user already exist")
}

func TestRegisterUser_Contract_Bad_Request(t *testing.T) {
	t.Parallel()

	svc := &authServiceMock{}
	server := newTestServer(svc)

	req := httptest.NewRequest(http.MethodPost, "/users/register", nil)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRegisterUser_Contract_InternalServerError(t *testing.T) {
	t.Parallel()

	svc := &authServiceMock{
		registerFn: func(_ context.Context, _ dto.RegisterInput) (*domain.User, error) {
			return nil, errors.New("database unavailable")
		},
	}
	server := newTestServer(svc)

	req := httptest.NewRequest(http.MethodPost, "/users/register", strings.NewReader(`{
		"login":"john",
		"password":"secret",
		"name":"John",
		"last_name":"Doe"
	}`))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	resp, err := io.ReadAll(rec.Body)
	require.NoError(t, err)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Equal(t, "internal server error\n", string(resp))
}

func TestLoginUser_Contract_Ok(t *testing.T) {
	t.Parallel()

	var gotInput dto.LoginInput
	var gotInputSet bool

	svc := &authServiceMock{
		loginFn: func(_ context.Context, input dto.LoginInput) (*domain.AuthResult, error) {
			gotInput = input
			gotInputSet = true
			return &domain.AuthResult{
				UserID:       1,
				UserLogin:    "john",
				UserName:     "John",
				UserLastName: "Doe",
				SessionID:    "abc123",
			}, nil
		},
	}
	server := newTestServer(svc)

	req := httptest.NewRequest(http.MethodPost, "/users/login", strings.NewReader(`{
		"login":"john",
		"password":"secret"
	}`))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.True(t, gotInputSet)
	require.Equal(t, "john", gotInput.Login)
	require.Equal(t, "secret", gotInput.Password)

	var resp api.LoginResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	require.Equal(t, int64(1), resp.Id)
	require.Equal(t, "john", resp.Login)
	require.Equal(t, "John", resp.Name)
	require.Equal(t, "Doe", resp.LastName)
}

func TestLoginUser_Contract_Unauthorized(t *testing.T) {
	t.Parallel()

	svc := &authServiceMock{
		loginFn: func(_ context.Context, _ dto.LoginInput) (*domain.AuthResult, error) {
			return nil, errs.ErrPasswordMismatch
		},
	}
	server := newTestServer(svc)

	req := httptest.NewRequest(http.MethodPost, "/users/login", strings.NewReader(`{
		"login":"john",
		"password":"secret"
	}`))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
	require.Contains(t, rec.Body.String(), "password mismatch")
}

func TestLoginUser_Contract_Not_Found(t *testing.T) {
	t.Parallel()

	svc := &authServiceMock{
		loginFn: func(_ context.Context, _ dto.LoginInput) (*domain.AuthResult, error) {
			return nil, errs.ErrUserNotFound
		},
	}
	server := newTestServer(svc)

	req := httptest.NewRequest(http.MethodPost, "/users/login", strings.NewReader(`{
		"login":"john",
		"password":"secret"
	}`))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
	require.Contains(t, rec.Body.String(), "user not found")
}

func TestLoginUser_Contract_Bad_Request(t *testing.T) {
	t.Parallel()

	svc := &authServiceMock{}
	server := newTestServer(svc)

	req := httptest.NewRequest(http.MethodPost, "/users/login", nil)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestLoginUser_Contract_InternalServerError(t *testing.T) {
	t.Parallel()

	svc := &authServiceMock{
		loginFn: func(_ context.Context, _ dto.LoginInput) (*domain.AuthResult, error) {
			return nil, errors.New("database unavailable")
		},
	}
	server := newTestServer(svc)

	req := httptest.NewRequest(http.MethodPost, "/users/login", strings.NewReader(`{
		"login":"john",
		"password":"secret"
	}`))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	resp, err := io.ReadAll(rec.Body)
	require.NoError(t, err)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Equal(t, "internal server error\n", string(resp))
}

func TestLogoutUser_Contract_No_Content(t *testing.T) {
	t.Parallel()

	var gotID *int64

	svc := &authServiceMock{
		logoutFn: func(_ context.Context, id int64) error {
			gotID = &id
			return nil
		},
	}
	server := newTestServer(svc)

	req := httptest.NewRequest(http.MethodPost, "/users/1/logout", nil)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
	require.Equal(t, int64(1), *gotID)
}

func TestLogoutUser_Contract_Unauthorized(t *testing.T) {
	t.Parallel()

	svc := &authServiceMock{
		logoutFn: func(_ context.Context, id int64) error {
			return errs.ErrSessionNotFound
		},
	}
	server := newTestServer(svc)

	req := httptest.NewRequest(http.MethodPost, "/users/1/logout", nil)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
	require.Contains(t, rec.Body.String(), "session not found")
}

func TestLogoutUser_Contract_Not_Found(t *testing.T) {
	t.Parallel()

	svc := &authServiceMock{
		logoutFn: func(_ context.Context, id int64) error {
			return errs.ErrUserNotFound
		},
	}
	server := newTestServer(svc)

	req := httptest.NewRequest(http.MethodPost, "/users/1/logout", nil)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
	require.Contains(t, rec.Body.String(), "user not found")
}

func TestLogoutUser_Contract_Forbidden(t *testing.T) {
	t.Parallel()

	svc := &authServiceMock{
		logoutFn: func(_ context.Context, id int64) error {
			return errs.ErrAccessDenied
		},
	}
	server := newTestServer(svc)

	req := httptest.NewRequest(http.MethodPost, "/users/1/logout", nil)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	require.Equal(t, http.StatusForbidden, rec.Code)
	require.Contains(t, rec.Body.String(), "access denied")
}

func TestLogoutUser_Contract_InternalServerError(t *testing.T) {
	t.Parallel()

	svc := &authServiceMock{
		logoutFn: func(_ context.Context, id int64) error {
			return errors.New("database unavailable")
		},
	}
	server := newTestServer(svc)

	req := httptest.NewRequest(http.MethodPost, "/users/1/logout", nil)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	resp, err := io.ReadAll(rec.Body)
	require.NoError(t, err)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Equal(t, "internal server error\n", string(resp))
}

func newTestServer(svc handler.AuthService) http.Handler {
	h := handler.New(svc, nil)
	return api.Handler(api.NewStrictHandlerWithOptions(h, nil, api.StrictHTTPServerOptions{
		RequestErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		},
		ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, "internal server error", http.StatusInternalServerError)
		},
	}))
}
