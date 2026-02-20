//nolint:revive
package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"p2p_wallet/internal/api"
	"p2p_wallet/internal/domain"
)

var defaultSessionID = "abc123"

type ctxKey string

type Service interface {
	Register(ctx context.Context, input api.RegisterRequest) (*domain.User, error)
	Login(ctx context.Context, input api.LoginRequest) (*domain.User, error)
	Logout(ctx context.Context, id int64) error
	GetUser(ctx context.Context, id int64) (*domain.User, error)
}

type handler struct {
	userSrv Service
}

var _ api.ServerInterface = (*handler)(nil)

func New(srv Service) *handler {
	return &handler{userSrv: srv}
}

func (h *handler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var req api.LoginRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		resp := api.ErrorResponse{
			Code:    "wrong parameters",
			Message: err.Error(),
		}

		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	user, err := h.userSrv.Login(r.Context(), req)
	if err != nil {
		resp := api.ErrorResponse{
			Code:    "cannot login user",
			Message: err.Error(),
		}

		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	resp := api.UserResponse{
		CreatedAt: user.CreatedAt,
		Id:        user.ID,
		LastName:  user.LastName,
		Login:     user.Login,
		Name:      user.FirstName,
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    defaultSessionID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   3600,
	})

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req api.RegisterRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		resp := api.ErrorResponse{
			Code:    "wrong parameters",
			Message: err.Error(),
		}

		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	user, err := h.userSrv.Register(r.Context(), req)
	if err != nil {
		resp := api.ErrorResponse{
			Code:    "cannot create user",
			Message: err.Error(),
		}

		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	resp := api.UserResponse{
		CreatedAt: user.CreatedAt,
		Id:        user.ID,
		LastName:  user.LastName,
		Login:     user.Login,
		Name:      user.FirstName,
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *handler) GetUserById(w http.ResponseWriter, r *http.Request, id api.UserId) {
	if id <= 0 {
		resp := api.ErrorResponse{
			Code:    "invalid user id",
			Message: "user id must be greater than zero",
		}

		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	c, err := r.Cookie("session_id")
	if err != nil {
		resp := api.ErrorResponse{
			Code:    "authorization failed",
			Message: "user session is empty",
		}

		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	ctx := context.WithValue(r.Context(), ctxKey("sessionID"), c.Value)

	user, err := h.userSrv.GetUser(ctx, id)
	if err != nil {
		resp := api.ErrorResponse{
			Code:    "cannot get user",
			Message: err.Error(),
		}

		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	resp := api.UserResponse{
		CreatedAt: user.CreatedAt,
		Id:        user.ID,
		LastName:  user.LastName,
		Login:     user.Login,
		Name:      user.FirstName,
		UpdatedAt: user.UpdatedAt,
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *handler) LogoutUser(w http.ResponseWriter, r *http.Request, id api.UserId) {
	if id <= 0 {
		resp := api.ErrorResponse{
			Code:    "invalid user id",
			Message: "user id must be greater than zero",
		}

		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	c, err := r.Cookie("session_id")
	if err != nil {
		resp := api.ErrorResponse{
			Code:    "authorization failed",
			Message: "user session is empty",
		}

		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	ctx := context.WithValue(r.Context(), ctxKey("sessionID"), c.Value)

	err = h.userSrv.Logout(ctx, id)
	if err != nil {
		resp := api.ErrorResponse{
			Code:    "cannot get user",
			Message: err.Error(),
		}

		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
