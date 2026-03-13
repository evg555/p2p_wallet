package repository

import (
	"fmt"

	"p2p_wallet/internal/domain"
)

const (
	statusActive  = "active"
	statusBlocked = "blocked"
)

func newWalletStatus(status string) (domain.WalletStatus, error) {
	switch status {
	case statusActive:
		return domain.StatusActive, nil
	case statusBlocked:
		return domain.StatusBlocked, nil
	default:
		return 0, fmt.Errorf("unknown wallet status %s", status)
	}
}
