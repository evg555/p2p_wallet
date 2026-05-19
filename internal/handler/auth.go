package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"p2p_wallet/internal/api"
	"p2p_wallet/internal/domain"
	"p2p_wallet/internal/dto"
	"p2p_wallet/internal/errs"
)

type AuthService interface {
	Register(ctx context.Context, input dto.RegisterInput) (*domain.User, error)
	Login(ctx context.Context, input dto.LoginInput) (*domain.AuthResult, error)
	Logout(ctx context.Context)
}

func (h *handler) LoginUser(ctx context.Context, req api.LoginUserRequestObject) (api.LoginUserResponseObject, error) {
	res, err := h.userSrv.Login(ctx, reqToLoginDTO(req.Body))
	if err != nil {
		var resp api.LoginUserResponseObject
		switch true {
		case errors.Is(err, errs.ErrUserNotFound):
			resp = api.LoginUser404JSONResponse{
				Code:    "not found",
				Message: err.Error(),
			}
		case errors.Is(err, errs.ErrPasswordMismatch):
			resp = api.LoginUser401JSONResponse{
				Code:    "unauthorized",
				Message: err.Error(),
			}
		default:
			return nil, err
		}

		return resp, nil
	}

	resp := api.LoginUser200JSONResponse{
		Body: api.LoginResponse{
			Id:       int64(res.UserID),
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
	user, err := h.userSrv.Register(ctx, reqToRegisterDTO(req.Body))
	if err != nil {
		var resp api.RegisterUserResponseObject
		switch true {
		case errors.Is(err, errs.ErrUserAlreadyExist):
			resp = api.RegisterUser409JSONResponse{
				Code:    "conflict",
				Message: err.Error(),
			}
		default:
			return nil, err
		}

		return resp, nil
	}

	resp := api.RegisterUser201JSONResponse{
		CreatedAt: user.CreatedAt,
		Id:        int64(user.ID),
		LastName:  user.LastName,
		Login:     user.Login,
		Name:      user.FirstName,
	}

	return resp, nil
}

func (h *handler) LogoutUser(ctx context.Context, _ api.LogoutUserRequestObject) (api.LogoutUserResponseObject, error) {
	h.userSrv.Logout(ctx)

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
