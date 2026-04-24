package httpserver

import (
	"context"
	"encoding/json"
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

func NewServer(
	cfg config.ServerConfig,
	log Logger,
	h api.StrictServerInterface,
	checker readinessChecker,
	sessionRepo SessionRepository,
	userRepo UserRepository,
) *server {
	router := buildRouter(log, h, checker, sessionRepo, userRepo)
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

func buildRouter(
	log Logger,
	h api.StrictServerInterface,
	checker readinessChecker,
	sessionRepo SessionRepository,
	userRepo UserRepository,
) http.Handler {
	loader := openapi3.NewLoader()
	swagger, err := loader.LoadFromFile("spec/openapi/p2p-wallet.yaml")
	if err != nil {
		log.Error("load openapi", "error", err)
	}
	if err = swagger.Validate(context.Background()); err != nil {
		log.Error("validate openapi", "error", err)
	}

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

	handlerWrapper := api.ServerInterfaceWrapper{
		Handler: strictHandler,
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		},
	}

	validator := middleware.OapiRequestValidatorWithOptions(swagger, &middleware.Options{
		Options: openapi3filter.Options{
			AuthenticationFunc: authenticateRequest,
		},
		ErrorHandlerWithOpts: func(_ context.Context, err error, w http.ResponseWriter, _ *http.Request, opts middleware.ErrorHandlerOpts) {
			var securityErr *openapi3filter.SecurityRequirementsError
			if errors.As(err, &securityErr) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(api.ErrorResponse{
					Code:    "unauthorized",
					Message: err.Error(),
				})
				return
			}

			http.Error(w, err.Error(), opts.StatusCode)
		},
	})

	metrics := newHTTPMetrics()
	rootRouter := chi.NewRouter()
	registerProbeEndpoints(rootRouter, checker)

	apiRouter := chi.NewRouter()
	apiRouter.Use(RequestIDMiddleware)
	apiRouter.Use(metrics.Middleware)
	apiRouter.Use(AccessLogMiddleware(log))
	apiRouter.Method(http.MethodGet, metricPath, metrics.Handler())

	publicRouter := chi.NewRouter()
	publicRouter.Use(validator)
	publicRouter.Post("/users/login", handlerWrapper.LoginUser)
	publicRouter.Post("/users/register", handlerWrapper.RegisterUser)

	protectedRouter := chi.NewRouter()
	protectedRouter.Use(validator)
	protectedRouter.Use(AuthMiddleware(log, sessionRepo, userRepo))
	protectedRouter.Post("/users/logout", handlerWrapper.LogoutUser)
	protectedRouter.Post("/wallets", handlerWrapper.CreateWallet)
	protectedRouter.Get("/wallets/me", handlerWrapper.ListUserWallets)
	protectedRouter.Post("/balance/transfer", handlerWrapper.TransferBalance)

	apiRouter.Mount("/", publicRouter)
	apiRouter.Mount("/", protectedRouter)
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
