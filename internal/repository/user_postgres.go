package repository

import (
	"context"
	"database/sql"
	"fmt"

	"p2p_wallet/internal/domain"
	"p2p_wallet/internal/infra/persistence/postgres"
	"p2p_wallet/internal/service"
)

var _ service.UserRepo = (*userPostgresRepo)(nil)

type userPostgresRepo struct {
	client *sql.DB
}

func (r userPostgresRepo) Save(ctx context.Context, user *domain.User) (*domain.User, error) {

	return user, nil
}

func (r userPostgresRepo) GetByLogin(ctx context.Context, login string) (*domain.User, error) {

	return nil, nil
}

func (r userPostgresRepo) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	return nil, nil
}

func NewUserPostgresRepo() *userPostgresRepo {
	client, err := postgres.NewClient()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize postgres client: %v", err))
	}

	return &userPostgresRepo{
		client: client,
	}
}
