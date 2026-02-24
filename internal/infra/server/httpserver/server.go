package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"p2p_wallet/internal/api"
	"p2p_wallet/internal/shared/config"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/go-chi/chi/v5"
	middleware "github.com/oapi-codegen/nethttp-middleware"
)

type server struct {
	addr string
	log  Logger
	srv  *http.Server
}

type Logger interface {
	Info(msg string, keysAndValues ...any)
	Error(msg string, keysAndValues ...any)
	Warn(msg string, keysAndValues ...any)
}

func NewServer(cfg config.ServerConfig, log Logger, h api.ServerInterface) *server {
	router := buildRouter(log, h)
	addr := buildAddress(cfg)

	srv := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
	}

	return &server{
		addr: addr,
		log:  log,
		srv:  srv,
	}
}

func (s *server) Start() error {
	s.log.Info("http server started", "addr", s.addr)
	return s.srv.ListenAndServe()
}

func (s *server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}

func (s *server) Close(ctx context.Context) {
	if err := s.srv.Shutdown(ctx); err != nil {
		s.log.Error("http server shutdown failed", "error", err)
		_ = s.srv.Close()
	}
	s.log.Info("http server stopped")
}

func buildRouter(log Logger, h api.ServerInterface) http.Handler {
	loader := openapi3.NewLoader()
	swagger, err := loader.LoadFromFile("spec/openapi/users.yaml")
	if err != nil {
		log.Error("load openapi", "error", err)
	}
	if err = swagger.Validate(context.Background()); err != nil {
		log.Error("validate openapi", "error", err)
	}

	r := chi.NewRouter()
	r.Use(loggerMiddleware(log))
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

func buildAddress(cfg config.ServerConfig) string {
	return net.JoinHostPort(cfg.Host, cfg.Port)
}
