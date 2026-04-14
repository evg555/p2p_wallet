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

func TestTransferBalance_Contract_Ok(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, time.March, 7, 15, 0, 0, 0, time.UTC)
	var gotInput dto.TransferBalanceInput
	var gotInputSet bool

	svc := &balanceServiceMock{
		transferFn: func(_ context.Context, input dto.TransferBalanceInput) (*domain.Transaction, error) {
			gotInput = input
			gotInputSet = true
			return &domain.Transaction{
				ID:        101,
				Status:    "completed",
				CreatedAt: createdAt,
				Entries: []domain.Entry{
					{
						ID:       1,
						WalletID: 10,
						Amount:   -500,
						Currency: "USD",
					},
					{
						ID:       2,
						WalletID: 20,
						Amount:   500,
						Currency: "USD",
					},
				},
			}, nil
		},
	}
	server := newTestBalanceServer(svc)

	req := httptest.NewRequest(http.MethodPost, "/balance/transfer", strings.NewReader(`{
		"from_wallet_id":10,
		"to_wallet_id":20,
		"amount":500
	}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Idempotency-Key", "transfer-123")

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.True(t, gotInputSet)
	require.Equal(t, "transfer-123", gotInput.IdempotencyKey)
	require.Equal(t, int64(10), gotInput.FromWalletID)
	require.Equal(t, int64(20), gotInput.ToWalletID)
	require.Equal(t, int64(500), gotInput.Amount)

	var resp api.BalanceTransferResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	require.Equal(t, int64(101), resp.Transaction.Id)
	require.Equal(t, api.LedgerTransactionStatus("completed"), resp.Transaction.Status)
	require.Equal(t, createdAt, resp.Transaction.CreatedAt)
	require.Len(t, resp.Transaction.Entries, 2)
	require.Equal(t, int64(1), resp.Transaction.Entries[0].Id)
	require.Equal(t, int64(10), resp.Transaction.Entries[0].WalletId)
	require.Equal(t, int64(-500), resp.Transaction.Entries[0].Amount)
	require.Equal(t, api.LedgerEntryCurrency("USD"), resp.Transaction.Entries[0].Currency)
	require.Equal(t, int64(2), resp.Transaction.Entries[1].Id)
	require.Equal(t, int64(20), resp.Transaction.Entries[1].WalletId)
	require.Equal(t, int64(500), resp.Transaction.Entries[1].Amount)
	require.Equal(t, api.LedgerEntryCurrency("USD"), resp.Transaction.Entries[1].Currency)
}

func TestTransferBalance_Contract_BadRequest(t *testing.T) {
	t.Parallel()

	svc := &balanceServiceMock{}
	server := newTestBalanceServer(svc)

	req := httptest.NewRequest(http.MethodPost, "/balance/transfer", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Idempotency-Key", "transfer-123")

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Contains(t, rec.Body.String(), "can't decode JSON body: EOF")
}

func TestTransferBalance_Contract_Unauthorized(t *testing.T) {
	t.Parallel()

	svc := &balanceServiceMock{
		transferFn: func(_ context.Context, _ dto.TransferBalanceInput) (*domain.Transaction, error) {
			return nil, errs.ErrSessionNotFound
		},
	}
	server := newTestBalanceServer(svc)

	req := httptest.NewRequest(http.MethodPost, "/balance/transfer", strings.NewReader(`{
		"from_wallet_id":10,
		"to_wallet_id":20,
		"amount":500
	}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Idempotency-Key", "transfer-123")

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
	require.Contains(t, rec.Body.String(), "session not found")
}

func TestTransferBalance_Contract_UnprocessableEntity(t *testing.T) {
	t.Parallel()

	svc := &balanceServiceMock{
		transferFn: func(_ context.Context, _ dto.TransferBalanceInput) (*domain.Transaction, error) {
			return nil, errs.ErrNotEnoughMoney
		},
	}
	server := newTestBalanceServer(svc)

	req := httptest.NewRequest(http.MethodPost, "/balance/transfer", strings.NewReader(`{
		"from_wallet_id":10,
		"to_wallet_id":20,
		"amount":500
	}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Idempotency-Key", "transfer-123")

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	require.Contains(t, rec.Body.String(), "not enough money for transfer")
}

func TestTransferBalance_Contract_InternalServerError(t *testing.T) {
	t.Parallel()

	svc := &balanceServiceMock{
		transferFn: func(_ context.Context, _ dto.TransferBalanceInput) (*domain.Transaction, error) {
			return nil, errors.New("database unavailable")
		},
	}
	server := newTestBalanceServer(svc)

	req := httptest.NewRequest(http.MethodPost, "/balance/transfer", strings.NewReader(`{
		"from_wallet_id":10,
		"to_wallet_id":20,
		"amount":500
	}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Idempotency-Key", "transfer-123")

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	resp, err := io.ReadAll(rec.Body)
	require.NoError(t, err)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Equal(t, "internal server error\n", string(resp))
}

func newTestBalanceServer(svc handler.BalanceService) http.Handler {
	h := handler.New(nil, nil, svc)
	return api.Handler(api.NewStrictHandlerWithOptions(h, nil, api.StrictHTTPServerOptions{
		RequestErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		},
		ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, "internal server error", http.StatusInternalServerError)
		},
	}))
}
