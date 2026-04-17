package service

import (
	"context"
	"errors"
	"fmt"

	"p2p_wallet/internal/domain"
	"p2p_wallet/internal/dto"
	"p2p_wallet/internal/errs"
)

type balanceService struct {
	log         Logger
	userRepo    UserRepo
	walletRepo  WalletRepo
	sessionRepo SessionRepo
}

func NewBalanceService(
	log Logger,
	userRepo UserRepo,
	walletRepo WalletRepo,
	sessionRepo SessionRepo,
) *balanceService {
	return &balanceService{
		log:         log,
		userRepo:    userRepo,
		walletRepo:  walletRepo,
		sessionRepo: sessionRepo,
	}
}

func (b *balanceService) Transfer(ctx context.Context, input dto.TransferBalanceInput) (*domain.Transaction, error) {
	walletFrom, err := b.walletRepo.FindByID(ctx, domain.WalletID(input.FromWalletID))
	if err != nil {
		if errors.Is(err, errs.ErrWalletNotFound) {
			b.log.Warn("transfer wallet failed", withReqID(ctx, "from_wallet_id", input.FromWalletID, "error", "wallet not found")...)
		}
		return nil, fmt.Errorf("transfer wallet failed: %w", err)
	}

	_, err = b.findUser(ctx, walletFrom.UserID)
	if err != nil {
		return nil, err
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

// TODO: extract to middleware ?
func (b *balanceService) findUser(ctx context.Context, userID domain.UserID) (*domain.User, error) {
	user, err := b.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			b.log.Warn("transfer wallet failed", withReqID(ctx, "user_id", userID, "error", err.Error())...)
		}

		return nil, fmt.Errorf("userRepo: failed to get user by id: %w", err)
	}

	sessionID, ok := ctx.Value(domain.CtxKey(domain.SessionKey)).(string)
	if !ok {
		b.log.Warn("transfer wallet failed", withReqID(ctx, "user_id", userID, "error", "session not found")...)
		return nil, errs.ErrSessionNotFound
	}

	existSession, err := b.sessionRepo.Get(domain.SessionID(sessionID))
	if err != nil || !existSession.Equal(sessionID) {
		b.log.Warn("transfer wallet failed", withReqID(ctx, "user_id", userID, "error", "session not found")...)
		return nil, errs.ErrSessionNotFound
	}

	return user, nil
}
