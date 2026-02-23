package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"p2p_wallet/internal/domain"
	"p2p_wallet/internal/errs"
	"p2p_wallet/internal/infra/persistence/postgres"
	"p2p_wallet/internal/service"
	"p2p_wallet/internal/shared/config"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgconn"
)

var _ service.UserRepo = (*userPostgresRepo)(nil)

type userPostgresRepo struct {
	client *sql.DB
}

func (r *userPostgresRepo) Save(ctx context.Context, user *domain.User) (*domain.User, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	query, args, err := psql.Insert("users").
		Columns("name", "last_name", "login", "password").
		Values(user.FirstName, user.LastName, user.Login, user.Password).
		Suffix("RETURNING id, name, last_name, login, password, created_at, updated_at").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build save user query: %w", err)
	}

	savedUser := &domain.User{}
	err = r.client.QueryRowContext(
		ctx,
		query,
		args...,
	).Scan(
		&savedUser.ID,
		&savedUser.FirstName,
		&savedUser.LastName,
		&savedUser.Login,
		&savedUser.Password,
		&savedUser.CreatedAt,
		&savedUser.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, errs.ErrUserAlreadyExist
		}

		return nil, fmt.Errorf("save user: %w", err)
	}

	return savedUser, nil
}

func (r *userPostgresRepo) GetByLogin(ctx context.Context, login string) (*domain.User, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	query, args, err := psql.Select("id", "name", "last_name", "login", "password", "created_at", "updated_at").
		From("users").
		Where(sq.Eq{"login": login}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build get user by login query: %w", err)
	}

	user := &domain.User{}
	err = r.client.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Login,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errs.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by login: %w", err)
	}

	return user, nil
}

func (r *userPostgresRepo) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	query, args, err := psql.Select("id", "name", "last_name", "login", "password", "created_at", "updated_at").
		From("users").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build get user by id query: %w", err)
	}

	user := &domain.User{}
	err = r.client.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Login,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errs.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return user, nil
}

func NewUserPostgresRepo(cfg config.PostgresConfig) *userPostgresRepo {
	client, err := postgres.NewClient(buildDSN(cfg))
	if err != nil {
		panic(fmt.Sprintf("failed to initialize postgres client: %v", err))
	}

	return &userPostgresRepo{
		client: client,
	}
}

func buildDSN(cfg config.PostgresConfig) string {
	// dsn = "postgres://dbuser:dbpass@localhost:5432/p2p_wallet?sslmode=disable"
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database, cfg.SSLMode,
	)
}
