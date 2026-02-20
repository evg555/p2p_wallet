package service

type Repo interface {
	Register()
	Login()
	Logout()
	GetUser()
}

type service struct {
	repo Repo
}

func New(repo Repo) *service {
	return &service{repo: repo}
}

func (s *service) Register() {
	panic("implement me")
}

func (s *service) Login() {
	panic("implement me")
}
func (s *service) Logout() {
	panic("implement me")
}

func (s *service) GetUser() {
	panic("implement me")
}
