# Arch Stats - Go Backend

The backend is a statically compiled Go service that powers the Arch Stats archery tracking
platform. It exposes a REST API, serves the embedded Vue 3 SPA, and pushes real-time updates
over WebSocket.

## Architecture Overview

| Layer | Package | Responsibility |
| ----- | ------- | -------------- |
| Entrypoint | `cmd/arch-stats/` | CLI flags, wiring, graceful shutdown |
| Configuration | `internal/config/` | Environment-based config (`envconfig` style) |
| HTTP Routing | `internal/handler/` | Chi v5 route handlers, SPA serving |
| Middleware | `internal/middleware/` | Auth, logging, CORS, request-ID |
| Auth | `internal/auth/` | JWT + Google OAuth, session tokens |
| Services | `internal/service/` | Business logic, orchestration |
| Repositories | `internal/repository/` | pgx/v5 pool, Squirrel SQL, goose migrations |
| Models | `internal/model/` | Domain structs with JSON tags |
| Errors | `internal/apperror/` | Sentinel errors, `Wrap` helper |
| WebSocket | `internal/websocket/` | Real-time hub, Postgres NOTIFY relay |
| Embed | `embed.go` | `//go:embed` for bundled frontend assets |

**Data flow:** Sensor/bot inserts rows → PostgreSQL `NOTIFY` → WebSocket hub → frontend
renders live updates.

## Prerequisites

Install these tools before starting development:

