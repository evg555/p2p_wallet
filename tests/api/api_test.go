package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"p2p_wallet/internal/api"
	"p2p_wallet/internal/handler"
	"p2p_wallet/internal/repository"
	"p2p_wallet/internal/service"
)

func newTestHandler(t *testing.T) http.Handler {
	t.Helper()
	repo := repository.New()
	srv := service.New(repo)
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
		t.Fatal("expected session cookie after login")
	}

	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "session_id" && c.Value != "" {
			sessionCookie = c
			break
		}
	}
	if sessionCookie == nil {
		t.Fatalf("session cookie is missing or empty: %+v", cookies)
	}

	getRec := performRequest(
		t,
		h,
		http.MethodGet,
		"/users/"+stringifyInt64(registered.Id),
		nil,
		sessionCookie,
	)
	if getRec.Code != http.StatusOK {
		t.Fatalf("unexpected get user status: got %d want %d", getRec.Code, http.StatusOK)
	}

	var gotUser api.UserResponse
	if err := json.NewDecoder(getRec.Body).Decode(&gotUser); err != nil {
		t.Fatalf("decode get user response: %v", err)
	}
	if gotUser.Id != registered.Id || gotUser.Login != "u1" {
		t.Fatalf("unexpected get user response body: %+v", gotUser)
	}
	if gotUser.CreatedAt.IsZero() {
		t.Fatal("expected non-zero created_at in get user response")
	}

	logoutRec := performRequest(
		t,
		h,
		http.MethodPost,
		"/users/"+stringifyInt64(registered.Id)+"/logout",
		nil,
		sessionCookie,
	)
	if logoutRec.Code != http.StatusNoContent {
		t.Fatalf("unexpected logout status: got %d want %d", logoutRec.Code, http.StatusNoContent)
	}
}

func TestGetUserByIDRequiresCookieE2E(t *testing.T) {
	h := newTestHandler(t)
	rec := performRequest(t, h, http.MethodGet, "/users/1", nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusUnauthorized)
	}
}

func stringifyInt64(v int64) string {
	return strconv.FormatInt(v, 10)
}
