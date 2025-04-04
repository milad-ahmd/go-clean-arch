package http

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/milad-ahmd/go-clean-arch/pkg/config"
	"github.com/milad-ahmd/go-clean-arch/pkg/logger"
	"github.com/milad-ahmd/go-clean-arch/pkg/middleware"
)

// Server represents the HTTP server
type Server struct {
	server *http.Server
	router *mux.Router
	logger logger.Logger
}

// NewServer creates a new HTTP server
func NewServer(cfg *config.Config, logger logger.Logger) *Server {
	router := mux.NewRouter()

	// Create the server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	return &Server{
		server: server,
		router: router,
		logger: logger,
	}
}

// Router returns the router
func (s *Server) Router() *mux.Router {
	return s.router
}

// SetupMiddleware sets up the middleware
func (s *Server) SetupMiddleware() {
	// Apply middleware to all routes
	s.router.Use(func(next http.Handler) http.Handler {
		return middleware.Logger(s.logger)(next)
	})
	s.router.Use(func(next http.Handler) http.Handler {
		return middleware.CORS()(next)
	})
	s.router.Use(func(next http.Handler) http.Handler {
		return middleware.Recover(s.logger)(next)
	})
}

// Start starts the server
func (s *Server) Start() error {
	s.logger.Info("Starting HTTP server on " + s.server.Addr)
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down HTTP server")
	return s.server.Shutdown(ctx)
}
