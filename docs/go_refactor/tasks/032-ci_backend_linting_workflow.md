# Task 032: Rewrite Backend Linting CI Workflow

## Git Branch

`refactor/032-ci-backend-linting-workflow`

## Objective

Rewrite the `.github/workflows/backend_linting.yaml` CI workflow for Go. Replace Ruff + Ty +
pytest with `golangci-lint` + `go test`. The workflow must lint, vet, and test the Go backend
on every PR that touches backend files.

## Dependencies

- Task 025 (Go backend compiles and tests pass)
- Task 005 (goose migrations — tests may need migrations)

## Acceptance Criteria

- [x] `.github/workflows/backend_linting.yaml` is rewritten with:
    - Trigger: `pull_request` on paths `backend/**/*.go`, `backend/go.mod`, `backend/go.sum`
    - Jobs:
      1. **lint**: runs `golangci-lint run` on the backend
      2. **test**: runs `go test ./... -v -count=1` with a PostgreSQL service container
         and goose migrations (for integration tests)
    - Go version: `1.27.0`
    - Uses `golangci-lint-action` for caching
- [x] The `uv-setup` custom action is no longer referenced.
- [x] PostgreSQL service container runs version 17.
- [x] Migrations run via the Go binary (`./arch-stats migrate`) instead of Flyway.
- [x] The workflow passes when pushed to a branch (dry run).

## Files to Modify

| Action | Path |
| ------ | ---- |
| Modify | `.github/workflows/backend_linting.yaml` |

## Reference

- Current workflow: [backend_linting.yaml](file:///home/juanpa/Projects/arch-stats/.github/workflows/backend_linting.yaml)
- Plan §11: `golangci-lint` + `go test`

## Steps

- [x] **Step 1: Rewrite the workflow**

  Replace the entire contents of `.github/workflows/backend_linting.yaml`:

  ```yaml
  name: Backend Lint & Test

  on:
    pull_request:
      paths:
        - backend/**/*.go
        - backend/go.mod
        - backend/go.sum
        - .github/workflows/backend_linting.yaml

  permissions:
    contents: read

  defaults:
    run:
      shell: bash
      working-directory: backend

  jobs:
    lint:
      runs-on: ubuntu-latest
      steps:
        - name: Checkout code
          uses: actions/checkout@v4
        - name: Set up Go
          uses: actions/setup-go@v5
          with:
            go-version: "1.27.0"
        - name: Run golangci-lint
          uses: golangci/golangci-lint-action@v9
          with:
            version: latest
            working-directory: backend

    vet:
      runs-on: ubuntu-latest
      steps:
        - name: Checkout code
          uses: actions/checkout@v4
        - name: Set up Go
          uses: actions/setup-go@v5
          with:
            go-version: "1.27.0"
        - name: Run go vet
          run: go vet ./...

    test:
      runs-on: ubuntu-latest
      needs: [lint, vet]
      services:
        postgres:
          image: postgres:17
          env:
            POSTGRES_USER: "admin"
            POSTGRES_PASSWORD: "NOT_A_REAL_PASSWORD"
            POSTGRES_DB: "arch-stats"
          ports:
            - 5432:5432
          options: >-
            --health-cmd "pg_isready -U admin"
            --health-interval 10s
            --health-timeout 5s
            --health-retries 5
      steps:
        - name: Checkout code
          uses: actions/checkout@v4
        - name: Set up Go
          uses: actions/setup-go@v5
          with:
            go-version: "1.27.0"
        - name: Build binary
          run: go build -o ./arch-stats ./cmd/arch-stats
        - name: Run migrations
          env:
            POSTGRES_USER: "admin"
            POSTGRES_PASSWORD: "NOT_A_REAL_PASSWORD"
            POSTGRES_DB: "arch-stats"
            POSTGRES_HOST: "127.0.0.1"
            POSTGRES_PORT: "5432"
            ARCH_STATS_DEV_MODE: "true"
            ARCH_STATS_GOOGLE_OAUTH_CLIENT_ID: "NOT_A_REAL_CLIENT_ID"
            ARCH_STATS_JWT_SECRET: "dev-insecure-change-me"
          run: ./arch-stats migrate
        - name: Run tests
          env:
            POSTGRES_USER: "admin"
            POSTGRES_PASSWORD: "NOT_A_REAL_PASSWORD"
            POSTGRES_DB: "arch-stats"
            POSTGRES_HOST: "127.0.0.1"
            POSTGRES_PORT: "5432"
            ARCH_STATS_DEV_MODE: "true"
            ARCH_STATS_GOOGLE_OAUTH_CLIENT_ID: "NOT_A_REAL_CLIENT_ID"
            ARCH_STATS_JWT_SECRET: "dev-insecure-change-me"
          run: go test ./... -v -count=1 -race
  ```

- [x] **Step 2: Validate YAML syntax**

  ```bash
  python3 -c "import yaml; yaml.safe_load(open('.github/workflows/backend_linting.yaml'))"
  ```

- [x] **Step 3: Verify no Python references remain**

  ```bash
  grep -i "ruff\|pytest\|uv-setup\|uvicorn\|flyway" .github/workflows/backend_linting.yaml
  ```

  Expected: no results.

- [x] **Step 4: Commit**

  ```bash
  git add -A
  git commit -m "ci: rewrite backend linting workflow for Go (golangci-lint + go test)"
  ```

## Verification

- `cat .github/workflows/backend_linting.yaml` — references Go tooling, not Python.
- YAML is valid.
- No references to Ruff, pytest, uv, Flyway, or Uvicorn.
