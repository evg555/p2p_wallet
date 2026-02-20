package repository

import (
	"p2p_wallet/internal/service"
)

var _ service.Repo = (*repo)(nil)

type repo struct{}

func (r repo) Register() {
	//TODO implement me
	panic("implement me")
}

func (r repo) Login() {
	//TODO implement me
	panic("implement me")
}

func (r repo) Logout() {
	//TODO implement me
	panic("implement me")
}

func (r repo) GetUser() {
	//TODO implement me
	panic("implement me")
}

func New() *repo {
	return &repo{}
}
