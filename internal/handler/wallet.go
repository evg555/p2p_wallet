package handler

import (
	"context"
	"errors"
	"time"

	"p2p_wallet/internal/api"
	"p2p_wallet/internal/domain"
	"p2p_wallet/internal/dto"
	"p2p_wallet/internal/errs"
)

type WalletService interface {
	CreateWallet(ctx context.Context, input dto.CreateWalletInput) (*domain.Wallet, error)
	ListWallets(ctx context.Context, userID int64) ([]*domain.Wallet, error)
}

func (h *handler) CreateWallet(ctx context.Context, req api.CreateWalletRequestObject) (api.CreateWalletResponseObject, error) {
	if req.Body == nil {
		return api.CreateWallet400JSONResponse{
			Code:    "bad request",
			Message: "request body is required",
		}, nil
	}

	wallet, err := h.walletSrv.CreateWallet(ctx, reqToCreateWalletDTO(req.Body))
	if err != nil {
		var resp api.CreateWalletResponseObject
		switch true {
		case errors.Is(err, errs.ErrSessionNotFound):
			resp = api.CreateWallet401JSONResponse{
				Code:    "not found",
				Message: err.Error(),
			}
		case errors.Is(err, errs.ErrWalletAlreadyExist):
			resp = api.CreateWallet409JSONResponse{
				Code:    "conflict",
				Message: err.Error(),
			}
		default:
			return nil, err
		}

		return resp, nil
	}

	resp := api.CreateWallet201JSONResponse{
		Id:        int64(wallet.ID),
		UserId:    int64(wallet.UserID),
		Currency:  api.WalletResponseCurrency(wallet.Currency),
		Status:    api.WalletResponseStatus(wallet.Status),
		CreatedAt: wallet.CreatedAt,
	}

	return resp, nil
}

func (h *handler) ListUserWallets(ctx context.Context, req api.ListUserWalletsRequestObject) (api.ListUserWalletsResponseObject, error) {
	wallets, err := h.walletSrv.ListWallets(ctx, req.UserId)
	if err != nil {
		var resp api.ListUserWalletsResponseObject
		switch true {
		case errors.Is(err, errs.ErrSessionNotFound):
			resp = api.ListUserWallets401JSONResponse{
				Code:    "not found",
				Message: err.Error(),
			}
		case errors.Is(err, errs.ErrAccessDenied):
			resp = api.ListUserWallets403JSONResponse{
				Code:    "conflict",
				Message: err.Error(),
			}
		default:
			return nil, err
		}

		return resp, nil
	}

	respWallets := make([]api.WalletItem, 0, len(wallets))

	for _, wallet := range wallets {
		var updatedAt time.Time
		if wallet.UpdatedAt != nil {
			updatedAt = *wallet.UpdatedAt
		}

		respWallets = append(respWallets, api.WalletItem{
			Id:          int64(wallet.ID),
			Currency:    api.WalletItemCurrency(wallet.Currency),
			Status:      api.WalletItemStatus(wallet.Status),
			HeldAmount:  wallet.HeldAmount,
			TotalAmount: wallet.TotalAmount,
			CreatedAt:   wallet.CreatedAt,
			UpdatedAt:   updatedAt,
		})
	}

	return api.ListUserWallets200JSONResponse{Wallets: respWallets}, nil
}
