package domain

import (
	"time"
)

type Transaction struct {
	ID        int64
	Status    string
	CreatedAt time.Time
	Entries   []Entry
}

type Entry struct {
	ID       int64
	WalletID int64
	Amount   int64
	Currency string
}
