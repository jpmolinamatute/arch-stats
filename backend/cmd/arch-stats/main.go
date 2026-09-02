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

	logger := config.NewLogger(cfg.DevMode)
	slog.SetDefault(logger)
	logger.Info("arch-stats starting", "dev_mode", cfg.DevMode)

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

	// Standalone migration CLI subcommand
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		slog.Info("running database migrations...")
		if err := repository.RunMigrations(ctx, pool, "migrations"); err != nil {
			slog.Error("migration failed", "error", err)
			return err
		}
		slog.Info("migrations applied successfully")
		return nil
	}

	// Startup auto-migration if configured
	if cfg.ApplyMigrationsOnStart {
		slog.Info("applying database migrations on startup...")
		if err := repository.RunMigrations(ctx, pool, "migrations"); err != nil {
			slog.Error("startup migration failed", "error", err)
			return err
		}
		slog.Info("startup migrations applied successfully")
	}

	// Log current database schema version
	version, err := repository.GetSchemaVersion(ctx, pool)
	if err != nil {
		slog.Warn("could not read schema version", "error", err)
	} else {
		slog.Info("database schema version", "version", version)
	}

	slog.Info("arch-stats server running, waiting for shutdown signal...")
	<-ctx.Done()
	slog.Info("shutting down gracefully...")
	return nil
}
