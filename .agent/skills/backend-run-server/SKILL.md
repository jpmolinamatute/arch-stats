---
name: backend-run-server
description: How to start the Go backend server in development mode (with air live-reload) or production mode
---

# Run Backend Server

This skill describes how to run the Go backend server in local development or production mode.

## Prerequisites

1. **Docker Infrastructure:** PostgreSQL 17 container must be running:
   ```bash
   docker compose -f docker/docker-compose.yaml up -d
   ```
2. **Environment Variables:** Verify `backend/.env` exists or required environment variables are set (see `internal/config/config.go`).

## Execution Options

### Option 1: Live-Reload Dev Mode with `air` (Recommended for development)

From the `backend/` directory:

```bash
cd backend
air
```

Or trigger the VS Code task: `"Start Go Server (air)"`.

`air` monitors `.go` and `.toml` files in `backend/` and automatically recompiles and restarts the server on change.

### Option 2: One-Shot Run (Direct execution)

To run without live reload:

```bash
cd backend
go run ./cmd/arch-stats
```

### Option 3: Compile and Run Binary (Production-like)

```bash
cd backend
go build -o arch-stats ./cmd/arch-stats
./arch-stats
```

## Verification

- Check stdout for the startup banner: `arch-stats starting (dev_mode=true)`
- Verify database connection pool initialization succeeds.
- Confirm server is listening on the configured port (default `:8080`).
