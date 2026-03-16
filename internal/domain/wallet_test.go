package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWallet(t *testing.T) {
	t.Run("wallet created", func(t *testing.T) {
		userID := UserID(42)
		currency := "EUR"

		wallet, err := NewWallet(userID, currency)
		assert.Nil(t, err)
		assert.NotNil(t, wallet)
		assert.Equal(t, userID, wallet.UserID)
		assert.Equal(t, CurrencyEUR, wallet.Currency)
		assert.Equal(t, StatusActive, wallet.Status)
		assert.Zero(t, wallet.TotalAmount)
		assert.Zero(t, wallet.HeldAmount)
		assert.False(t, wallet.CreatedAt.IsZero())
		assert.Nil(t, wallet.UpdatedAt)
	})

	t.Run("currency is incorrect", func(t *testing.T) {
		_, err := NewWallet(UserID(42), "RUB")
		assert.Error(t, err)
		assert.ErrorContains(t, err, "wallet: unknown currency RUB")
	})
}

func TestNewCurrency(t *testing.T) {
	t.Run("known currency", func(t *testing.T) {
		currency, err := NewCurrency("USD")
		assert.Nil(t, err)
		assert.Equal(t, CurrencyUSD, currency)
		assert.Equal(t, "USD", currency.String())
	})

	t.Run("unknown currency", func(t *testing.T) {
		_, err := NewCurrency("BTC")
		assert.Error(t, err)
		assert.ErrorContains(t, err, "unknown currency BTC")
	})
}
