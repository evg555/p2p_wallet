package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"p2p_wallet/internal/shared/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewClient(cfg config.PostgresConfig) (*sql.DB, error) {
	db, err := sql.Open("pgx", buildDSN(cfg))
	if err != nil {
		return nil, fmt.Errorf("open postgres connection: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return db, nil
}

func buildDSN(cfg config.PostgresConfig) string {
	// dsn = "postgres://dbuser:dbpass@localhost:5432/p2p_wallet?sslmode=disable"
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database, cfg.SSLMode,
	)
}
