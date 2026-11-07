package main

import (
	"context"
	"log/slog"
	"net/http" 
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/naastyurasova21/template/internal/config"
	"github.com/naastyurasova21/template/internal/container"
	"github.com/naastyurasova21/template/internal/infra/telemetry"
)
const otelShutdownTimeout = 5 * time.Second
func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config load failed", "err", err)
		os.Exit(1)
	}

	logger := telemetry.NewLogger(cfg.LogLevel)
	slog.SetDefault(logger)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	shutdownOtel, err := telemetry.Init(ctx, cfg)
	if err != nil {
		logger.Error("otel init failed", "err", err)
		os.Exit(1)
	}
	c, err := container.New(ctx, cfg, logger)
	if err != nil {
		logger.Error("container init failed", "err", err)
		os.Exit(1)
	}
	defer c.Pool.Close()

	srvErr := make(chan error, 1)
	go func() {
		logger.Info("http server starting", "addr", c.Server.Addr)
		if err := c.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			srvErr <- err
		}
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case err := <-srvErr:
		logger.Error("http server failed", "err", err)
		os.Exit(1)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := c.Server.Shutdown(shutdownCtx); err != nil {
		logger.Error("http server shutdown failed", "err", err)
		os.Exit(1)
	}
    otelCtx, otelCancel := context.WithTimeout(context.Background(), otelShutdownTimeout)
	defer otelCancel()
	if err := shutdownOtel(otelCtx); err != nil {
		logger.Error("otel shutdown failed", "err", err)
	}

	logger.Info("server stopped")
}
