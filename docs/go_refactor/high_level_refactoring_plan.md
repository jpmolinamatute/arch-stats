# High Level Refactoring Plan

We need to refactor the backend written in Python 3.14 into Go 1.27.0. The main reason of this
refactor is to make the backend more performant since it will run on a Raspberry Pi 5. Since there
is NOT production code, this refactor can be done from scratch.

## Repository Strategy

- Rename `backend/` to `backend-old/` to keep as reference during the port.
- Create a new `backend/` directory with the Go module rooted at `backend/` (`go.mod` in `backend/`).
- Once the Go refactor is complete, tested, and agreed upon, delete `backend-old/`.

## Go Project Layout

```text
backend/
├── go.mod
├── go.sum
├── cmd/
│   └── arch-stats/
│       └── main.go          # Entry point, dependency wiring
├── internal/
│   ├── apperror/            # Custom domain error types (ErrNotFound, ErrUnauthorized, etc.)
│   ├── auth/                # Google One Tap verification, JWT, session tokens
│   ├── config/               # Configuration loading via envconfig
│   ├── handler/             # HTTP handlers (one file per domain: archer, session, slot, etc.)
│   ├── middleware/          # Auth middleware, error-to-HTTP mapper, logging, CORS
│   ├── model/               # Domain structs with JSON tags and validation
│   ├── repository/          # Database access layer (pgx + squirrel)
│   ├── service/             # Business logic layer
│   └── websocket/           # WebSocket hub, goroutine-based fan-out
├── migrations/              # SQL migration files (goose format)
├── specs/                   # swaggo/swag generated OpenAPI spec
└── embed.go                 # //go:embed directive for frontend assets
```

## Implementation Order

Bottom-up, building foundation layers first:

1) Config → 2) DB/migrations → 3) Repository layer → 4) Auth → 5) Handlers → 6) WebSocket
→ 7) Frontend serving → 8) CI/CD → 9) Deployment scripts

This plan will consist of these main sections:

## 1. Third-party Dependencies Mapping

Map every Python dependency to its Go equivalent. Prioritize native Go standard library over
third-party packages.

| Python Dependency | Go Equivalent | Notes |
| ----------------- | ------------- | ----- |
| `asyncpg` | **`pgx`** (direct API, not `database/sql`) | Native PostgreSQL types, connection pooling via `pgxpool`, built-in LISTEN/NOTIFY |
| `fastapi` / `starlette` / `uvicorn` | **`chi`** (HTTP router) | Lightweight, stdlib-compatible, clean route grouping for 7 v0 routers |
| `pydantic` / `pydantic-settings` | **`kelseyhightower/envconfig`** (config) + Go structs with JSON tags (validation) | Env-var-to-struct mapping with validation tags |
| `google-auth` | **`google.golang.org/api/idtoken`** | `idtoken.Validate()` = direct equivalent of `verify_oauth2_token()` |
| `pyjwt` | **`golang-jwt/jwt/v5`** | JWT signing and verification |
| `passlib[bcrypt]` | `golang.org/x/crypto/bcrypt` | |
| `httpx` / `requests` | `net/http` client | |
| `uvloop` | Not needed | Go runtime handles concurrency natively |
| OpenAPI generation | **`swaggo/swag`** | Annotation-based, generates spec from Go code comments |
| SQL builder (`sql_statement_builder.py`) | **`squirrel`** | Replaces 389-line custom builder with a well-maintained query builder |

## 2. Authentication System Porting Strategy

The current auth flow is non-trivial (Google One Tap → JWT → SHA-256 hashed session tokens) and is
the highest-risk area to port.

**Library choices:**

- **Google ID token verification**: `google.golang.org/api/idtoken` — `idtoken.Validate(ctx, token,
  clientID)` handles key rotation, caching, and clock skew.
- **JWT signing/verification**: `golang-jwt/jwt/v5` — for minting and verifying the app's own JWTs.
- **Session token generation**: `crypto/rand` (generate) + `crypto/sha256` (hash).

**Requirements:**

- Compatibility with existing session/token behavior so current frontend auth flow remains unchanged.
- Same cookie names, JWT claims structure, and token formats.
- Auth middleware pattern: validate session token → attach user context → reject unauthorized.

## 3. Frontend Integration

The Vue 3 SPA depends on the backend in two ways that must be preserved:

- **Static file serving**: Use `//go:embed` to embed the built frontend into the Go binary, resulting
  in a single self-contained binary for deployment. In dev mode (via `--dev` flag), serve from the
  filesystem instead so the Vite dev server handles the SPA with HMR.
- **OpenAPI spec generation**: Use `swaggo/swag` annotations on Go handlers. Run `swag init` to
  generate `specs/swagger.json`. The frontend's `npm run generate:types` pipeline consumes this spec
  to produce TypeScript types.
- **API contract**: Keep the same endpoints and JSON shapes. Minor improvements (e.g., better error
  responses) are acceptable as long as the frontend types are updated accordingly.

