package domain

const (
	CurrencyUSD Currency = iota
	CurrencyEUR
)

var currencies = []string{"USD", "EUR"}

type Currency int64

func (c Currency) String() string {
	return currencies[c]
}

type Money struct {
	amount   int64
	currency Currency
}

func NewMoney(amount int64, currency Currency) Money {
	return Money{
		amount:   amount,
		currency: currency,
	}
}

func (m Money) Invert() Money {
	return Money{
		amount:   -m.amount,
		currency: m.currency,
	}
}

func (m Money) Amount() int64 {
	return m.amount
}

func (m Money) Currency() string {
	return m.currency.String()
}
