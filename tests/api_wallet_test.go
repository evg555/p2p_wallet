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

func TestCreateWallet_Contract_Created(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, time.March, 6, 12, 0, 0, 0, time.UTC)
	var gotInput dto.CreateWalletInput
	var gotInputSet bool

	svc := &walletServiceMock{
		createWalletFn: func(_ context.Context, input dto.CreateWalletInput) (*domain.Wallet, error) {
			gotInput = input
			gotInputSet = true
			return &domain.Wallet{
				ID:        1,
				UserID:    10,
				Currency:  domain.CurrencyUSD,
				Status:    domain.StatusActive,
				CreatedAt: createdAt,
			}, nil
		},
	}
	server := newTestWalletServer(svc)

	req := httptest.NewRequest(http.MethodPost, "/wallets", strings.NewReader(`{
		"user_id":10,
		"currency":"USD"
	}`))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
	require.True(t, gotInputSet)
	require.Equal(t, int64(10), gotInput.UserID)
	require.Equal(t, "USD", gotInput.Currency)

	var resp api.WalletResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	require.Equal(t, int64(1), resp.Id)
	require.Equal(t, int64(10), resp.UserId)
	require.Equal(t, api.USD, resp.Currency)
	require.Equal(t, api.WalletResponseStatusActive, resp.Status)
	require.Equal(t, createdAt, resp.CreatedAt)
}

func TestCreateWallet_Contract_Conflict(t *testing.T) {
	t.Parallel()

	svc := &walletServiceMock{
		createWalletFn: func(_ context.Context, _ dto.CreateWalletInput) (*domain.Wallet, error) {
			return nil, errs.ErrWalletAlreadyExist
		},
	}
	server := newTestWalletServer(svc)

	req := httptest.NewRequest(http.MethodPost, "/wallets", strings.NewReader(`{
		"user_id":10,
		"currency":"USD"
	}`))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	require.Equal(t, http.StatusConflict, rec.Code)
	require.Contains(t, rec.Body.String(), "wallet already exist")

}

func TestCreateWallet_Contract_Bad_Request(t *testing.T) {
	t.Parallel()

	svc := &walletServiceMock{}
	server := newTestWalletServer(svc)

	req := httptest.NewRequest(http.MethodPost, "/wallets", nil)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateWallet_Contract_Unauthorized(t *testing.T) {
	t.Parallel()

	svc := &walletServiceMock{
		createWalletFn: func(_ context.Context, _ dto.CreateWalletInput) (*domain.Wallet, error) {
			return nil, errs.ErrSessionNotFound
		},
	}
	server := newTestWalletServer(svc)

	req := httptest.NewRequest(http.MethodPost, "/wallets", strings.NewReader(`{
		"user_id":10,
		"currency":"USD"
	}`))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
	require.Contains(t, rec.Body.String(), "session not found")
}

func TestCreateWallet_Contract_InternalServerError(t *testing.T) {
	t.Parallel()

	svc := &walletServiceMock{
		createWalletFn: func(_ context.Context, _ dto.CreateWalletInput) (*domain.Wallet, error) {
			return nil, errors.New("database unavailable")
		},
	}
	server := newTestWalletServer(svc)

	req := httptest.NewRequest(http.MethodPost, "/wallets", strings.NewReader(`{
		"user_id":10,
		"currency":"USD"
	}`))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	resp, err := io.ReadAll(rec.Body)
	require.NoError(t, err)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Equal(t, "internal server error\n", string(resp))
}

func TestListUserWallets_Contract_Ok(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, time.March, 6, 10, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, time.March, 6, 11, 0, 0, 0, time.UTC)
	var gotUserID *int64

	svc := &walletServiceMock{
		listWalletsFn: func(_ context.Context, userID int64) ([]*domain.Wallet, error) {
			gotUserID = &userID
			return []*domain.Wallet{
				{
					ID:          1,
					UserID:      domain.UserID(userID),
					Currency:    domain.CurrencyUSD,
					Status:      domain.StatusActive,
					TotalAmount: 10010,
					HeldAmount:  100,
					CreatedAt:   createdAt,
					UpdatedAt:   &updatedAt,
				},
			}, nil
		},
	}
	server := newTestWalletServer(svc)

	req := httptest.NewRequest(http.MethodGet, "/wallets/10", nil)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, gotUserID)
	require.Equal(t, int64(10), *gotUserID)

	var resp api.ListWalletsResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	require.Len(t, resp.Wallets, 1)
	require.Equal(t, int64(1), resp.Wallets[0].Id)
	require.Equal(t, api.WalletItemCurrencyUSD, resp.Wallets[0].Currency)
	require.Equal(t, int64(10010), resp.Wallets[0].TotalAmount)
	require.Equal(t, int64(100), resp.Wallets[0].HeldAmount)
	require.Equal(t, api.WalletItemStatusActive, resp.Wallets[0].Status)
	require.Equal(t, createdAt, resp.Wallets[0].CreatedAt)
	require.Equal(t, updatedAt, resp.Wallets[0].UpdatedAt)
}

func TestListUserWallets_Contract_Unauthorized(t *testing.T) {
	t.Parallel()

	svc := &walletServiceMock{
		listWalletsFn: func(_ context.Context, _ int64) ([]*domain.Wallet, error) {
			return nil, errs.ErrSessionNotFound
		},
	}
	server := newTestWalletServer(svc)

	req := httptest.NewRequest(http.MethodGet, "/wallets/10", nil)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
	require.Contains(t, rec.Body.String(), "session not found")
}

func TestListUserWallets_Contract_InternalServerError(t *testing.T) {
	t.Parallel()

	svc := &walletServiceMock{
		listWalletsFn: func(_ context.Context, _ int64) ([]*domain.Wallet, error) {
			return nil, errors.New("database unavailable")
		},
	}
	server := newTestWalletServer(svc)

	req := httptest.NewRequest(http.MethodGet, "/wallets/10", nil)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	resp, err := io.ReadAll(rec.Body)
	require.NoError(t, err)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Equal(t, "internal server error\n", string(resp))
}

func newTestWalletServer(svc handler.WalletService) http.Handler {
	h := handler.New(nil, svc)
	return api.Handler(api.NewStrictHandlerWithOptions(h, nil, api.StrictHTTPServerOptions{
		RequestErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		},
		ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, "internal server error", http.StatusInternalServerError)
		},
	}))
}
