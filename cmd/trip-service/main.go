package main

import (
	"context"
	"log/slog"
	"net/http" 
	"os"
	"os/signal"
	"syscall"

	"github.com/naastyurasova21/template/internal/config"
	"github.com/naastyurasova21/template/internal/container"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config load failed", "err", err)
		os.Exit(1)
	}

	logger := newLogger(cfg.LogLevel)
	slog.SetDefault(logger)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

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
	logger.Info("server stopped")
}

func newLogger(level string) *slog.Logger {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "info":
		lvl = slog.LevelInfo
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl}))
}
