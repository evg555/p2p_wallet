package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"net/http"

	"p2p_wallet/internal/api"
	"p2p_wallet/internal/domain"
	"p2p_wallet/internal/handler"
)

type userServiceStub struct {
	registerFn func(ctx context.Context, input api.RegisterRequest) (*domain.User, error)
	loginFn    func(ctx context.Context, input api.LoginRequest) (*domain.User, error)
	logoutFn   func(ctx context.Context, id int64) error
	getUserFn  func(ctx context.Context, id int64) (*domain.User, error)
}

func (s *userServiceStub) Register(ctx context.Context, input api.RegisterRequest) (*domain.User, error) {
	return s.registerFn(ctx, input)
}

func (s *userServiceStub) Login(ctx context.Context, input api.LoginRequest) (*domain.User, error) {
	return s.loginFn(ctx, input)
}

func (s *userServiceStub) Logout(ctx context.Context, id int64) error {
	return s.logoutFn(ctx, id)
}

func (s *userServiceStub) GetUser(ctx context.Context, id int64) (*domain.User, error) {
	return s.getUserFn(ctx, id)
}

func newTestHandler(t *testing.T, svc *userServiceStub) http.Handler {
	t.Helper()
	h := handler.New(svc)
	return api.Handler(h)
}

func performRequest(
	t *testing.T,
	h http.Handler,
	method string,
	path string,
	body []byte,
	cookies ...*http.Cookie,
) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	for _, c := range cookies {
		req.AddCookie(c)
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestRegisterUser(t *testing.T) {
	now := time.Date(2026, 2, 20, 10, 0, 0, 0, time.UTC)
	h := newTestHandler(t, &userServiceStub{
		registerFn: func(ctx context.Context, input api.RegisterRequest) (*domain.User, error) {
			return &domain.User{
				ID:        1,
				Login:     input.Login,
				FirstName: input.Name,
				LastName:  input.LastName,
				CreatedAt: now,
			}, nil
		},
		loginFn:   func(ctx context.Context, input api.LoginRequest) (*domain.User, error) { return nil, nil },
		logoutFn:  func(ctx context.Context, id int64) error { return nil },
		getUserFn: func(ctx context.Context, id int64) (*domain.User, error) { return nil, nil },
	})
	rec := performRequest(
		t,
		h,
		http.MethodPost,
		"/users/register",
		[]byte(`{"login":"u1","password":"p1","name":"John","last_name":"Doe"}`),
	)

	if rec.Code != http.StatusCreated {
		t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusCreated)
	}

	var got api.UserResponse
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Id != 1 || got.Login != "u1" || got.Name != "John" || got.LastName != "Doe" {
		t.Fatalf("unexpected response body: %+v", got)
	}
}

func TestLoginUserSetsCookie(t *testing.T) {
	now := time.Date(2026, 2, 20, 10, 0, 0, 0, time.UTC)
	h := newTestHandler(t, &userServiceStub{
		registerFn: func(ctx context.Context, input api.RegisterRequest) (*domain.User, error) { return nil, nil },
		loginFn: func(ctx context.Context, input api.LoginRequest) (*domain.User, error) {
			return &domain.User{
				ID:        2,
				Login:     input.Login,
				FirstName: "Jane",
				LastName:  "Roe",
				CreatedAt: now,
			}, nil
		},
		logoutFn:  func(ctx context.Context, id int64) error { return nil },
		getUserFn: func(ctx context.Context, id int64) (*domain.User, error) { return nil, nil },
	})
	rec := performRequest(
		t,
		h,
		http.MethodPost,
		"/users/login",
		[]byte(`{"login":"u2","password":"p2"}`),
	)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusOK)
	}

	resp := rec.Result()
	defer resp.Body.Close()
	cookies := resp.Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected session cookie to be set")
	}
	if cookies[0].Name != "session_id" || cookies[0].Value == "" {
		t.Fatalf("unexpected cookie: %+v", cookies[0])
	}
}

func TestGetUserByIDRequiresCookie(t *testing.T) {
	h := newTestHandler(t, &userServiceStub{
		registerFn: func(ctx context.Context, input api.RegisterRequest) (*domain.User, error) { return nil, nil },
		loginFn:    func(ctx context.Context, input api.LoginRequest) (*domain.User, error) { return nil, nil },
		logoutFn:   func(ctx context.Context, id int64) error { return nil },
		getUserFn:  func(ctx context.Context, id int64) (*domain.User, error) { return nil, nil },
	})
	rec := performRequest(t, h, http.MethodGet, "/users/1", nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestGetUserByID(t *testing.T) {
	now := time.Date(2026, 2, 20, 10, 0, 0, 0, time.UTC)
	h := newTestHandler(t, &userServiceStub{
		registerFn: func(ctx context.Context, input api.RegisterRequest) (*domain.User, error) { return nil, nil },
		loginFn:    func(ctx context.Context, input api.LoginRequest) (*domain.User, error) { return nil, nil },
		logoutFn:   func(ctx context.Context, id int64) error { return nil },
		getUserFn: func(ctx context.Context, id int64) (*domain.User, error) {
			return &domain.User{
				ID:        id,
				Login:     "u3",
				FirstName: "Alice",
				LastName:  "Smith",
				CreatedAt: now,
			}, nil
		},
	})
	rec := performRequest(
		t,
		h,
		http.MethodGet,
		"/users/3",
		nil,
		&http.Cookie{Name: "session_id", Value: "session-1"},
	)
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusOK)
	}

	var got api.UserResponse
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Id != 3 || got.Login != "u3" {
		t.Fatalf("unexpected response body: %+v", got)
	}
}

func TestLogoutUser(t *testing.T) {
	called := false
	h := newTestHandler(t, &userServiceStub{
		registerFn: func(ctx context.Context, input api.RegisterRequest) (*domain.User, error) { return nil, nil },
		loginFn:    func(ctx context.Context, input api.LoginRequest) (*domain.User, error) { return nil, nil },
		logoutFn: func(ctx context.Context, id int64) error {
			called = true
			if id != 7 {
				t.Fatalf("unexpected id: got %d want %d", id, 7)
			}
			return nil
		},
		getUserFn: func(ctx context.Context, id int64) (*domain.User, error) { return nil, nil },
	})
	rec := performRequest(
		t,
		h,
		http.MethodPost,
		"/users/7/logout",
		nil,
		&http.Cookie{Name: "session_id", Value: "session-7"},
	)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusNoContent)
	}
	if !called {
		t.Fatal("expected logout service to be called")
	}
}
