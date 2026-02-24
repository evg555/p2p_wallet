package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"

	"p2p_wallet/internal/api"
	"p2p_wallet/internal/errs"
)

func writeAPIError(w http.ResponseWriter, r *http.Request, err error) {
	var (
		apiErr  *errs.APIError
		errResp api.ErrorResponse
	)

	w.Header().Set("Content-Type", "application/json")

	if errors.As(err, &apiErr) {
		switch apiErr.Code {
		case http.StatusNotFound:
			w.WriteHeader(http.StatusNotFound)
			errResp = api.ErrorResponse{
				Code: "not_found",
			}
		case http.StatusBadRequest:
			w.WriteHeader(http.StatusBadRequest)
			errResp = api.ErrorResponse{
				Code: "bad request",
			}
		case http.StatusUnauthorized:
			w.WriteHeader(http.StatusUnauthorized)
			errResp = api.ErrorResponse{
				Code: "unauthorized",
			}
		case http.StatusConflict:
			w.WriteHeader(http.StatusConflict)
			errResp = api.ErrorResponse{
				Code: "conflict",
			}
		case http.StatusForbidden:
			w.WriteHeader(http.StatusForbidden)
			errResp = api.ErrorResponse{
				Code: "access denied",
			}
		default:
			w.WriteHeader(http.StatusInternalServerError)
			apiErr.Message = ""
			errResp = api.ErrorResponse{
				Code: "internal server error",
			}
		}
	}

	errResp.Message = apiErr.Message
	_ = json.NewEncoder(w).Encode(errResp)
}
