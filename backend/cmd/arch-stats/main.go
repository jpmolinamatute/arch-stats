package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/jpmolinamatute/arch-stats/backend/internal/config"
	"github.com/jpmolinamatute/arch-stats/backend/internal/repository"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		return err
	}

	if err := cfg.Validate(); err != nil {
		slog.Error("invalid configuration", "error", err)
		return err
	}

	dsn, err := cfg.DatabaseURL()
	if err != nil {
		slog.Error("failed to construct database URL", "error", err)
		return err
	}

	slog.Info("connecting to database...")
	pool, err := repository.NewPool(ctx, dsn, int32(cfg.PostgresPoolMinSize), int32(cfg.PostgresPoolMaxSize))
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		return err
	}
	defer pool.Close()

	slog.Info("database connection pool initialized",
		"min_conns", cfg.PostgresPoolMinSize,
		"max_conns", cfg.PostgresPoolMaxSize,
	)

	slog.Info("arch-stats server running, waiting for shutdown signal...")
	<-ctx.Done()
	slog.Info("shutting down gracefully...")
	return nil
}
