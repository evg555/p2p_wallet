package domain

import (
	"fmt"
	"time"

	"p2p_wallet/internal/errs"
)

const (
	TypeTransfer = "transfer"
)

type Transaction struct {
	ID             int64
	IdempotencyKey string
	CreatedAt      time.Time
	Entries        []Entry
}

type Entry struct {
	ID       int64
	WalletID WalletID
	Money    Money
}

func NewTransaction(idempKey string, amount int64, fromWallet *Wallet, toWallet *Wallet) (*Transaction, error) {
	err := validateTransaction(amount, fromWallet, toWallet)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	money := NewMoney(amount, fromWallet.Currency)

	entryFrom := Entry{
		WalletID: fromWallet.ID,
		Money:    money.Invert(),
	}

	entryTo := Entry{
		WalletID: toWallet.ID,
		Money:    money,
	}

	transaction := &Transaction{
		IdempotencyKey: idempKey,
		Entries:        []Entry{entryFrom, entryTo},
		CreatedAt:      time.Now().UTC(),
	}

	return transaction, nil
}

func validateTransaction(amount int64, fromWallet *Wallet, toWallet *Wallet) error {
	err := validateAmount(amount)
	if err != nil {
		return err
	}

	err = validateWallets(fromWallet, toWallet)
	if err != nil {
		return err
	}

	return nil
}

func validateAmount(amount int64) error {
	if amount <= 0 {
		return errs.ErrNotPositiveAmount
	}
	return nil
}

func validateWallets(fromWallet *Wallet, toWallet *Wallet) error {
	if fromWallet.Currency != toWallet.Currency {
		return errs.ErrCurrencyMismatch
	}

	return nil
}
