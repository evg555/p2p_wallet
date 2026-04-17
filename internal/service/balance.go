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
	walletSource, err := b.walletRepo.FindByID(ctx, domain.WalletID(input.FromWalletID))
	if err != nil {
		b.log.Warn("transfer wallet failed", withReqID(ctx, "from_wallet_id", input.FromWalletID, "error", "wallet not found")...)
		return nil, fmt.Errorf("transfer wallet failed: %w", err)
	}

	user, err := b.findUser(ctx, walletSource.UserID)
	if err != nil {
		return nil, err
	}

	_ = user
	return nil, nil
}

func (b *balanceService) findUser(ctx context.Context, userID domain.UserID) (*domain.User, error) {
	user, err := b.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			b.log.Warn("create wallet failed", withReqID(ctx, "user_id", userID, "error", err.Error())...)
		}

		return nil, fmt.Errorf("userRepo: failed to get user by id: %w", err)
	}

	sessionID, ok := ctx.Value(domain.CtxKey(domain.SessionKey)).(string)
	if !ok {
		b.log.Warn("create wallet failed", withReqID(ctx, "user_id", userID, "error", "session not found")...)
		return nil, errs.ErrSessionNotFound
	}

	existSession, err := b.sessionRepo.Get(domain.SessionID(sessionID))
	if err != nil || !existSession.Equal(sessionID) {
		b.log.Warn("create wallet failed", withReqID(ctx, "user_id", userID, "error", "session not found")...)
		return nil, errs.ErrSessionNotFound
	}

	return user, nil
}
