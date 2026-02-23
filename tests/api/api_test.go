package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"p2p_wallet/internal/api"
	"p2p_wallet/internal/domain"
	"p2p_wallet/internal/handler"
	"p2p_wallet/internal/repository"
	"p2p_wallet/internal/service"
	"p2p_wallet/internal/shared/config"
	"p2p_wallet/internal/shared/logger"

	"github.com/stretchr/testify/require"
)

func newTestHandler(t *testing.T) http.Handler {
	t.Helper()

	log, err := logger.New(config.LoggerConfig{Level: "info"})
	require.NoError(t, err)

	userRepo := repository.NewUserCacheRepo()
	sessionRepo := repository.NewSessionRepo()
	srv := service.New(log, userRepo, sessionRepo)
	h := handler.New(srv)
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

func TestUserFlowE2E(t *testing.T) {
	h := newTestHandler(t)

	registerRec := performRequest(
		t,
		h,
		http.MethodPost,
		"/users/register",
		[]byte(`{"login":"u1","password":"p1","name":"John","last_name":"Doe"}`),
	)
	if registerRec.Code != http.StatusCreated {
		t.Fatalf("unexpected register status: got %d want %d", registerRec.Code, http.StatusCreated)
	}

	var registered api.UserResponse
	if err := json.NewDecoder(registerRec.Body).Decode(&registered); err != nil {
		t.Fatalf("decode register response: %v", err)
	}
	if registered.Id <= 0 {
		t.Fatalf("unexpected user id: %d", registered.Id)
	}
	if registered.CreatedAt.IsZero() {
		t.Fatal("expected non-zero created_at after register")
	}
	if registered.Login != "u1" || registered.Name != "John" || registered.LastName != "Doe" {
		t.Fatalf("unexpected register response body: %+v", registered)
	}

	loginRec := performRequest(
		t,
		h,
		http.MethodPost,
		"/users/login",
		[]byte(`{"login":"u1","password":"p1"}`),
	)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("unexpected login status: got %d want %d", loginRec.Code, http.StatusOK)
	}

	loginResp := loginRec.Result()
	defer loginResp.Body.Close()
	cookies := loginResp.Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected session cookie after login, Set-Cookie=%q", loginResp.Header.Values("Set-Cookie"))
	}

	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == domain.SessionKey && c.Value != "" {
			sessionCookie = c
			break
		}
	}
	if sessionCookie == nil {
		t.Fatalf("session cookie is missing or empty: %+v", cookies)
	}

	logoutRec := performRequest(
		t,
		h,
		http.MethodPost,
		"/users/"+strconv.FormatInt(registered.Id, 10)+"/logout",
		nil,
		sessionCookie,
	)
	if logoutRec.Code != http.StatusNoContent {
		t.Fatalf("unexpected logout status: got %d want %d", logoutRec.Code, http.StatusNoContent)
	}
}
