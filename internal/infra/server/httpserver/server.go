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
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
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

type readinessChecker interface {
	Ping(ctx context.Context) error
}

func NewServer(cfg config.ServerConfig, log Logger, h api.StrictServerInterface, checker readinessChecker) *server {
	router := buildRouter(log, h, checker)
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

func buildRouter(log Logger, h api.StrictServerInterface, checker readinessChecker) http.Handler {
	loader := openapi3.NewLoader()
	swagger, err := loader.LoadFromFile("spec/openapi/users.yaml")
	if err != nil {
		log.Error("load openapi", "error", err)
	}
	if err = swagger.Validate(context.Background()); err != nil {
		log.Error("validate openapi", "error", err)
	}

	metrics := newHTTPMetrics()
	rootRouter := chi.NewRouter()
	registerProbeEndpoints(rootRouter, checker)

	apiRouter := chi.NewRouter()
	apiRouter.Use(RequestIDMiddleware)
	apiRouter.Use(metrics.Middleware)
	apiRouter.Use(AccessLogMiddleware(log))
	apiRouter.Use(SessionMiddleware)
	apiRouter.Method(http.MethodGet, metricPath, metrics.Handler())

	openAPIRouter := chi.NewRouter()
	openAPIRouter.Use(middleware.OapiRequestValidatorWithOptions(swagger, &middleware.Options{
		Options: openapi3filter.Options{
			AuthenticationFunc: authenticateRequest,
		},
	}))

	strictHandler := api.NewStrictHandlerWithOptions(h, []api.StrictMiddlewareFunc{
		TracingMiddleware(),
		ErrorLoggingMiddleware(log),
	}, api.StrictHTTPServerOptions{
		RequestErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		},
		ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, "internal server error", http.StatusInternalServerError)
		},
	})

	apiRouter.Mount("/", api.HandlerFromMux(strictHandler, openAPIRouter))
	rootRouter.Mount("/", apiRouter)

	return otelhttp.NewHandler(
		rootRouter,
		"http.request",
		otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
			return r.Method + " " + routePattern(r)
		}),
		otelhttp.WithFilter(func(r *http.Request) bool {
			return !isObservabilityExcludedPath(r.URL.Path)
		}),
	)
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
