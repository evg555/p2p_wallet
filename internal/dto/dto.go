package dto

type LoginInput struct {
	Login    string
	Password string
}

type RegisterInput struct {
	LoginInput
	Name     string
	LastName string
}

type CreateWalletInput struct {
	UserID   int64
	Currency string
}

type TransferBalanceInput struct {
	IdempotencyKey string
	FromWalletID   int64
	ToWalletID     int64
	Amount         int64
}
