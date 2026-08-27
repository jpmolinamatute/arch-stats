# Task 003: Create `internal/config/` with envconfig

## Git Branch

`refactor/003-config-package`

## Objective

Create the `internal/config/` package that loads all application configuration from environment
variables using `kelseyhightower/envconfig`. The config struct must replicate all settings from
the Python `backend-old/src/core/settings.py`, including the socket-vs-TCP detection logic and
cross-field validation.

## Dependencies

- Task 001 (Go module scaffold exists)

## Acceptance Criteria

- [ ] `backend/internal/config/config.go` defines a `Config` struct with `envconfig` tags covering:
    - PostgreSQL connection: user, db, port, password, host, socket_dir, pool min/max, timeouts
    - App runtime: dev_mode, server_port, ws_channel, apply_migrations_on_start
    - Session: ttl_hours, token_bytes
    - Auth: google_oauth_client_id, jwt_secret, jwt_algorithm, jwt_ttl_minutes
- [ ] A `Load() (*Config, error)` function loads config from environment variables with the
  prefix `ARCH_STATS` (or appropriate per field).
- [ ] A `DSNHost() string` method implements socket-vs-TCP detection matching the Python
  `postgres_dsn_host` computed field.
- [ ] A `Validate() error` method performs cross-field validation (production requires socket dir).
- [ ] A `DatabaseURL() string` method constructs the full PostgreSQL connection string.
- [ ] Unit tests in `backend/internal/config/config_test.go` cover:
    - Loading config from env vars
    - Socket-vs-TCP fallback logic
    - Production mode validation (missing socket dir → error)
    - Dev mode allows TCP fallback
- [ ] `go test ./internal/config/...` passes.
- [ ] `go vet ./...` reports no issues.

## Files to Create/Modify

| Action | Path |
| ------ | ---- |
| Create | `backend/internal/config/config.go` |
| Create | `backend/internal/config/config_test.go` |
| Delete | `backend/internal/config/.gitkeep` |
| Modify | `backend/go.mod` (add envconfig dependency) |

## Reference

