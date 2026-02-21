//nolint:revive
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"p2p_wallet/internal/api"
	"p2p_wallet/internal/domain"
	"p2p_wallet/internal/errs"
)

type Service interface {
	Register(ctx context.Context, input api.RegisterRequest) (*domain.User, error)
	Login(ctx context.Context, input api.LoginRequest) (*domain.AuthResult, error)
	Logout(ctx context.Context, id int64) error
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
		resp := api.LoginUser400JSONResponse{
			Code:    "wrong parameters",
			Message: err.Error(),
		}
		_ = resp.VisitLoginUserResponse(w)
		return
	}

	res, err := h.userSrv.Login(r.Context(), req)
	if err != nil {
		code := "cannot login user"
		switch true {
		case errors.Is(err, errs.ErrUserNotFound):
			resp := api.LoginUser404JSONResponse{
				Code:    code,
				Message: err.Error(),
			}
			_ = resp.VisitLoginUserResponse(w)
		case errors.Is(err, errs.ErrPasswordMismatch):
			resp := api.LoginUser401JSONResponse{
				Code:    code,
				Message: err.Error(),
			}
			_ = resp.VisitLoginUserResponse(w)
		default:
			_ = response500Error(w, code, err)
		}

		return
	}

	resp := api.LoginUser200JSONResponse{
		Body: api.LoginResponse{
			Id:       res.UserID,
			LastName: res.UserLastName,
			Login:    res.UserLogin,
			Name:     res.UserName,
		},
		Headers: api.LoginUser200ResponseHeaders{
			SetCookie: (&http.Cookie{
				Name:     domain.SessionKey,
				Value:    string(res.SessionID),
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				MaxAge:   3600,
			}).String(),
		},
	}

	_ = resp.VisitLoginUserResponse(w)
}

func (h *handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req api.RegisterRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		resp := api.RegisterUser400JSONResponse{
			Code:    "wrong parameters",
			Message: err.Error(),
		}
		_ = resp.VisitRegisterUserResponse(w)
		return
	}

	user, err := h.userSrv.Register(r.Context(), req)
	if err != nil {
		code := "cannot create user"
		switch true {
		case errors.Is(err, errs.ErrUserAlreadyExist):
			resp := api.RegisterUser409JSONResponse{
				Code:    code,
				Message: err.Error(),
			}
			_ = resp.VisitRegisterUserResponse(w)
		default:
			_ = response500Error(w, code, err)
		}
		return
	}

	resp := api.RegisterUser201JSONResponse{
		CreatedAt: user.CreatedAt,
		Id:        user.ID,
		LastName:  user.LastName,
		Login:     user.Login,
		Name:      user.FirstName,
	}
	_ = resp.VisitRegisterUserResponse(w)
}

func (h *handler) LogoutUser(w http.ResponseWriter, r *http.Request, id api.UserId) {
	if id <= 0 {
		resp := api.LogoutUser400JSONResponse{
			Code:    "invalid user id",
			Message: "id must be greater than zero",
		}
		_ = resp.VisitLogoutUserResponse(w)
		return
	}

	code := "cannot logout user"

	c, err := r.Cookie(domain.SessionKey)
	if err != nil || c.Value == "" {
		resp := api.LogoutUser401JSONResponse{
			Code:    code,
			Message: "user session is empty",
		}
		_ = resp.VisitLogoutUserResponse(w)
		return
	}

	ctx := context.WithValue(r.Context(), domain.CtxKey(domain.SessionKey), c.Value)

	err = h.userSrv.Logout(ctx, id)
	if err != nil {
		switch true {
		case errors.Is(err, errs.ErrUserNotFound):
			resp := api.LogoutUser404JSONResponse{
				Code:    code,
				Message: err.Error(),
			}
			_ = resp.VisitLogoutUserResponse(w)
		case errors.Is(err, errs.ErrSessionNotFound):
			resp := api.LogoutUser401JSONResponse{
				Code:    code,
				Message: err.Error(),
			}
			_ = resp.VisitLogoutUserResponse(w)
		case errors.Is(err, errs.ErrAccessDenied):
			resp := api.LogoutUser403JSONResponse{
				Code:    code,
				Message: err.Error(),
			}
			_ = resp.VisitLogoutUserResponse(w)
		default:
			_ = response500Error(w, code, err)
		}
		return
	}

	resp := api.LogoutUser204Response{
		Headers: api.LogoutUser204ResponseHeaders{
			SetCookie: (&http.Cookie{
				Name:     domain.SessionKey,
				Value:    "",
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				Expires:  time.Unix(0, 0),
				MaxAge:   -1,
			}).String(),
		},
	}
	_ = resp.VisitLogoutUserResponse(w)
}

func response500Error(w http.ResponseWriter, code string, err error) error {
	response := api.ErrorResponse{
		Code:    code,
		Message: err.Error(),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)

	return json.NewEncoder(w).Encode(response)

}
