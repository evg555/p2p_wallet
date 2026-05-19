package domain

import (
	"fmt"
	"time"
)

// const precision = 2

type WalletID int64

type WalletStatus int64

func (w WalletID) Int64() int64 {
	return int64(w)
}

func (s WalletStatus) String() string {
	statuses := []string{"active", "blocked"}
	return statuses[s]
}

const (
	StatusActive WalletStatus = iota
	StatusBlocked
)

type Wallet struct {
	ID          WalletID
	Currency    Currency
	Status      WalletStatus
	UserID      UserID
	TotalAmount int64
	HeldAmount  int64
	CreatedAt   time.Time
	UpdatedAt   *time.Time
}

func NewWallet(userID UserID, currency string) (*Wallet, error) {
	newCurrency, err := NewCurrency(currency)
	if err != nil {
		return nil, fmt.Errorf("wallet: %w", err)
	}

	return &Wallet{
		UserID:    userID,
		Currency:  newCurrency,
		Status:    StatusActive,
		CreatedAt: time.Now().UTC(),
	}, nil
}

func NewCurrency(currency string) (Currency, error) {
	switch currency {
	case "EUR":
		return CurrencyEUR, nil
	case "USD":
		return CurrencyUSD, nil
	default:
		return 0, fmt.Errorf("unknown currency %s", currency)
	}
}
