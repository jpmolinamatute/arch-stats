---
name: backend-linting
description: How to run golangci-lint and gofumpt to lint, format, and check Go backend code
---

# Go Backend Linting & Formatting

We use `golangci-lint` (which includes `govet`, `staticcheck`, `errcheck`, `gofumpt`, `revive`, and
`ineffassign`) to lint, format, and statically analyze all Go code under `backend/`. Configuration is
defined in `backend/.golangci.yml`.

## How to Run

### 1. Auto-fix Formatting and Lint Issues

To automatically fix format errors and auto-fixable lint issues:

```bash
cd backend
golangci-lint run --fix ./...
# or format directly with gofumpt:
gofumpt -l -w .
```

### 2. Full Project Go Verification Script (From project root)

Runs formatter (`gofumpt`), linter (`golangci-lint`), and unit tests (`go test`):

```bash
./scripts/linting.bash --go
```

## IDE Integration

Antigravity IDE / VS Code automatically formats on save using `gofumpt` and displays inline lint
diagnostics via `golangci-lint`.
