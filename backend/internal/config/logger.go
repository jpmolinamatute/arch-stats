package config

import (
	"log/slog"
	"os"
)

// NewLogger creates a structured logger configured for the given mode.
// Dev mode uses human-readable text output at Debug level.
// Prod mode uses JSON output at Info level to stdout.
func NewLogger(devMode bool) *slog.Logger {
	var handler slog.Handler

	if devMode {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	}

	return slog.New(handler)
}
