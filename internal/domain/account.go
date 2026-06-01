package domain

import (
	"time"
)

type AccountID int64

type AccountType string

const (
	TypeAvailable   AccountType = "available"
	TypeHeld        AccountType = "held"
	TypeExternalIn  AccountType = "external_in"
	TypeExternalOut AccountType = "external_out"
	TypeFees        AccountType = "fees"
)

type Account struct {
	ID        AccountID
	Type      AccountType
	Currency  Currency
	Status    WalletStatus
	CreatedAt time.Time
	UpdatedAt *time.Time
}
