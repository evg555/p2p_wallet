package postgres

import (
	"context"
	"fmt"
	"time"

	"p2p_wallet/internal/shared/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewClient(cfg config.PostgresConfig) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, buildDSN(cfg))
	if err != nil {
		return nil, fmt.Errorf("open postgres connection: %w", err)
	}

	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return pool, nil
}

func buildDSN(cfg config.PostgresConfig) string {
	// dsn = "postgres://dbuser:dbpass@localhost:5432/p2p_wallet?sslmode=disable"
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database, cfg.SSLMode,
	)
}
