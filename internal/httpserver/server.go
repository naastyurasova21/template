package httpserver

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/naastyurasova21/template/api"
	"github.com/naastyurasova21/template/internal/handler"
)

const (
	readTimeout       = 10 * time.Second
	readHeaderTimeout = 5 * time.Second
	writeTimeout      = 10 * time.Second
	idleTimeout       = 60 * time.Second
)

type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
}

func New(addr string, h *handler.Handler, logger *slog.Logger) *Server {
	router := chi.NewRouter()

	apiHandler := api.HandlerWithOptions(h, api.ChiServerOptions{
		BaseRouter:       router,
		ErrorHandlerFunc: handler.ErrorHandler,
	})

	return &Server{
		logger: logger,
		httpServer: &http.Server{
			Addr:              addr,
			Handler:           apiHandler,
			ReadTimeout:       readTimeout,
			ReadHeaderTimeout: readHeaderTimeout,
			WriteTimeout:      writeTimeout,
			IdleTimeout:       idleTimeout,
		},
	}
}

func (s *Server) Start() error {
	s.logger.Info("http server starting", "addr", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
