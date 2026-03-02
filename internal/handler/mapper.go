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
