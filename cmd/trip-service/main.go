package main

import (
	"context"
	"errors"
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
	if err := run(); err != nil {
		slog.Error("service failed", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger := telemetry.NewLogger(cfg.LogLevel)
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	shutdownOtel, err := telemetry.Init(ctx, cfg)
	if err != nil {
		return err
	}
	defer func() {
		otelCtx, cancel := context.WithTimeout(context.Background(), otelShutdownTimeout)
		defer cancel()
		if err := shutdownOtel(otelCtx); err != nil {
			logger.Error("otel shutdown failed", "err", err)
		}
	}()

	c, err := container.New(ctx, cfg, logger)
	if err != nil {
		return err
	}
	defer c.Pool.Close()

	srvErr := make(chan error, 1)
	go func() {
		logger.Info("http server starting", "addr", c.Server.Addr)
		if err := c.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			srvErr <- err
		}
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case err := <-srvErr:
		return err
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := c.Server.Shutdown(shutdownCtx); err != nil {
		return err
	}

	logger.Info("server stopped")
	return nil
}
