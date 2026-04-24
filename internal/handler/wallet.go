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
	ListWallets(ctx context.Context) ([]*domain.Wallet, error)
}

func (h *handler) CreateWallet(ctx context.Context, req api.CreateWalletRequestObject) (api.CreateWalletResponseObject, error) {
	wallet, err := h.walletSrv.CreateWallet(ctx, reqToCreateWalletDTO(req.Body))
	if err != nil {
		var resp api.CreateWalletResponseObject
		switch true {
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
		Currency:  api.WalletResponseCurrency(wallet.Currency.String()),
		Status:    api.WalletResponseStatus(wallet.Status.String()),
		CreatedAt: wallet.CreatedAt,
	}

	return resp, nil
}

func (h *handler) ListUserWallets(ctx context.Context, _ api.ListUserWalletsRequestObject) (api.ListUserWalletsResponseObject, error) {
	wallets, err := h.walletSrv.ListWallets(ctx)
	if err != nil {
		return nil, err
	}

	respWallets := make([]api.WalletItem, 0, len(wallets))

	for _, wallet := range wallets {
		var updatedAt time.Time
		if wallet.UpdatedAt != nil {
			updatedAt = *wallet.UpdatedAt
		}

		respWallets = append(respWallets, api.WalletItem{
			Id:          int64(wallet.ID),
			Currency:    api.WalletItemCurrency(wallet.Currency.String()),
			Status:      api.WalletItemStatus(wallet.Status.String()),
			HeldAmount:  wallet.HeldAmount,
			TotalAmount: wallet.TotalAmount,
			CreatedAt:   wallet.CreatedAt,
			UpdatedAt:   updatedAt,
		})
	}

	return api.ListUserWallets200JSONResponse{Wallets: respWallets}, nil
}