| Tool | Version | Purpose |
| ---- | ------- | ------- |
| [Go](https://go.dev/dl/) | 1.27.0+ | Language runtime |
| [Docker](https://docs.docker.com/get-docker/) & Docker Compose | latest | Local PostgreSQL database |
| [air](https://github.com/air-verse/air) | latest | Live-reload dev server |
| [swag](https://github.com/swaggo/swag) | latest | OpenAPI spec generation |
| [golangci-lint](https://golangci-lint.run/welcome/install/) | latest | Go linter aggregate |
| [gofumpt](https://github.com/mvdan/gofumpt) | latest | Strict Go formatter |
| [Node.js & npm](https://nodejs.org/) | LTS | Frontend type generation |

Install the Go tools:

```bash
go install github.com/air-verse/air@latest
go install github.com/swaggo/swag/cmd/swag@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install mvdan.cc/gofumpt@latest
```

## Getting Started

### 1. Clone and install dependencies

```bash
git clone --recurse-submodules https://github.com/jpmolinamatute/arch-stats.git
cd arch-stats/backend
go mod download
```

> **Note:** The `--recurse-submodules` flag is required because database migrations live in a
> private Git submodule (`backend/migrations` →
> `git@github.com:jpmolinamatute/arch-stats-migrations.git`). You need SSH access to this
> repository. If you already cloned without the flag, initialize the submodule manually:
>
> ```bash
> git submodule update --init --recursive
> ```

### 2. Start the database

```bash
docker compose -f ../docker/docker-compose.yaml --profile dev up -d
```

This starts a PostgreSQL 17 container with custom configuration. The database credentials are
loaded from `docker/.env`.

### 3. Run database migrations

Migrations are SQL files stored in a **private Git submodule** at `backend/migrations/`
(sourced from `git@github.com:jpmolinamatute/arch-stats-migrations.git`). At runtime, the
binary uses [goose](https://github.com/pressly/goose) to apply them:

```bash
go run ./cmd/arch-stats migrate
```

Alternatively, set `APPLY_DB_MIGRATIONS_ON_START=true` in your `.env` to auto-migrate on
server startup.

### 4. Start the development server

**With hot reload (recommended):**

```bash
air
```

`air` watches `.go` and `.toml` files, rebuilds, and restarts automatically. Configuration
lives in `.air.toml`.

**Without hot reload:**

```bash
go run ./cmd/arch-stats
```

Add `--dev` for development mode (pretty logging, relaxed CORS):

```bash
go run ./cmd/arch-stats --dev
```

The server listens on the port defined by `ARCH_STATS_SERVER_PORT` (default `8000`).

### 5. VS Code task integration

The workspace includes pre-configured VS Code tasks in `.vscode/tasks.json`:

| Task | Description |
| ---- | ----------- |
| `Start Docker Compose` | Launches the PostgreSQL container |
| `Stop Docker Compose` | Stops the container |
| `Stop & Remove Volumes` | Stops containers and removes volumes |
| `Start Go Server (air)` | Runs `air` in the backend directory |
| `Start Vite Server` | Runs the Vue 3 frontend dev server |
| `Run Go Tests` | Executes `go test` on the backend |

Access tasks via **Terminal → Run Task** or `Ctrl+Shift+P → Tasks: Run Task`.

## Code Quality, Testing, and OpenAPI

### Running tests

```bash
# All tests
go test ./... -v

# With race detection
go test -race ./... -v -count=1

# Specific package
go test -v ./internal/service/...
```

### Linting and static analysis

The project uses [golangci-lint](https://golangci-lint.run/) with the configuration in
`.golangci.yml`. Enabled linters include `gocritic`, `misspell`, `revive`, `prealloc`, and
`gofumpt` formatting.

```bash
# Run linter
golangci-lint run ./...

# Run with auto-fix
golangci-lint run --fix ./...

# Run go vet
go vet ./...

# Run gofumpt formatter
gofumpt -l -w .
```

**Full lint suite** (lint + format + tests via the project script):

```bash
./scripts/linting.bash --go
```

### OpenAPI spec generation

The API is annotated with [swag](https://github.com/swaggo/swag) comments. Regenerate the
Swagger 2.0 spec after changing handler annotations:

```bash
swag init -g cmd/arch-stats/main.go -o specs/
```

### Syncing frontend TypeScript types

After modifying API response models, regenerate the frontend types so the Vue 3 SPA stays in
sync:

```bash
cd ../frontend
npm run generate:types
```

Or use the project script (handles offline spec conversion automatically):

```bash
../scripts/generate_fe_types.bash
```

This runs `swag init` (if needed), converts the Swagger 2.0 spec to OpenAPI 3.0, enriches it,
and generates TypeScript types into `frontend/src/types/types.generated.ts`.

## GitHub Actions CI/CD Pipelines

### `backend_linting.yaml` - Lint, Format & Test

**Trigger:** Pull requests that touch `backend/**/*.go`, `backend/go.mod`, `backend/go.sum`,
`backend/.golangci.yml`, or the workflow file itself.

**Jobs:**

| Job | What it does |
| --- | ------------ |
| `lint` | Runs `golangci-lint` via the official action |
| `format-check` | Verifies all files are `gofumpt`-formatted |
| `vet` | Runs `go vet ./...` |
| `test` | Fetches migrations (`MIGRATIONS_PAT`), builds, migrates PG 17, `go test -race` |

The `test` job depends on `lint`, `format-check`, and `vet` passing first. It provisions a
PostgreSQL 17 service container with health checks, checks out the private migrations
submodule using `secrets.MIGRATIONS_REPOSITORY` and `secrets.MIGRATIONS_PAT`, and runs the
full test suite with race detection.

### `frontend_linting.yaml` - Frontend Linting & Tests

**Trigger:** Pull requests touching `frontend/**`.

**Jobs:**

| Job | What it does |
| --- | ------------ |
| `lint` | Runs ESLint via `npm run lint` |
| `tests` | Runs frontend tests via `npm run test` (depends on `lint`) |

### `bash_linting.yaml` - Bash Linting & Formatting

**Trigger:** Pull requests touching `scripts/*.bash`.

**Jobs:**

| Job | What it does |
| --- | --- |
| `format` | Checks formatting with `shfmt` (4-space indent) |
| `lint` | Runs `shellcheck` with bash dialect |

### `build_artifact.yaml` - Release Pipeline

**Trigger:** Push to `main` branch.

**Pipeline steps (single job):**

1. **Checkout** code and set up Go + Node.js
2. **Generate OpenAPI spec** `swag init -g cmd/arch-stats/main.go -o specs/`
3. **Generate frontend types** `npm run generate:types`
4. **Build frontend** `npm run build` (output goes to `backend/frontend/`)
5. **Cross-compile Go binary** targets `linux/arm64` with
   `CGO_ENABLED=0 go build -ldflags="-s -w"` for Raspberry Pi 5
6. **Generate SHA256 checksum** `sha256sum arch-stats > arch-stats.sha256`
7. **Create GitHub Release** Tags `v<run_number>`, attaches the binary and checksum

The binary embeds the compiled frontend via `//go:embed all:frontend` in `embed.go`, producing
a single self-contained executable.

## Raspberry Pi 5 Deployment

### Target deployment model

The production environment is a Raspberry Pi 5 running Debian (Bookworm). The deployment
consists of:

| Component | Path |
| --------- | ---- |
| Application binary | `/opt/arch-stats/arch-stats` |
| Environment config | `/opt/arch-stats/.env` |
| SQL migrations | `/opt/arch-stats/migrations/*.sql` |
| systemd service | `/etc/systemd/system/arch-stats.service` |

The systemd unit (`scripts/templates/arch-stats.service.j2`) runs the binary as a dedicated
`arch-stats` system user:

```ini
[Service]
Type=simple
User=arch-stats
WorkingDirectory=/opt/arch-stats
ExecStart=/opt/arch-stats/arch-stats
EnvironmentFile=/opt/arch-stats/.env
Restart=on-failure
RestartSec=5
```

### Deployment scripts

Three scripts orchestrate deployment from your workstation:

- **`scripts/deploy.bash`** Main entry point. Auto-detects install vs.
  update, renders templates, uploads assets via SCP, runs the installer
  over SSH.
- **`scripts/remote_installer.bash`** Runs on the Pi during fresh
  install. Installs OS packages, creates the app user, configures
  PostgreSQL, sets up Cloudflared, registers the systemd service.
- **`scripts/install_app.bash`** Downloads the latest GitHub Release
  binary, verifies SHA256, installs to `/opt/arch-stats/`, runs
  migrations, restarts the service.

**Deploy to a physical Raspberry Pi 5:**

```bash
# Fresh install (auto-detected)
./scripts/deploy.bash -i ~/.ssh/id_rsa pi@192.168.1.100

# Explicit install
./scripts/deploy.bash -i ~/.ssh/id_rsa pi@192.168.1.100 install

# Update only (binary + migrations)
./scripts/deploy.bash -i ~/.ssh/id_rsa pi@192.168.1.100 update

# Uninstall
./scripts/deploy.bash -i ~/.ssh/id_rsa pi@192.168.1.100 uninstall
```

Required environment variables for fresh install:

- `GITHUB_TOKEN` GitHub PAT for downloading release artifacts
- `ARCH_STATS_GOOGLE_OAUTH_CLIENT_ID` Google OAuth client ID
- `ARCH_STATS_JWT_SECRET` JWT signing secret
- `CLOUDFLARED_TUNNEL_ID` Cloudflare tunnel ID

### Testing with the Docker Raspberry Pi emulator

The project includes a Docker-based emulator that mimics the Raspberry Pi environment for
testing deployment scripts locally without physical hardware.

> **Host aliases setup:** To use the same hostnames from
> [`docker-compose.yaml`](../docker/docker-compose.yaml) on your host machine, add the
> following entries to `/etc/hosts`:
>
> ```text
> 127.0.0.1        localhost db emulator
> ::1              localhost db emulator
> ```
>
> This lets you reference the `db` and `emulator` services by name both inside Docker
> (via the bridge network) and directly from the host.

**1. Start the emulator:**

```bash
docker compose -f docker/docker-compose.yaml --profile emulator up -d
```

This builds the emulator image from `docker/Dockerfile.rpi` (based on
`jrei/systemd-debian:12`) and starts it with SSH on port `2222`. The container runs systemd,
PostgreSQL, and SSH everything the real Pi would have.

**2. Connect via SSH:**

```bash
ssh -p 2222 -i docker/ssh/arch_stats_dev root@emulator
```

The emulator uses a pre-configured SSH key pair located in `docker/ssh/`.

**3. Run deployment scripts against the emulator:**

```bash
# Full deployment
./scripts/deploy.bash -p 2222 -i docker/ssh/arch_stats_dev root@emulator

# Or target a specific action
./scripts/deploy.bash -p 2222 -i docker/ssh/arch_stats_dev root@emulator install
```

**4. Verify the service inside the emulator:**

```bash
# Check systemd service status
ssh -p 2222 -i docker/ssh/arch_stats_dev root@emulator "systemctl status arch-stats.service"

# Verify the API responds
ssh -p 2222 -i docker/ssh/arch_stats_dev root@emulator \
    "curl -s http://emulator:8001/api/v0/health"
```

**5. Stop the emulator:**

```bash
docker compose -f docker/docker-compose.yaml --profile emulator down
```
