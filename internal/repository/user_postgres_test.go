package repository

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"testing"
	"time"

	"p2p_wallet/internal/domain"
	"p2p_wallet/internal/errs"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestUserPostgresRepoSave(t *testing.T) {
	ctx, repo := newPostgresRepoForIntegration(t)

	input, err := domain.NewUser("user-save-it", "secret", "John", "Doe")
	require.NoError(t, err)

	saved, err := repo.Save(ctx, input)
	require.NoError(t, err)
	require.NotNil(t, saved)
	require.Positive(t, saved.ID)
	require.Equal(t, input.Login, saved.Login)
	require.Equal(t, input.Password, saved.Password)
	require.Equal(t, input.FirstName, saved.FirstName)
	require.Equal(t, input.LastName, saved.LastName)
	require.False(t, saved.CreatedAt.IsZero())
	require.Nil(t, saved.UpdatedAt)

	duplicateInput, err := domain.NewUser("user-save-it", "another-secret", "Jane", "Roe")
	require.NoError(t, err)

	duplicateSaved, err := repo.Save(ctx, duplicateInput)
	require.ErrorIs(t, err, errs.ErrUserAlreadyExist)
	require.Nil(t, duplicateSaved)
}

func TestUserPostgresRepoGetByLogin(t *testing.T) {
	t.Run("user exists", func(t *testing.T) {
		ctx, repo := newPostgresRepoForIntegration(t)

		input, err := domain.NewUser("user-login-it", "secret", "John", "Doe")
		require.NoError(t, err)

		saved, err := repo.Save(ctx, input)
		require.NoError(t, err)

		got, err := repo.GetByLogin(ctx, input.Login)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, saved.ID, got.ID)
		require.Equal(t, saved.Login, got.Login)
		require.Equal(t, saved.FirstName, got.FirstName)
		require.Equal(t, saved.LastName, got.LastName)
		require.Equal(t, saved.Password, got.Password)
	})

	t.Run("user does not exist", func(t *testing.T) {
		ctx, repo := newPostgresRepoForIntegration(t)

		got, err := repo.GetByLogin(ctx, "missing-login")
		require.ErrorIs(t, err, errs.ErrUserNotFound)
		require.Nil(t, got)
	})
}

func TestUserPostgresRepoGetByID(t *testing.T) {
	t.Run("user exists", func(t *testing.T) {
		ctx, repo := newPostgresRepoForIntegration(t)

		input, err := domain.NewUser("user-id-it", "secret", "John", "Doe")
		require.NoError(t, err)

		saved, err := repo.Save(ctx, input)
		require.NoError(t, err)

		got, err := repo.GetByID(ctx, int64(saved.ID))
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, saved.ID, got.ID)
		require.Equal(t, saved.Login, got.Login)
		require.Equal(t, saved.FirstName, got.FirstName)
		require.Equal(t, saved.LastName, got.LastName)
		require.Equal(t, saved.Password, got.Password)
	})

	t.Run("user does not exist", func(t *testing.T) {
		ctx, repo := newPostgresRepoForIntegration(t)

		got, err := repo.GetByID(ctx, 9999999999)
		require.ErrorIs(t, err, errs.ErrUserNotFound)
		require.Nil(t, got)
	})
}

func newPostgresRepoForIntegration(t *testing.T) (context.Context, *userPostgresRepo) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	t.Cleanup(cancel)

	pgContainer, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("p2p_wallet"),
		postgres.WithUsername("dbuser"),
		postgres.WithPassword("dbpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(1*time.Minute),
		),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = pgContainer.Terminate(context.Background())
	})

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = db.Close()
	})

	require.NoError(t, db.PingContext(ctx))
	require.NoError(t, applyAllMigrations(t, db))

	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	require.NoError(t, pool.Ping(ctx))

	return ctx, &userPostgresRepo{client: pool}
}

func applyAllMigrations(t *testing.T, db *sql.DB) error {
	t.Helper()

	migrationRoot := migrationRootDir(t)
	dirs, err := migrationDirsWithSQL(migrationRoot)
	if err != nil {
		return err
	}

	if err = goose.SetDialect("postgres"); err != nil {
		return err
	}

	for _, dir := range dirs {
		if err = goose.Up(db, dir); err != nil {
			return err
		}
	}

	return nil
}

func migrationRootDir(t *testing.T) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok)

	return filepath.Join(filepath.Dir(filename), "..", "..", "migration")
}

func migrationDirsWithSQL(root string) ([]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	dirs := make([]string, 0, len(entries)+1)

	rootMatches, err := filepath.Glob(filepath.Join(root, "*.sql"))
	if err != nil {
		return nil, err
	}
	if len(rootMatches) > 0 {
		dirs = append(dirs, root)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		matches, globErr := filepath.Glob(filepath.Join(root, entry.Name(), "*.sql"))
		if globErr != nil {
			return nil, globErr
		}
		if len(matches) > 0 {
			dirs = append(dirs, filepath.Join(root, entry.Name()))
		}
	}

	slices.Sort(dirs)
	return dirs, nil
}
