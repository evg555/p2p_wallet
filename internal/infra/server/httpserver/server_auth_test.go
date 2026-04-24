package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"p2p_wallet/internal/api"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	middleware "github.com/oapi-codegen/nethttp-middleware"
	"github.com/stretchr/testify/require"
)

func TestValidator_ProtectedRouteWithoutCookie_ReturnsUnauthorizedJSON(t *testing.T) {
	t.Parallel()

	loader := openapi3.NewLoader()
	swagger, err := loader.LoadFromFile("../../../../spec/openapi/p2p-wallet.yaml")
	require.NoError(t, err)

	validator := middleware.OapiRequestValidatorWithOptions(swagger, &middleware.Options{
		Options: openapi3filter.Options{
			AuthenticationFunc: authenticateRequest,
		},
		ErrorHandlerWithOpts: func(_ context.Context, err error, w http.ResponseWriter, _ *http.Request, opts middleware.ErrorHandlerOpts) {
			var securityErr *openapi3filter.SecurityRequirementsError
			if errors.As(err, &securityErr) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(api.ErrorResponse{
					Code:    "unauthorized",
					Message: err.Error(),
				})
				return
			}

			http.Error(w, err.Error(), opts.StatusCode)
		},
	})

	router := validator(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/wallets/me", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
	require.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var resp api.ErrorResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	require.Equal(t, "unauthorized", resp.Code)
	require.Contains(t, resp.Message, "security requirements failed")
}
