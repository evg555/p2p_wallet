package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	middleware "github.com/oapi-codegen/nethttp-middleware"

	"p2p_wallet/internal/api"
	"p2p_wallet/internal/handler"
	"p2p_wallet/internal/repository"
	"p2p_wallet/internal/service"
)

var srvAddress = "localhost:8080"

func main() {
	repo := repository.New()
	userSrv := service.New(repo)
	h := handler.New(userSrv)
	router := buildRouter(h)

	srv := &http.Server{
		Addr:              srvAddress,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("http server started on :%s", srvAddress)
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
	r.Use(middleware.OapiRequestValidator(swagger))

	return api.HandlerFromMux(h, r)
}
