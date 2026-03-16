package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"p2p_wallet/internal/handler"
	"p2p_wallet/internal/infra/persistence/postgres"
	"p2p_wallet/internal/infra/server/httpserver"
	"p2p_wallet/internal/repository"
	"p2p_wallet/internal/service"
	"p2p_wallet/internal/shared/config"
	"p2p_wallet/internal/shared/logger"
	"p2p_wallet/internal/shared/tracing"
)

// Example: VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo unknown) COMMIT_SHA=$(git rev-parse --short HEAD) BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ) docker compose build
var (
	VERSION    = "dev"
	COMMIT_SHA = "unknown" //nolint:revive
	BUILD_TIME = "unknown" //nolint:revive
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	log, err := logger.New(cfg.Logger)
	if err != nil {
		panic(err)
	}

	log = log.With("version", VERSION).With("env", cfg.Environment)

	log.Info(
		"build info",
		"commit_sha", COMMIT_SHA,
		"build_time", BUILD_TIME,
	)

	traceShutdown, err := tracing.Init(context.Background(), cfg.Tracing, "p2p-wallet", VERSION, cfg.Environment)
	if err != nil {
		panic(err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = traceShutdown(shutdownCtx)
	}()

	postgresClient, err := postgres.NewClient(cfg.Postgres)
	if err != nil {
		panic(err)
	}
	defer postgresClient.Close() //nolint:errcheck

	var userRepo service.UserRepo = repository.NewUserPostgresRepo(postgresClient)
	userRepo = repository.NewUserRepoWithTracing(userRepo)
	userRepo = repository.NewUserRepoWithMetrics(userRepo)

	var walletRepo service.WalletRepo = repository.NewWalletPostgresRepo(postgresClient)
	walletRepo = repository.NewWalletRepoWithCache(walletRepo)
	walletRepo = repository.NewWalletRepoWithTracing(walletRepo)
	walletRepo = repository.NewWalletRepoWithMetrics(walletRepo)

	sessionRepo := repository.NewSessionRepo()
	userSrv := service.NewAuthService(log, userRepo, sessionRepo)
	walletSrv := service.NewWalletService(log, userRepo, walletRepo, sessionRepo)
	h := handler.New(userSrv, walletSrv)
	srv := httpserver.NewServer(cfg.ServerConfig, log, h, postgresClient)

	go func() {
		if err = srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server failed", "error", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Info("http server is stopping...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	srv.Close(ctx)
}
