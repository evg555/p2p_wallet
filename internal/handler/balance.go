package handler

import (
	"context"
	"errors"

	"p2p_wallet/internal/api"
	"p2p_wallet/internal/domain"
	"p2p_wallet/internal/dto"
	"p2p_wallet/internal/errs"
)

type BalanceService interface {
	Transfer(ctx context.Context, input dto.TransferBalanceInput) (*domain.Transaction, error)
}

func (h *handler) TransferBalance(ctx context.Context, req api.TransferBalanceRequestObject) (api.TransferBalanceResponseObject, error) {
	if req.Body == nil {
		return api.TransferBalance400JSONResponse{
			Code:    "bad request",
			Message: "request body is required",
		}, nil
	}

	transaction, err := h.balanceSrv.Transfer(ctx, reqToTransferBalanceDTO(req.Params, req.Body))
	if err != nil {
		var resp api.TransferBalanceResponseObject
		switch {
		case errors.Is(err, errs.ErrSessionNotFound) || errors.Is(err, errs.ErrUserNotFound):
			resp = api.TransferBalance401JSONResponse{
				Code:    "unauthorized",
				Message: err.Error(),
			}
		case errors.Is(err, errs.ErrWalletNotFound) ||
			errors.Is(err, errs.ErrWalletMismatch) ||
			errors.Is(err, errs.ErrNotEnoughMoney) ||
			errors.Is(err, errs.ErrCurrencyMismatch):
			resp = api.TransferBalance422JSONResponse{
				Code:    "unprocessable entity",
				Message: err.Error(),
			}
		default:
			return nil, err
		}

		return resp, nil
	}

	entries := make([]api.LedgerEntry, 0, len(transaction.Entries))

	for _, entry := range transaction.Entries {
		entries = append(entries, api.LedgerEntry{
			Amount:   entry.Amount,
			Currency: api.LedgerEntryCurrency(entry.Currency),
			Id:       entry.ID,
			WalletId: entry.WalletID,
		})
	}

	resp := api.TransferBalance200JSONResponse{
		Transaction: api.LedgerTransaction{
			Id:        transaction.ID,
			Status:    api.LedgerTransactionStatus(transaction.Status),
			CreatedAt: transaction.CreatedAt,
			Entries:   entries,
		},
	}

	return resp, nil

}
