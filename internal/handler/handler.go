//nolint:revive
package handler

import (
	"p2p_wallet/internal/api"
)

type handler struct {
	userSrv   AuthService
	walletSrv WalletService
}

var _ api.StrictServerInterface = (*handler)(nil)

func New(
	userSrv AuthService,
	walletSrv WalletService,
) *handler {
	return &handler{
		userSrv:   userSrv,
		walletSrv: walletSrv,
	}
}
