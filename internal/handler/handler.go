//nolint:revive
package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"p2p_wallet/internal/api"
	"p2p_wallet/internal/domain"
	"p2p_wallet/internal/errs"
)

type Service interface {
	Register(ctx context.Context, input *api.RegisterRequest) (*domain.User, error)
	Login(ctx context.Context, input *api.LoginRequest) (*domain.AuthResult, error)
	Logout(ctx context.Context, id int64) error
}

type handler struct {
	userSrv Service
}

var _ api.StrictServerInterface = (*handler)(nil)

func New(srv Service) *handler {
	return &handler{userSrv: srv}
}

func (h *handler) LoginUser(ctx context.Context, req api.LoginUserRequestObject) (api.LoginUserResponseObject, error) {
	if req.Body == nil {
		return nil, &errs.APIError{
			Code:    http.StatusBadRequest,
			Message: "request body is required",
		}
	}

	res, err := h.userSrv.Login(ctx, req.Body)
	if err != nil {
		var apiErr *errs.APIError
		switch true {
		case errors.Is(err, errs.ErrUserNotFound):
			apiErr = &errs.APIError{
				Code:    http.StatusNotFound,
				Message: err.Error(),
			}
		case errors.Is(err, errs.ErrPasswordMismatch):
			apiErr = &errs.APIError{
				Code:    http.StatusUnauthorized,
				Message: err.Error(),
			}
		default:
			apiErr = &errs.APIError{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			}
		}

		return nil, apiErr
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

	return resp, nil
}

func (h *handler) RegisterUser(ctx context.Context, req api.RegisterUserRequestObject) (api.RegisterUserResponseObject, error) {
	if req.Body == nil {
		return nil, &errs.APIError{
			Code:    http.StatusBadRequest,
			Message: "request body is required",
		}
	}

	user, err := h.userSrv.Register(ctx, req.Body)
	if err != nil {
		var apiErr *errs.APIError
		switch true {
		case errors.Is(err, errs.ErrUserAlreadyExist):
			apiErr = &errs.APIError{
				Code:    http.StatusConflict,
				Message: err.Error(),
			}
		default:
			apiErr = &errs.APIError{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			}
		}

		return nil, apiErr
	}

	resp := api.RegisterUser201JSONResponse{
		CreatedAt: user.CreatedAt,
		Id:        user.ID,
		LastName:  user.LastName,
		Login:     user.Login,
		Name:      user.FirstName,
	}

	return resp, nil
}

func (h *handler) LogoutUser(ctx context.Context, req api.LogoutUserRequestObject) (api.LogoutUserResponseObject, error) {
	//c, err := req.Cookie(domain.SessionKey)
	//if err != nil || c.Value == "" {
	//	resp := api.LogoutUser401JSONResponse{
	//		Code:    code,
	//		Message: "user session is empty",
	//	}
	//	_ = resp.VisitLogoutUserResponse(w)
	//	return
	//}
	//
	//ctx := context.WithValue(r.Context(), domain.CtxKey(domain.SessionKey), c.Value)

	err := h.userSrv.Logout(ctx, req.Id)
	if err != nil {
		var apiErr *errs.APIError
		switch true {
		case errors.Is(err, errs.ErrUserNotFound):
			apiErr = &errs.APIError{
				Code:    http.StatusNotFound,
				Message: err.Error(),
			}
		case errors.Is(err, errs.ErrSessionNotFound):
			apiErr = &errs.APIError{
				Code:    http.StatusUnauthorized,
				Message: err.Error(),
			}
		case errors.Is(err, errs.ErrAccessDenied):
			apiErr = &errs.APIError{
				Code:    http.StatusForbidden,
				Message: err.Error(),
			}
		default:
			apiErr = &errs.APIError{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			}
		}

		return nil, apiErr
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

	return resp, nil
}
