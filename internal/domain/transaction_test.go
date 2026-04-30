package domain

import (
	"testing"
	"time"

	"p2p_wallet/internal/errs"

	"github.com/stretchr/testify/assert"
)

func TestTransactionStatusString(t *testing.T) {
	assert.Equal(t, "new", StatusNew.String())
	assert.Equal(t, "succeed", StatusSucceed.String())
	assert.Equal(t, "failed", StatusFailed.String())
}

func TestNewTransaction(t *testing.T) {
	t.Run("transaction created", func(t *testing.T) {
		fromWallet := &Wallet{
			ID:          WalletID(1),
			Currency:    CurrencyUSD,
			TotalAmount: 200,
			HeldAmount:  25,
		}
		toWallet := &Wallet{
			ID:       WalletID(2),
			Currency: CurrencyUSD,
		}

		startedAt := time.Now()
		transaction, err := NewTransaction("idemp-1", 100, fromWallet, toWallet)
		finishedAt := time.Now()

		assert.Nil(t, err)
		assert.NotNil(t, transaction)
		assert.Equal(t, "idemp-1", transaction.IdempotencyKey)
		assert.Equal(t, StatusNew, transaction.Status)
		assert.Len(t, transaction.Entries, 2)
		assert.WithinRange(t, transaction.CreatedAt, startedAt, finishedAt)

		assert.Equal(t, WalletID(1), transaction.Entries[0].WalletID)
		assert.Equal(t, int64(-100), transaction.Entries[0].Money.Amount())
		assert.Equal(t, "USD", transaction.Entries[0].Money.Currency())

		assert.Equal(t, WalletID(2), transaction.Entries[1].WalletID)
		assert.Equal(t, int64(100), transaction.Entries[1].Money.Amount())
		assert.Equal(t, "USD", transaction.Entries[1].Money.Currency())
	})

	t.Run("amount is not positive", func(t *testing.T) {
		fromWallet := &Wallet{
			ID:          WalletID(1),
			Currency:    CurrencyUSD,
			TotalAmount: 200,
		}
		toWallet := &Wallet{
			ID:       WalletID(2),
			Currency: CurrencyUSD,
		}

		transaction, err := NewTransaction("idemp-1", 0, fromWallet, toWallet)

		assert.Nil(t, transaction)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "failed to create transaction")
		assert.ErrorIs(t, err, errs.ErrNotPositiveAmount)
	})

	t.Run("not enough available money", func(t *testing.T) {
		fromWallet := &Wallet{
			ID:          WalletID(1),
			Currency:    CurrencyUSD,
			TotalAmount: 100,
			HeldAmount:  50,
		}
		toWallet := &Wallet{
			ID:       WalletID(2),
			Currency: CurrencyUSD,
		}

		transaction, err := NewTransaction("idemp-1", 60, fromWallet, toWallet)

		assert.Nil(t, transaction)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrNotEnoughMoney)
	})

	t.Run("wallet currencies mismatch", func(t *testing.T) {
		fromWallet := &Wallet{
			ID:          WalletID(1),
			Currency:    CurrencyUSD,
			TotalAmount: 200,
		}
		toWallet := &Wallet{
			ID:       WalletID(2),
			Currency: CurrencyEUR,
		}

		transaction, err := NewTransaction("idemp-1", 100, fromWallet, toWallet)

		assert.Nil(t, transaction)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrCurrencyMismatch)
	})
}

func TestValidateAmount(t *testing.T) {
	assert.ErrorIs(t, validateAmount(0), errs.ErrNotPositiveAmount)
	assert.ErrorIs(t, validateAmount(-1), errs.ErrNotPositiveAmount)
	assert.Nil(t, validateAmount(1))
}

func TestValidateWallets(t *testing.T) {
	t.Run("wallets valid", func(t *testing.T) {
		fromWallet := &Wallet{
			Currency:    CurrencyEUR,
			TotalAmount: 200,
			HeldAmount:  50,
		}
		toWallet := &Wallet{
			Currency: CurrencyEUR,
		}

		assert.Nil(t, validateWallets(150, fromWallet, toWallet))
	})

	t.Run("not enough money", func(t *testing.T) {
		fromWallet := &Wallet{
			Currency:    CurrencyEUR,
			TotalAmount: 100,
			HeldAmount:  1,
		}
		toWallet := &Wallet{
			Currency: CurrencyEUR,
		}

		assert.ErrorIs(t, validateWallets(100, fromWallet, toWallet), errs.ErrNotEnoughMoney)
	})

	t.Run("currency mismatch", func(t *testing.T) {
		fromWallet := &Wallet{
			Currency:    CurrencyUSD,
			TotalAmount: 100,
		}
		toWallet := &Wallet{
			Currency: CurrencyEUR,
		}

		assert.ErrorIs(t, validateWallets(50, fromWallet, toWallet), errs.ErrCurrencyMismatch)
	})
}
