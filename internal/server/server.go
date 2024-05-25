package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v2"
	"github.com/hpcsc/chi-api/internal/config"
	"github.com/hpcsc/chi-api/internal/response"
	"github.com/hpcsc/chi-api/internal/usecase"
	"github.com/hpcsc/chi-api/internal/usecase/root"
	"github.com/hpcsc/chi-api/internal/usecase/user"
	nethttpmiddleware "github.com/oapi-codegen/nethttp-middleware"
)

const openAPISchemaPath = "openapi/spec.yaml"

// serverHandler is used purely to ensure combination of all child handlers satisfy overall `ServerInterface`
// .i.e. no route is missing
var _ ServerInterface = (*serverHandler)(nil)

// need to use type alias since all child route handlers have the same name `RouteHandler` (in different packages)
// we cannot embed structs with the same name in `serverHandler` below
type rootHandler = root.RouteHandler
type userHandler = user.RouteHandler

// serverHandler is combination of all child route handlers
type serverHandler struct {
	rootHandler
	userHandler
}

func New(name string, cfg *config.Config, logger *slog.Logger) (*Server, error) {
	handler, err := newServerHandler(name, cfg, logger)
	if err != nil {
		return nil, err
	}

	return &Server{
		cfg: cfg,
		httpServer: &http.Server{
			Addr:    fmt.Sprintf(":%s", cfg.Port),
			Handler: handler,
		},
		logger: logger,
	}, nil
}

func newServerHandler(name string, cfg *config.Config, logger *slog.Logger) (http.Handler, error) {
	r := chi.NewRouter()

	r.Use(httplog.RequestLogger(httplog.NewLogger(name, httplog.Options{
		JSON:           true,
		LogLevel:       slog.LevelInfo,
		Concise:        true,
		RequestHeaders: true,
	})))

	r.Use(middleware.Recoverer)

	apiRouter, err := apiSubRouter(cfg, logger)
	if err != nil {
		return nil, err
	}

	r.Mount("/api", apiRouter)
	r.Mount("/z", docSubRouter())

	return r, nil
}

func docSubRouter() http.Handler {
	router := chi.NewRouter()
	staticFileServer := http.FileServer(http.Dir("openapi"))
	router.Mount("/v3/api-docs", http.StripPrefix("/z/v3/api-docs", staticFileServer))
	return router
}

func apiSubRouter(cfg *config.Config, logger *slog.Logger) (http.Handler, error) {
	openAPISchema, err := openapi3.NewLoader().LoadFromFile(openAPISchemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load OpenAPI schema from %s: %v", openAPISchemaPath, err)
	}

	router := chi.NewRouter()

	router.Use(nethttpmiddleware.OapiRequestValidatorWithOptions(openAPISchema, &nethttpmiddleware.Options{
		ErrorHandler: openapiRequestValidatorErrorHandler,
	}))

	if err := usecase.Register(router, cfg, logger); err != nil {
		return nil, err
	}

	return router, nil
}

func openapiRequestValidatorErrorHandler(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(response.Fail(message))
}

type Server struct {
	cfg        *config.Config
	httpServer *http.Server
	logger     *slog.Logger
}

func (s *Server) Start() {
	s.logger.Info(fmt.Sprintf("listening at %v", s.httpServer.Addr))
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.logger.Error(fmt.Sprintf("failed to start server: %v", err))
	}
}

func (s *Server) Shutdown() {
	// Shutdown signal with grace period of 30 seconds
	withTimeoutCtx, cancelTimeout := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelTimeout()

	go func() {
		<-withTimeoutCtx.Done()
		if errors.Is(withTimeoutCtx.Err(), context.DeadlineExceeded) {
			s.logger.Error("graceful shutdown timed out")
		}
	}()

	s.httpServer.SetKeepAlivesEnabled(false)
	if err := s.httpServer.Shutdown(withTimeoutCtx); err != nil {
		s.logger.Error(fmt.Sprintf("failed to gracefully shutdown server: %v", err))
	} else {
		s.logger.Info("server shutdown")
	}
}
