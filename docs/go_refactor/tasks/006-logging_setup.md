# Task 006: Configure `log/slog` Structured Logging

## Git Branch

`refactor/006-logging-setup`

## Objective

Set up structured logging using Go's standard library `log/slog`. Create a logger factory
function that configures JSON or text output based on dev/prod mode. Wire the logger into
`main.go` via constructor injection so all layers receive it.

## Dependencies

- Task 003 (config package with `DevMode` field)

## Acceptance Criteria

- [ ] `backend/internal/config/logger.go` provides a `NewLogger(devMode bool) *slog.Logger`
  function that:
    - In dev mode: uses `slog.NewTextHandler` with `LevelDebug`
    - In prod mode: uses `slog.NewJSONHandler` with `LevelInfo` writing to `os.Stdout`
- [ ] Unit tests verify both dev and prod logger configurations.
- [ ] `main.go` is updated to create the logger from config and set it as the default.
- [ ] `go test ./internal/config/...` passes.
- [ ] `go vet ./...` reports no issues.

## Files to Create/Modify

| Action | Path |
| ------ | ---- |
| Create | `backend/internal/config/logger.go` |
| Create | `backend/internal/config/logger_test.go` |
| Modify | `backend/cmd/arch-stats/main.go` |

## Steps

- [ ] **Step 1: Write the failing tests**

  Create `backend/internal/config/logger_test.go`:

  ```go
  package config_test

  import (
      "testing"

      "github.com/jpmolinamatute/arch-stats/backend/internal/config"
  )

  func TestNewLogger_DevMode(t *testing.T) {
      logger := config.NewLogger(true)
      if logger == nil {
          t.Fatal("NewLogger(true) returned nil")
      }
      if !logger.Enabled(nil, slog.LevelDebug) {
          t.Error("dev logger should be enabled at Debug level")
      }
  }

  func TestNewLogger_ProdMode(t *testing.T) {
      logger := config.NewLogger(false)
      if logger == nil {
          t.Fatal("NewLogger(false) returned nil")
      }
      if logger.Enabled(nil, slog.LevelDebug) {
          t.Error("prod logger should NOT be enabled at Debug level")
      }
      if !logger.Enabled(nil, slog.LevelInfo) {
          t.Error("prod logger should be enabled at Info level")
      }
  }
  ```

- [ ] **Step 2: Run tests to verify they fail**

  ```bash
  cd backend
  go test ./internal/config/... -v
  ```

- [ ] **Step 3: Implement the logger factory**

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

- [ ] **Step 4: Run tests to verify they pass**

  ```bash
  cd backend
  go test ./internal/config/... -v
  ```

- [ ] **Step 5: Update `main.go`**

  Add logger initialization after config loading:

  ```go
  logger := config.NewLogger(cfg.DevMode)
  slog.SetDefault(logger)
  logger.Info("arch-stats starting", "dev_mode", cfg.DevMode)
  ```

- [ ] **Step 6: Run go vet and build**

  ```bash
  cd backend
  go vet ./...
  go build ./cmd/arch-stats
  ```

- [ ] **Step 7: Commit**

  ```bash
  git add -A
  git commit -m "feat: add structured logging with slog, dev/prod modes"
  ```

## Verification

- `cd backend && go test ./internal/config/... -v` — all tests pass.
- `cd backend && go vet ./...` — clean.
- `cd backend && go build ./cmd/arch-stats` — compiles.