The Python settings to replicate are in
[settings.py](file:///home/juanpa/Projects/arch-stats/backend-old/src/core/settings.py).

Key fields to port:

| Python Field | Go Field | Env Var |
| ------------ | -------- | ------- |
| `postgres_user` | `PostgresUser` | `POSTGRES_USER` |
| `postgres_db` | `PostgresDB` | `POSTGRES_DB` |
| `postgres_port` | `PostgresPort` | `POSTGRES_PORT` |
| `postgres_password` | `PostgresPassword` | `POSTGRES_PASSWORD` |
| `postgres_host` | `PostgresHost` | `POSTGRES_HOST` |
| `postgres_socket_dir` | `PostgresSocketDir` | `POSTGRES_SOCKET_DIR` |
| `postgres_pool_min_size` | `PostgresPoolMinSize` | `POSTGRES_POOL_MIN_SIZE` |
| `postgres_pool_max_size` | `PostgresPoolMaxSize` | `POSTGRES_POOL_MAX_SIZE` |
| `arch_stats_dev_mode` | `DevMode` | `ARCH_STATS_DEV_MODE` |
| `arch_stats_server_port` | `ServerPort` | `ARCH_STATS_SERVER_PORT` |
| `arch_stats_ws_channel` | `WSChannel` | `ARCH_STATS_WS_CHANNEL` |
| `arch_stats_google_oauth_client_id` | `GoogleOAuthClientID` | `ARCH_STATS_GOOGLE_OAUTH_CLIENT_ID` |
| `arch_stats_jwt_secret` | `JWTSecret` | `ARCH_STATS_JWT_SECRET` |
| `arch_stats_jwt_algorithm` | `JWTAlgorithm` | `ARCH_STATS_JWT_ALGORITHM` |
| `arch_stats_jwt_ttl_minutes` | `JWTTTLMinutes` | `ARCH_STATS_JWT_TTL_MINUTES` |
| `session_ttl_hours` | `SessionTTLHours` | `SESSION_TTL_HOURS` |
| `session_token_bytes` | `SessionTokenBytes` | `SESSION_TOKEN_BYTES` |
| `apply_db_migrations_on_start` | `ApplyMigrationsOnStart` | `APPLY_DB_MIGRATIONS_ON_START` |

## Steps

- [ ] **Step 1: Add the envconfig dependency**

  ```bash
  cd backend
  go get github.com/kelseyhightower/envconfig
  ```

- [ ] **Step 2: Write the failing tests**

  Create `backend/internal/config/config_test.go`:

  ```go
  package config_test

  import (
      "os"
      "testing"

      "github.com/jpmolinamatute/arch-stats/backend/internal/config"
  )

  func setRequiredEnv(t *testing.T) {
      t.Helper()
      t.Setenv("POSTGRES_USER", "testuser")
      t.Setenv("POSTGRES_DB", "testdb")
      t.Setenv("ARCH_STATS_GOOGLE_OAUTH_CLIENT_ID", "test-client-id")
      t.Setenv("ARCH_STATS_JWT_SECRET", "test-secret-key-minimum-length")
      t.Setenv("ARCH_STATS_DEV_MODE", "true")
  }

  func TestLoadConfig(t *testing.T) {
      setRequiredEnv(t)
      t.Setenv("POSTGRES_PORT", "5433")
      t.Setenv("ARCH_STATS_SERVER_PORT", "9000")

      cfg, err := config.Load()
      if err != nil {
          t.Fatalf("Load() error: %v", err)
      }

      if cfg.PostgresUser != "testuser" {
          t.Errorf("PostgresUser = %q, want %q", cfg.PostgresUser, "testuser")
      }
      if cfg.PostgresPort != 5433 {
          t.Errorf("PostgresPort = %d, want %d", cfg.PostgresPort, 5433)
      }
      if cfg.ServerPort != 9000 {
          t.Errorf("ServerPort = %d, want %d", cfg.ServerPort, 9000)
      }
  }

  func TestDefaults(t *testing.T) {
      setRequiredEnv(t)

      cfg, err := config.Load()
      if err != nil {
          t.Fatalf("Load() error: %v", err)
      }

      if cfg.PostgresPort != 5432 {
          t.Errorf("default PostgresPort = %d, want 5432", cfg.PostgresPort)
      }
      if cfg.PostgresPoolMaxSize != 10 {
          t.Errorf("default PostgresPoolMaxSize = %d, want 10", cfg.PostgresPoolMaxSize)
      }
      if cfg.JWTAlgorithm != "HS256" {
          t.Errorf("default JWTAlgorithm = %q, want HS256", cfg.JWTAlgorithm)
      }
      if cfg.SessionTTLHours != 24 {
          t.Errorf("default SessionTTLHours = %d, want 24", cfg.SessionTTLHours)
      }
  }

  func TestDSNHost_SocketFallbackToTCP(t *testing.T) {
      setRequiredEnv(t)
      t.Setenv("POSTGRES_HOST", "db.example.com")
      // No socket dir set → should fall back to TCP host.

      cfg, err := config.Load()
      if err != nil {
          t.Fatalf("Load() error: %v", err)
      }

      host, err := cfg.DSNHost()
      if err != nil {
          t.Fatalf("DSNHost() error: %v", err)
      }
      if host != "db.example.com" {
          t.Errorf("DSNHost() = %q, want %q", host, "db.example.com")
      }
  }

  func TestDSNHost_NeitherSetReturnsError(t *testing.T) {
      setRequiredEnv(t)
      // Neither POSTGRES_HOST nor POSTGRES_SOCKET_DIR set.
      os.Unsetenv("POSTGRES_HOST")
      os.Unsetenv("POSTGRES_SOCKET_DIR")

      cfg, err := config.Load()
      if err != nil {
          t.Fatalf("Load() error: %v", err)
      }

      _, err = cfg.DSNHost()
      if err == nil {
          t.Error("DSNHost() should return error when neither host nor socket is set")
      }
  }

  func TestValidate_ProductionRequiresSocketDir(t *testing.T) {
      setRequiredEnv(t)
      t.Setenv("ARCH_STATS_DEV_MODE", "false")
      // No socket dir set in production → validation should fail.

      cfg, err := config.Load()
      if err != nil {
          t.Fatalf("Load() error: %v", err)
      }

      err = cfg.Validate()
      if err == nil {
          t.Error("Validate() should fail in production without POSTGRES_SOCKET_DIR")
      }
  }

  func TestValidate_DevModeAllowsTCP(t *testing.T) {
      setRequiredEnv(t)
      t.Setenv("ARCH_STATS_DEV_MODE", "true")
      t.Setenv("POSTGRES_HOST", "localhost")

      cfg, err := config.Load()
      if err != nil {
          t.Fatalf("Load() error: %v", err)
      }

      err = cfg.Validate()
      if err != nil {
          t.Errorf("Validate() should pass in dev mode with TCP host: %v", err)
      }
  }
  ```

- [ ] **Step 3: Run tests to verify they fail**

  ```bash
  cd backend
  go test ./internal/config/... -v
  ```

  Expected: compilation errors.

- [ ] **Step 4: Implement the config package**

  Create `backend/internal/config/config.go` with:
    - `Config` struct with `envconfig` tags
    - `Load() (*Config, error)` function
    - `DSNHost() (string, error)` method with socket-vs-TCP logic
    - `Validate() error` method with production mode checks
    - `DatabaseURL() (string, error)` method constructing pgx-compatible DSN

- [ ] **Step 5: Run tests to verify they pass**

  ```bash
  cd backend
  go test ./internal/config/... -v
  ```

  Expected: all tests PASS.

- [ ] **Step 6: Run go vet**

  ```bash
  cd backend
  go vet ./...
  ```

- [ ] **Step 7: Commit**

  ```bash
  rm -f backend/internal/config/.gitkeep
  git add -A
  git commit -m "feat: add internal/config with envconfig, socket/TCP detection, and validation"
  ```

## Verification

- `cd backend && go test ./internal/config/... -v` — all tests pass.
- `cd backend && go vet ./...` — clean.
- `cd backend && go build ./...` — compiles without errors.
