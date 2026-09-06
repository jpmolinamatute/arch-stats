package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jpmolinamatute/arch-stats/backend/internal/auth"
	"github.com/jpmolinamatute/arch-stats/backend/internal/config"
	"github.com/jpmolinamatute/arch-stats/backend/internal/handler"
	"github.com/jpmolinamatute/arch-stats/backend/internal/repository"
	"github.com/jpmolinamatute/arch-stats/backend/internal/service"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 1. Config
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		return err
	}

	// 2. Logger
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

	// 3. Database Connection Pool
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

	// 4. Migrations (Startup auto-migration if configured)
	if cfg.ApplyMigrationsOnStart {
		slog.Info("applying database migrations on startup...")
		if err := repository.RunMigrations(ctx, pool, "migrations"); err != nil {
			slog.Error("startup migration failed", "error", err)
			return err
		}
		slog.Info("startup migrations applied successfully")
	}

	// 5. Repositories
	archerRepo := repository.NewArcherRepo(pool)
	authSessionRepo := repository.NewAuthSessionRepo(pool)
	sessionRepo := repository.NewSessionRepo(pool)
	slotRepo := repository.NewSlotRepo(pool)
	shotRepo := repository.NewShotRepo(pool)
	faceRepo := repository.NewFaceRepo(pool)
	targetRepo := repository.NewTargetRepo(pool)
	maintenanceRepo := repository.NewMaintenanceRepo(pool)
	reportingRepo := repository.NewReportingRepo(pool)
	_ = reportingRepo // Reserved for reporting queries

	// 5b. Log Current Database Schema Version
	if ver, err := maintenanceRepo.GetSchemaVersion(ctx); err == nil {
		slog.Info("database schema version", "version", ver)
	} else {
		slog.Warn("could not read schema version", "error", err)
	}

	// 6. Services
	archerSvc := service.NewArcherService(archerRepo)
	sessionSvc := service.NewSessionService(sessionRepo)
	slotSvc := service.NewSlotService(slotRepo, sessionRepo)
	shotSvc := service.NewShotService(shotRepo, slotRepo)
	faceSvc := service.NewFaceService(faceRepo)
	_ = service.NewTargetService(targetRepo) // Reserved for target operations

	// 7. Auth Service
	authCfg := auth.Config{
		JWTSecret:           cfg.JWTSecret,
		JWTAlgorithm:        cfg.JWTAlgorithm,
		JWTTTLMinutes:       cfg.JWTTTLMinutes,
		SessionTokenBytes:   cfg.SessionTokenBytes,
		GoogleOAuthClientID: cfg.GoogleOAuthClientID,
	}
	authSvc := auth.NewService(archerRepo, authSessionRepo, authCfg)

	// 8. Handlers
	authHandlerCfg := handler.AuthHandlerConfig{
		JWTTTLMinutes: cfg.JWTTTLMinutes,
		DevMode:       cfg.DevMode,
	}
	authHandler := handler.NewAuthHandler(authSvc, archerSvc, authHandlerCfg)
	archerHandler := handler.NewArcherHandler(archerSvc)
	sessionHandler := handler.NewSessionHandler(sessionSvc)
	slotHandler := handler.NewSlotHandler(newSlotServiceAdapter(slotSvc, slotRepo, sessionRepo, targetRepo))
	shotHandler := handler.NewShotHandler(newShotServiceAdapter(shotSvc))
	faceHandler := handler.NewFaceHandler(faceSvc)
	healthHandler := handler.NewHealthHandler(maintenanceRepo)

	// 9. Build Chi Router
	routerDeps := RouterDeps{
		Cfg:            cfg,
		Logger:         logger,
		AuthSvc:        authSvc,
		AuthHandler:    authHandler,
		ArcherHandler:  archerHandler,
		SessionHandler: sessionHandler,
		SlotHandler:    slotHandler,
		ShotHandler:    shotHandler,
		FaceHandler:    faceHandler,
		HealthHandler:  healthHandler,
	}
	router := buildRouter(&routerDeps)

	// 10. HTTP Server with Graceful Shutdown
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.ServerPort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		slog.Info(fmt.Sprintf("arch-stats listening on :%d", cfg.ServerPort), "port", cfg.ServerPort, "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		slog.Info("shutting down gracefully...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			_ = srv.Close()
			return fmt.Errorf("graceful shutdown failed: %w", err)
		}
		slog.Info("server stopped gracefully")
	}

	return nil
}