## 4. Python-to-Go Module Mapping

Map the current Python modules to their Go equivalents in the layered architecture:

### `core/` → Split across `internal/` packages

| Python Module | Go Package | Notes |
| ------------- | ---------- | ----- |
| `settings.py` (163 lines) | `internal/config/` | `envconfig` struct with validation |
| `db_pool.py` | `internal/repository/` (pool init) | `pgxpool.New()` in `main.go`, passed via constructor injection |
| `authentication.py` (286 lines) | `internal/auth/` + `internal/middleware/` | Auth logic in `auth/`, HTTP middleware in `middleware/` |
| `logger.py` | stdlib `log/slog` | Configured in `main.go`, passed via constructor injection |
| `base_manager.py` | `internal/repository/` (base) | Common repository interface/patterns |
| `session_manager.py` | `internal/service/` + `internal/repository/` | Business logic in service,DB access in repository |
| `shot_manager.py` | `internal/service/` + `internal/repository/` | Same pattern |
| `slot_manager.py` | `internal/service/` + `internal/repository/` | Same pattern |
| `live_stats_manager.py` | `internal/websocket/` + `internal/service/` | Redesigned with goroutines + channels |
| `face_data.py` (14KB) | `internal/service/` + `internal/repository/` | High-effort item |

### `models/` → `internal/model/` + `internal/repository/`

| Python Module | Go Package | Notes |
| ------------- | ---------- | ----- |
| `parent_model.py` (12KB) | `internal/repository/` (base patterns) | Common CRUD operations |
| `sql_statement_builder.py` (14KB) | Replaced by `squirrel` | No port needed |
| Domain models (archer, session, etc.) | `internal/model/` | Go structs with JSON tags |

### `routers/` → `internal/handler/`

All 7 v0 routers map to handler files: `auth`, `archers`, `sessions`, `slots`, `shots`, `faces`,
`live_stats`.

### `schema/` → `internal/model/`

All Pydantic schemas become Go structs with JSON tags and validation. Request/response types
colocated with domain models or in handler-specific types.

## 5. Configuration Management

Use `kelseyhightower/envconfig` for env var loading with struct tags.

**Must replicate:**

- PostgreSQL socket-vs-TCP detection (Unix socket path or TCP host:port)
- Connection pool tuning parameters (min/max conns, timeouts)
- JWT settings (secret, expiration, algorithm)
- Dev mode toggle (affects logging level, CORS, and frontend serving mode)

**Approach:** Define a `Config` struct in `internal/config/` with `envconfig` tags. Add a `Validate()`
method for cross-field validation (e.g., socket-vs-TCP logic). Load in `main.go` and pass via
constructor injection.

## 6. Database Migration Tooling Transition

Replace Flyway with `pressly/goose`.

**Migration strategy:**

- Wipe the Flyway metadata table (`flyway_schema_history`) and let goose start from version 0.
- Goose runs all migrations fresh (they must be idempotent or the DB is at the right state).
- Existing SQL migration files remain the source of truth, add `-- +goose Up` / `-- +goose Down`
  markers and rename to goose naming convention.
- **Schema evolution conventions**: The migrations repository `README.md` must document how to
  author new migrations (naming, Up/Down rules, data migration guidelines). The test script must
  dynamically discover migration files. All migration PRs must pass the full up → test → down → up
  cycle. See task 005 for details.

**Integration:**

- **Embedded in binary**: Use goose as a Go library. Run migrations on app startup or via a CLI
  subcommand (`arch-stats migrate`).
- **Docker Compose**: Replace the Flyway container with a goose-based approach (either run goose from
  the Go binary or a lightweight goose container).
- **CI (GitHub Actions)**: Run migrations as part of the test setup.

## 7. Go Design Patterns

### HTTP Routing and Middleware

- **Router**: `chi` with route groups for `/api/v0/` prefix.
- **Middleware stack**: logging → recovery → CORS → auth → error-to-HTTP mapper.
- **OpenAPI**: `swaggo/swag` annotations on each handler function.

### Dependency Injection

- **Constructor injection**: Pass dependencies (db pool, logger, config) via struct constructors.
- **Wiring in `main.go`**: Create each layer explicitly:

  ```go
  pool := pgxpool.New(ctx, cfg.DatabaseURL)
  sessionRepo := repository.NewSessionRepo(pool)
  sessionSvc := service.NewSessionService(sessionRepo)
  sessionHandler := handler.NewSessionHandler(sessionSvc)
  ```

### Error Handling

- **Custom domain errors** in `internal/apperror/`: sentinel errors like `ErrNotFound`,
  `ErrUnauthorized`, `ErrConflict`, `ErrValidation`.
- **Error wrapping**: Use `fmt.Errorf("fetching session: %w", err)` to add context as errors
  propagate up through repository → service → handler.
- **Error middleware**: Centralized middleware maps domain errors to HTTP status codes (404, 401,
  409, 422, 500).

### Database Access

