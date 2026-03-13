package service

import (
	"context"
	"errors"
	"fmt"

	"p2p_wallet/internal/domain"
	"p2p_wallet/internal/dto"
	"p2p_wallet/internal/errs"
)

type WalletRepo interface {
	Save(ctx context.Context, wallet *domain.Wallet) (*domain.Wallet, error)
	FindByUserID(ctx context.Context, userID domain.UserID) ([]*domain.Wallet, error)
}

type walletService struct {
	log         Logger
	userRepo    UserRepo
	walletRepo  WalletRepo
	sessionRepo SessionRepo
}

func NewWalletService(
	log Logger,
	userRepo UserRepo,
	walletRepo WalletRepo,
	sessionRepo SessionRepo,
) *walletService {
	return &walletService{
		log:         log,
		userRepo:    userRepo,
		walletRepo:  walletRepo,
		sessionRepo: sessionRepo,
	}
}

func (s *walletService) CreateWallet(ctx context.Context, input dto.CreateWalletInput) (*domain.Wallet, error) {
	user, err := s.findUser(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	wallet, err := domain.NewWallet(user.ID, input.Currency)
	if err != nil {
		s.log.Warn("create wallet failed", withReqID(
			ctx, "user_id", input.UserID, "currency", input.Currency, "error", err.Error())...,
		)
		return nil, err
	}

	savedWallet, err := s.walletRepo.Save(ctx, wallet)
	if err != nil {
		if errors.Is(err, errs.ErrWalletAlreadyExist) {
			s.log.Warn("create wallet failed", withReqID(
				ctx, "user_id", input.UserID, "currency", input.Currency, "error", err.Error())...,
			)
		}
		return nil, err
	}

	s.log.Info("wallet created successfully", withReqID(
		ctx, "wallet_id", savedWallet.ID, "user_id", savedWallet.UserID, "currency", savedWallet.Currency,
	)...)
	return savedWallet, nil
}

func (s *walletService) ListWallets(ctx context.Context, userID int64) ([]*domain.Wallet, error) {
	user, err := s.findUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	wallets, err := s.walletRepo.FindByUserID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("walletRepo: failed to get wallets by user id: %w", err)
	}

	return wallets, nil
}

func (s *walletService) findUser(ctx context.Context, userID int64) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			s.log.Warn("create wallet failed", withReqID(ctx, "user_id", userID, "error", err.Error())...)
		}

		return nil, fmt.Errorf("userRepo: failed to get user by id: %w", err)
	}

	sessionID, ok := ctx.Value(domain.CtxKey(domain.SessionKey)).(string)
	if !ok {
		s.log.Warn("create wallet failed", withReqID(ctx, "user_id", userID, "error", "session not found")...)
		return nil, errs.ErrSessionNotFound
	}

	existSession, err := s.sessionRepo.Get(domain.SessionID(sessionID))
	if err != nil || !existSession.Equal(sessionID) {
		s.log.Warn("create wallet failed", withReqID(ctx, "user_id", userID, "error", "session not found")...)
		return nil, errs.ErrSessionNotFound
	}

	return user, nil
}
