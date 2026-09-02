# Logging Setup Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Configure structured logging using Go's standard library `log/slog` with environment-dependent formatting (text at Debug level for development, JSON at Info level for production) and wire it into application startup.

**Architecture:** A factory function `NewLogger(devMode bool) *slog.Logger` in `internal/config` builds an `*slog.Logger` with either `slog.NewTextHandler` or `slog.NewJSONHandler` writing to `os.Stdout`. In `main.go`, the logger is instantiated immediately following configuration loading and registered as the default logger via `slog.SetDefault(logger)`.

**Tech Stack:** Go 1.27+ standard library (`log/slog`, `os`, `context`).

**Spec:** [docs/go_refactor/tasks/006-logging_setup.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/006-logging_setup.md)

## Global Constraints

- Git branch: `refactor/006-logging-setup`
- Package path: `github.com/jpmolinamatute/arch-stats/backend/internal/config`
- Standard library only for logging (`log/slog`); no external logging dependencies.
- In dev mode: uses `slog.NewTextHandler(os.Stdout, ...)` with `slog.LevelDebug`.
- In prod mode: uses `slog.NewJSONHandler(os.Stdout, ...)` with `slog.LevelInfo`.
- All tests must pass via `go test ./internal/config/... -v`.
- `go vet ./...` must report no issues.
- `go build ./cmd/arch-stats` must compile without errors.

---

## File Structure

```
backend/
├── cmd/
│   └── arch-stats/
│       └── main.go                  # Wire config.NewLogger and slog.SetDefault
└── internal/
    └── config/
        ├── config.go                # Existing Config struct with DevMode field
        ├── logger.go                # [NEW] NewLogger factory function
        └── logger_test.go           # [NEW] Unit tests for NewLogger dev/prod modes
```

---

### Task 1: Create Logger Factory and Unit Tests

**Files:**
- Create: `backend/internal/config/logger_test.go`
- Create: `backend/internal/config/logger.go`

**Interfaces:**
- Consumes: `devMode bool`
- Produces: `config.NewLogger(devMode bool) *slog.Logger`

- [x] **Step 1: Write the failing tests**

Create `backend/internal/config/logger_test.go`:

```go
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
```

- [x] **Step 2: Run tests to verify they fail**

Run:
```bash
cd backend && go test ./internal/config/... -run TestNewLogger -v
```
Expected: FAIL with `undefined: config.NewLogger`

- [x] **Step 3: Implement the logger factory**

Create `backend/internal/config/logger.go`:

```go
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
```

- [x] **Step 4: Run tests to verify they pass**

Run:
```bash
cd backend && go test ./internal/config/... -v
```
Expected: PASS for all tests including `TestNewLogger_DevMode` and `TestNewLogger_ProdMode`.

- [x] **Step 5: Commit**

```bash
git add backend/internal/config/logger.go backend/internal/config/logger_test.go
git commit -m "feat(config): add NewLogger slog factory with dev and prod modes"
```

---

### Task 2: Wire Logger into `main.go`

**Files:**
- Modify: `backend/cmd/arch-stats/main.go:24-30`

**Interfaces:**
- Consumes: `config.NewLogger(cfg.DevMode) *slog.Logger`, `slog.SetDefault(*slog.Logger)`
- Produces: Default structured logger initialized before application subsystem startup

- [x] **Step 1: Update `main.go` with logger initialization**

In `backend/cmd/arch-stats/main.go`, after `cfg, err := config.Load()`, initialize the logger, set it as default, and log the startup banner:

```go
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
```

- [x] **Step 2: Run `go vet` and verify build**

Run:
```bash
cd backend && go vet ./... && go build ./cmd/arch-stats
```
Expected: Exit code 0, no vet issues, binary builds cleanly.

- [x] **Step 3: Commit**

```bash
git add backend/cmd/arch-stats/main.go
git commit -m "feat(main): initialize and set default slog logger from config"
```

---

### Task 3: Full Verification & Acceptance Check

**Files:**
- Verify: `backend/internal/config/logger.go`
- Verify: `backend/internal/config/logger_test.go`
- Verify: `backend/cmd/arch-stats/main.go`

- [x] **Step 1: Run entire test suite with race detector**

Run:
```bash
cd backend && go test -race ./... -v
```
Expected: All packages pass.

- [x] **Step 2: Run `go vet` across the repository**

Run:
```bash
cd backend && go vet ./...
```
Expected: Clean output with 0 errors.

- [x] **Step 3: Verify formatting and linting**

Run:
```bash
cd backend && golangci-lint run ./...
```
Expected: No linter errors reported.

- [x] **Step 4: Update task status document**

Update `docs/go_refactor/tasks/006-logging_setup.md` checking off all acceptance criteria and steps.

```bash
git add docs/go_refactor/tasks/006-logging_setup.md
git commit -m "docs: complete task 006 logging setup checklist"
```

All tasks are done.