- **`pgx` direct API** with `pgxpool` for connection pooling.
- **`squirrel`** for dynamic query building (SELECT, INSERT, UPDATE, DELETE with WHERE clauses).
  Raw SQL is acceptable for complex analytical/reporting queries and DDL operations (e.g.,
  materialized view refresh).
- **Repository pattern**: Each domain has a repository struct with methods like `FindByID`,
  `FindAll`, `Create`, `Update`, `Delete`.
- **Transaction support**: A `WithTx(ctx, pool, fn)` helper enables multi-repository atomic
  operations. The `DBTX` interface is satisfied by both `pgxpool.Pool` and `pgx.Tx`, so
  repositories work transparently within or outside transactions.
- **Reporting / analytics queries**: A separate `ReportingRepo` handles cross-domain read-only
  queries (joins, aggregations, window functions). This follows a lightweight CQRS separation:
  CRUD repositories serve the operational API, while the reporting repository serves future
  charts/reports/tables. Reporting queries may also be backed by PostgreSQL views.
- **Materialized view refresh**: A `MaintenanceRepo` owns `REFRESH MATERIALIZED VIEW CONCURRENTLY`
  operations. The `open_participants` view must be refreshed after slot/session state changes.
- **Schema version awareness**: The application logs the current goose migration version on startup
  and exposes it via a health endpoint.

## 8. Re-architect the Code from Scratch with TDD

Build the new Go backend from the ground up following TDD practices. Use the existing Python test
suite as a **behavioral specification**:

- Port existing endpoint tests (in `backend-old/tests/endpoints/`) as Go integration tests, they
  define the expected API behavior.
- Port existing model tests (in `backend-old/tests/models/`) as Go unit tests.
- Use **`testcontainers-go`** for integration tests, spin up a real PostgreSQL instance per test
  run, matching the current real-DB testing approach.
- Replicate the test fixtures (DB setup/teardown) using `testcontainers-go` + goose migrations.

## 9. WebSocket / Real-time Strategy

Redesign using Go's goroutines + channels (a natural fit for this pattern):

- A `pgx` connection listens for NOTIFY events in a dedicated goroutine.
- Each WebSocket client gets its own goroutine (using `gorilla/websocket` or `nhooyr.io/websocket`).
- A central "hub" goroutine broadcasts messages to connected clients via channels.
- This replaces the Python async approach with Go's native concurrency model.

## 10. Logging Approach

Use stdlib `log/slog` (Go 1.21+):

- Structured JSON output to stdout.
- Systemd journal captures stdout/stderr natively, no special integration needed.
- Configure log level based on dev/prod mode from config.
- Pass `*slog.Logger` via constructor injection to all layers.

## 11. Re-write CI Pipelines in GitHub

Each workflow changes fundamentally:

- **Backend linting** (`backend_linting.yaml`): Ruff + Ty + pytest → `golangci-lint` + `go test`.
- **Build artifact** (`build_artifact.yaml`): The pipeline becomes:
  1. Build frontend (`npm run build`).
  2. Cross-compile Go binary (`GOOS=linux GOARCH=arm64`) with embedded frontend assets.
  3. Publish the single binary as a GitHub Release.
- **Custom GitHub Actions**: `uv-setup` becomes irrelevant; `npm-setup` stays for the frontend.
- **Bash linting** (`bash_linting.yaml`): Review which scripts remain relevant.
- **Frontend linting** (`frontend_linting.yaml`): Unchanged.

## 12. Deployment Model Change

The deployment fundamentally changes from "tarball + venv + dependency install" to "single binary":

- **Build artifact**: Cross-compile on GitHub Actions (`GOOS=linux GOARCH=arm64`). The binary
  embeds the frontend via `//go:embed`. Publish as a GitHub Release.
- **Systemd service** (`Pi/arch-stats.service`): Update from
  `ExecStart=.venv/bin/uvicorn main:app` to `ExecStart=/opt/arch-stats/arch-stats` (single binary).
- **Installer scripts**: Simplify drastically, `install_app.bash` no longer needs `uv sync` or
  venv management; it becomes "download binary, verify checksum, place it, run migrations."
- **Remote installer**: Same stop/start flow but simpler internals.

## 13. Local Development Environment

Define the local dev story for Go:

- **Hot reload**: Use `air` for Go live-reload during development. Configure `.air.toml` in
  `backend/`.
- **Dev mode**: Pass `--dev` flag to the Go binary to serve frontend from disk (bypass `//go:embed`)
  so the Vite dev server handles frontend serving with HMR.
- **Docker Compose**: Replace the Flyway container, either run goose from the Go binary on startup
  or use a lightweight goose container.
- **VS Code tasks**: Update from Uvicorn/Vite tasks to Go/Vite tasks (start `air`, start Vite).
- **`.env` files**: Same approach, loaded by `envconfig`.
- **Update `.agent/` skills**: Reference Go tooling instead of Python (golangci-lint, go test, air).
