package handler

import (
	"p2p_wallet/internal/api"
	"p2p_wallet/internal/dto"
)

func reqToLoginDTO(req *api.LoginRequest) dto.LoginInput {
	return dto.LoginInput{
		Login:    req.Login,
		Password: req.Password,
	}
}

func reqToRegisterDTO(req *api.RegisterRequest) dto.RegisterInput {
	return dto.RegisterInput{
		LoginInput: dto.LoginInput{
			Login:    req.Login,
			Password: req.Password,
		},
		Name:     req.Name,
		LastName: req.LastName,
	}
}

func reqToCreateWalletDTO(req *api.CreateWalletRequest) dto.CreateWalletInput {
	return dto.CreateWalletInput{
		Currency: string(req.Currency),
	}
}

func reqToTransferBalanceDTO(params api.TransferBalanceParams, req *api.BalanceTransferRequest) dto.TransferBalanceInput {
	return dto.TransferBalanceInput{
		IdempotencyKey: params.XIdempotencyKey,
		FromWalletID:   req.FromWalletId,
		ToWalletID:     req.ToWalletId,
		Amount:         req.Amount,
	}
}
