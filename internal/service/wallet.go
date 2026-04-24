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

type walletService struct {
	log        Logger
	walletRepo WalletRepo
}

func NewWalletService(
	log Logger,
	walletRepo WalletRepo,
) *walletService {
	return &walletService{
		log:        log,
		walletRepo: walletRepo,
	}
}

func (s *walletService) CreateWallet(ctx context.Context, input dto.CreateWalletInput) (*domain.Wallet, error) {
	user := authctx.CurrentUser(ctx)

	wallet, err := domain.NewWallet(user.ID, input.Currency)
	if err != nil {
		s.log.Warn("create wallet failed", withReqID(
			ctx, "user_id", user.ID, "currency", input.Currency, "error", err.Error())...,
		)
		return nil, err
	}

	savedWallet, err := s.walletRepo.Save(ctx, wallet)
	if err != nil {
		if errors.Is(err, errs.ErrWalletAlreadyExist) {
			s.log.Warn("create wallet failed", withReqID(
				ctx, "user_id", user.ID, "currency", input.Currency, "error", err.Error())...,
			)
		}
		return nil, err
	}

	s.log.Info("wallet created successfully", withReqID(
		ctx, "wallet_id", savedWallet.ID, "user_id", savedWallet.UserID, "currency", savedWallet.Currency,
	)...)
	return savedWallet, nil
}

func (s *walletService) ListWallets(ctx context.Context) ([]*domain.Wallet, error) {
	user := authctx.CurrentUser(ctx)

	wallets, err := s.walletRepo.FindByUserID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("walletRepo: failed to get wallets by user id: %w", err)
	}

	return wallets, nil
}
