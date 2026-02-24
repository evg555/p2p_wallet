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
	"p2p_wallet/internal/infra/server/httpserver"
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
		return api.LoginUser400JSONResponse{
			Code:    "wrong parameters",
			Message: "request body is required",
		}, httpserver.ErrBadRequest
	}

	res, err := h.userSrv.Login(ctx, req.Body)
	if err != nil {
		code := "cannot login user"
		var resp api.LoginUserResponseObject
		switch true {
		case errors.Is(err, errs.ErrUserNotFound):
			err = httpserver.ErrNotFound
			resp = api.LoginUser404JSONResponse{
				Code:    code,
				Message: err.Error(),
			}
		case errors.Is(err, errs.ErrPasswordMismatch):
			err = httpserver.ErrUnauthorized
			resp = api.LoginUser401JSONResponse{
				Code:    code,
				Message: err.Error(),
			}
		default:
			err = httpserver.ErrInternalServer
			resp = nil
		}

		return resp, err
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
		return api.RegisterUser400JSONResponse{
			Code:    "wrong parameters",
			Message: "request body is required",
		}, httpserver.ErrBadRequest
	}

	user, err := h.userSrv.Register(ctx, req.Body)
	if err != nil {
		code := "cannot create user"
		var resp api.RegisterUserResponseObject
		switch true {
		case errors.Is(err, errs.ErrUserAlreadyExist):
			err = httpserver.ErrConflict
			resp = api.RegisterUser409JSONResponse{
				Code:    code,
				Message: err.Error(),
			}
		default:
			err = httpserver.ErrInternalServer
			resp = nil
		}

		return resp, err
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
	code := "cannot logout user"

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
		var resp api.LogoutUserResponseObject
		switch true {
		case errors.Is(err, errs.ErrUserNotFound):
			err = httpserver.ErrNotFound
			resp = api.LogoutUser404JSONResponse{
				Code:    code,
				Message: err.Error(),
			}
		case errors.Is(err, errs.ErrSessionNotFound):
			err = httpserver.ErrUnauthorized
			resp = api.LogoutUser401JSONResponse{
				Code:    code,
				Message: err.Error(),
			}
		case errors.Is(err, errs.ErrAccessDenied):
			err = httpserver.ErrAccessDenied
			resp = api.LogoutUser403JSONResponse{
				Code:    code,
				Message: err.Error(),
			}
		default:
			err = httpserver.ErrInternalServer
			resp = nil
		}

		return resp, err
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
