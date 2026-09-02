package config_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/jpmolinamatute/arch-stats/backend/internal/config"
)

func TestNewLogger_DevMode(t *testing.T) {
	logger := config.NewLogger(true)
	if logger == nil {
		t.Fatal("NewLogger(true) returned nil")
	}
	ctx := context.Background()
	if !logger.Enabled(ctx, slog.LevelDebug) {
		t.Error("dev logger should be enabled at Debug level")
	}
	if !logger.Enabled(ctx, slog.LevelInfo) {
		t.Error("dev logger should be enabled at Info level")
	}
}

func TestNewLogger_ProdMode(t *testing.T) {
	logger := config.NewLogger(false)
	if logger == nil {
		t.Fatal("NewLogger(false) returned nil")
	}
	ctx := context.Background()
	if logger.Enabled(ctx, slog.LevelDebug) {
		t.Error("prod logger should NOT be enabled at Debug level")
	}
	if !logger.Enabled(ctx, slog.LevelInfo) {
		t.Error("prod logger should be enabled at Info level")
	}
}
