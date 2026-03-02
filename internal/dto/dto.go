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
