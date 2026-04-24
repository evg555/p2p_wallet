package service

import (
	"context"
	"errors"
	"fmt"

	"p2p_wallet/internal/domain"
	"p2p_wallet/internal/domain/helpers/authctx"
	"p2p_wallet/internal/dto"
	"p2p_wallet/internal/errs"
)

type balanceService struct {
	log        Logger
	walletRepo WalletRepo
}

func NewBalanceService(
	log Logger,
	walletRepo WalletRepo,
) *balanceService {
	return &balanceService{
		log:        log,
		walletRepo: walletRepo,
	}
}

func (b *balanceService) Transfer(ctx context.Context, input dto.TransferBalanceInput) (*domain.Transaction, error) {
	user := authctx.CurrentUser(ctx)

	walletFrom, err := b.walletRepo.FindByID(ctx, domain.WalletID(input.FromWalletID))
	if err != nil {
		if errors.Is(err, errs.ErrWalletNotFound) {
			b.log.Warn("transfer wallet failed", withReqID(ctx, "from_wallet_id", input.FromWalletID, "error", "wallet not found")...)
		}
		return nil, fmt.Errorf("transfer wallet failed: %w", err)
	}

	if user.ID != walletFrom.UserID {
		return nil, fmt.Errorf("transfer wallet failed: wallet %d doesn't belong to user %d: %w", walletFrom.ID, user.ID, errs.ErrAccessDenied)
	}

	walletTo, err := b.walletRepo.FindByID(ctx, domain.WalletID(input.ToWalletID))
	if err != nil {
		if errors.Is(err, errs.ErrWalletNotFound) {
			b.log.Warn("transfer wallet failed", withReqID(ctx, "to_wallet_id", input.ToWalletID, "error", "wallet not found")...)
		}
		return nil, fmt.Errorf("transfer wallet failed: %w", err)
	}

	transaction, err := domain.NewTransaction(input.Amount, walletFrom, walletTo)
	if err != nil {
		b.log.Warn("transfer wallet failed", withReqID(
			ctx,
			"amount", input.Amount,
			"from_wallet_id", input.ToWalletID,
			"to_wallet_id", input.ToWalletID,
			"error", "wallet not found",
		)...)
		return nil, fmt.Errorf("transfer wallet failed: %w", err)
	}

	// TODO: save transaction

	return transaction, nil
}
