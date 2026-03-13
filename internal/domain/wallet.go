package domain

import (
	"time"
)

// const precision = 2

type WalletID int64

type Wallet struct {
	ID          WalletID
	Currency    string
	Status      string
	UserID      UserID
	TotalAmount int64
	HeldAmount  int64
	CreatedAt   time.Time
	UpdatedAt   *time.Time
}
