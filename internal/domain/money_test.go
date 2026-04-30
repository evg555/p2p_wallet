package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMoney(t *testing.T) {
	money := NewMoney(150, CurrencyUSD)

	assert.Equal(t, int64(150), money.Amount())
	assert.Equal(t, "USD", money.Currency())
}

func TestMoneyInvert(t *testing.T) {
	money := NewMoney(150, CurrencyEUR)

	inverted := money.Invert()

	assert.Equal(t, int64(-150), inverted.Amount())
	assert.Equal(t, "EUR", inverted.Currency())
	assert.Equal(t, int64(150), money.Amount())
	assert.Equal(t, "EUR", money.Currency())
}
