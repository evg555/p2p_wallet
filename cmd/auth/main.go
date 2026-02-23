package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"p2p_wallet/internal/api"
	"p2p_wallet/internal/handler"
	"p2p_wallet/internal/repository"
	"p2p_wallet/internal/service"
	"p2p_wallet/internal/shared/config"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/go-chi/chi/v5"
	middleware "github.com/oapi-codegen/nethttp-middleware"
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
		log.Fatal(err)
	}

	log.Printf("build info: VERSION=%s COMMIT_SHA=%s BUILD_TIME=%s", VERSION, COMMIT_SHA, BUILD_TIME)

	userRepo := repository.NewUserPostgresRepo(cfg.Postgres)
	sessionRepo := repository.NewSessionRepo()
	userSrv := service.New(userRepo, sessionRepo)
	h := handler.New(userSrv)
	router := buildRouter(h)

	srv := &http.Server{
		Addr:              srvAddress(cfg.ServerConfig),
		Handler:           router,
		ReadHeaderTimeout: cfg.ServerConfig.ReadHeaderTimeout,
		ReadTimeout:       cfg.ServerConfig.ReadTimeout,
	}

	go func() {
		log.Printf("http server started on %s", srvAddress(cfg.ServerConfig))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http server failed: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Printf("http server is stopping...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
		_ = srv.Close()
	}

	log.Printf("http server stopped")
}

func buildRouter(h api.ServerInterface) http.Handler {
	loader := openapi3.NewLoader()
	swagger, err := loader.LoadFromFile("spec/openapi/users.yaml")
	if err != nil {
		log.Fatalf("load openapi: %v", err)
	}
	if err = swagger.Validate(context.Background()); err != nil {
		log.Fatalf("validate openapi: %v", err)
	}

	r := chi.NewRouter()
	r.Use(middleware.OapiRequestValidatorWithOptions(swagger, &middleware.Options{
		Options: openapi3filter.Options{
			AuthenticationFunc: authenticateRequest,
		},
	}))

	return api.HandlerFromMux(h, r)
}

func authenticateRequest(_ context.Context, ai *openapi3filter.AuthenticationInput) error {
	if ai == nil || ai.RequestValidationInput == nil || ai.RequestValidationInput.Request == nil {
		return errors.New("invalid authentication input")
	}

	switch ai.SecuritySchemeName {
	case "SessionCookieAuth":
		cookie, err := ai.RequestValidationInput.Request.Cookie("session_id")
		if err != nil || cookie == nil || cookie.Value == "" {
			return errors.New("missing session_id cookie")
		}
		return nil
	default:
		return fmt.Errorf("unsupported security scheme: %s", ai.SecuritySchemeName)
	}
}

func srvAddress(cfg config.ServerConfig) string {
	return net.JoinHostPort(cfg.Host, cfg.Port)
}
