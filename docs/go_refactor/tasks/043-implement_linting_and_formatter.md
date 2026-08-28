# Task 043: Implement Go Linting and Formatting

## Git Branch

`refactor/043-implement-linting-and-formatter`

## Objective

Configure `golangci-lint` as the Go linter and `gofumpt` as the Go formatter, integrate both
into the pre-commit hook via `scripts/linting.bash`, and configure Antigravity-IDE (VS Code) to
use the same tools for on-save linting and formatting. After this task, every `git commit` that
touches Go files will automatically lint and format them, and the IDE will provide real-time
feedback using the identical configuration.

## Dependencies

- Task 001 (Go module and project scaffold — `backend/go.mod` exists)

## Acceptance Criteria

- [ ] `backend/.golangci.yml` exists with a curated linter set including at minimum: `govet`,
  `staticcheck`, `errcheck`, `gosimple`, `unused`, `gofumpt`, `revive`, and `ineffassign`.
- [ ] `golangci-lint run ./...` passes from the `backend/` directory with zero findings (or all
  existing findings are fixed).
- [ ] `gofumpt -l .` from `backend/` reports no files needing formatting (all Go files are
  already formatted).
- [ ] `scripts/linting.bash` has a new `run_go_checks` function that runs `gofumpt -l -w .`,
  `golangci-lint run ./...`, and `go test ./... -count=1` from the `backend/` directory.
- [ ] `scripts/linting.bash` usage text includes `--go` flag.
- [ ] `scripts/linting.bash` auto-detects staged `backend/**/*.go` files and triggers Go checks
  (mirroring the existing staged-file detection pattern for `frontend/`, `backend-old/`, and
  `scripts/`).
- [ ] `.vscode/settings.json` configures Go format-on-save using `gofumpt` via the `gopls`
  language server.
- [ ] `.vscode/extensions.json` recommends the `golang.go` extension.
- [ ] Running `git commit` with staged `.go` files triggers the Go linter and formatter via the
  existing pre-commit hook symlink (`.git/hooks/pre-commit` → `../../scripts/linting.bash`).
- [ ] `.github/workflows/backend_linting.yaml` is rewritten to replace all Python jobs with Go
  equivalents:
    - Trigger: `pull_request` on paths `backend/**/*.go`, `backend/go.mod`, `backend/go.sum`,
      and the workflow file itself.
    - Jobs:
      1. **lint**: runs `golangci-lint run ./...` using `golangci-lint-action`
      2. **format-check**: runs `gofumpt -l .` and fails if any files are unformatted
      3. **vet**: runs `go vet ./...`
      4. **test**: runs `go test ./... -v -count=1 -race` with a PostgreSQL 17 service container
         and goose migrations
    - Go version: `1.27.0`
    - No references to Python, uv, Ruff, Ty, Flyway, or `uv-setup` remain.

## Files to Create/Modify

| Action | Path |
| ------ | ---- |
| Create | `backend/.golangci.yml` |
| Modify | `scripts/linting.bash` |
| Modify | `.vscode/settings.json` |
| Modify | `.vscode/extensions.json` |
| Modify | `.github/workflows/backend_linting.yaml` |

## Reference

- Current linting script:
  [linting.bash](file:///home/juanpa/Projects/arch-stats/scripts/linting.bash)
- VS Code settings:
  [settings.json](file:///home/juanpa/Projects/arch-stats/.vscode/settings.json)
- VS Code extensions:
  [extensions.json](file:///home/juanpa/Projects/arch-stats/.vscode/extensions.json)
- Current CI workflow:
  [backend_linting.yaml](file:///home/juanpa/Projects/arch-stats/.github/workflows/backend_linting.yaml)
- CI linting workflow task:
  [task 032](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/032-ci_backend_linting_workflow.md)
- Agent skills update task:
  [task 035](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/035-agent_skills_update.md)

## Steps

- [ ] **Step 1: Create `backend/.golangci.yml`**

  Create the golangci-lint configuration file at `backend/.golangci.yml`:

  ```yaml
  run:
    timeout: 5m

  linters:
    enable:
      - errcheck
      - govet
      - staticcheck
      - gosimple
      - unused
      - ineffassign
      - revive
      - gofumpt
      - gocritic
      - misspell
      - prealloc
      - nolintlint

  linters-settings:
    gofumpt:
      extra-rules: true
    revive:
      rules:
        - name: blank-imports
        - name: dot-imports
        - name: exported
        - name: unexported-return
        - name: var-naming
    gocritic:
      enabled-tags:
        - diagnostic
        - style
        - performance
    misspell:
      locale: US
    nolintlint:
      require-explanation: true
      require-specific: true

  issues:
    max-issues-per-linter: 0
    max-same-issues: 0
  ```

- [ ] **Step 2: Install `golangci-lint` and `gofumpt`**

  ```bash
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
  go install mvdan.cc/gofumpt@latest
  ```

  Verify both are on `$PATH`:

  ```bash
  golangci-lint version
  gofumpt --version
  ```

- [ ] **Step 3: Format all existing Go files with `gofumpt`**

  ```bash
  cd backend
  gofumpt -l -w .
  ```

  Then verify no files are reported as unformatted:

  ```bash
  gofumpt -l .
  ```

  Expected: no output (all files formatted).

- [ ] **Step 4: Run `golangci-lint` and fix all findings**

  ```bash
  cd backend
  golangci-lint run ./...
  ```

  Fix any reported issues until the linter passes cleanly. If specific linter rules are
  inapplicable to this codebase, disable them in `.golangci.yml` with a comment explaining why.

- [ ] **Step 5: Add `run_go_checks` function to `scripts/linting.bash`**

  Add the new function after the existing `run_bash_checks` function (after line 73):

  ```bash
  run_go_checks() {
      cd "${ROOT_DIR}/backend"
      log_info "Running Go formatter (gofumpt)..."
      gofumpt -l -w .
      log_info "Running Go linter (golangci-lint)..."
      golangci-lint run ./...
      log_info "Running Go tests..."
      go test ./... -count=1
      cd -
  }
  ```

- [ ] **Step 6: Update `usage()` in `scripts/linting.bash`**

  Update the usage text to include the `--go` option:

  ```bash
  usage() {
      cat <<'EOF'
  Usage: scripts/linting.bash [--backend] [--frontend] [--scripts] [--go]

  When one or more flags are provided, only the selected checks run and staged file detection is skipped.

  Options:
      --backend   Run Python format/lint/type/tests for backend-old
      --frontend  Run JS/TS lint/format/tests for frontend
      --scripts   Run shellcheck and shfmt over scripts/*.bash
      --go        Run Go linter and formatter for backend
      -h, --help       Show this help and exit
  EOF
  }
  ```

- [ ] **Step 7: Update `main()` in `scripts/linting.bash`**

  Add `--go` flag handling and staged-file detection for `backend/` Go files:

  In the flag-parsing section, add the `--go` case:

  ```bash
  local needs_go=false
  ```

  ```bash
  --go)
      needs_go=true
      ;;
  ```

  In the staged-file detection block, add detection for Go files in `backend/`:

  ```bash
  elif [[ $file =~ ^backend/.*\.go$ ]] || [[ $file =~ ^backend/go\.(mod|sum)$ ]]; then
      needs_go=true
  ```

  At the end of `main()`, add the Go checks invocation:

  ```bash
  if $needs_go; then
      run_go_checks
  fi
  ```

- [ ] **Step 8: Update `.vscode/settings.json` for Go**

  Add Go language settings to `.vscode/settings.json`:

  ```json
  "go.lintTool": "golangci-lint",
  "go.lintFlags": ["--config=${workspaceFolder}/backend/.golangci.yml"],
  "go.lintOnSave": "workspace",
  "gopls": {
      "formatting.gofumpt": true
  },
  "[go]": {
      "editor.formatOnSave": true,
      "editor.defaultFormatter": "golang.go",
      "editor.codeActionsOnSave": {
          "source.organizeImports": "explicit"
      }
  }
  ```

- [ ] **Step 9: Update `.vscode/extensions.json`**

  Add `golang.go` to the recommendations list:

  ```json
  "golang.go"
  ```

- [ ] **Step 10: Rewrite `.github/workflows/backend_linting.yaml`**

  Replace the entire contents of `.github/workflows/backend_linting.yaml`:

  ```yaml
  name: Backend Lint, Format & Test

  on:
    pull_request:
      paths:
        - backend/**/*.go
        - backend/go.mod
        - backend/go.sum
        - backend/.golangci.yml
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
          uses: golangci/golangci-lint-action@v6
          with:
            version: latest
            working-directory: backend

    format-check:
      runs-on: ubuntu-latest
      steps:
        - name: Checkout code
          uses: actions/checkout@v4
        - name: Set up Go
          uses: actions/setup-go@v5
          with:
            go-version: "1.27.0"
        - name: Install gofumpt
          run: go install mvdan.cc/gofumpt@latest
        - name: Check formatting
          run: |
            unformatted=$(gofumpt -l .)
            if [ -n "$unformatted" ]; then
              echo "The following files are not formatted with gofumpt:"
              echo "$unformatted"
              exit 1
            fi

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
      needs: [lint, format-check, vet]
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

- [ ] **Step 11: Verify no Python references remain in the workflow**

  ```bash
  grep -iE "ruff|pytest|uv-setup|uvicorn|flyway|uv run|pyproject|python" .github/workflows/backend_linting.yaml
  ```

  Expected: no results.

- [ ] **Step 12: Validate YAML syntax**

  ```bash
  uv run python -c "import yaml; yaml.safe_load(open('.github/workflows/backend_linting.yaml'))"
  ```

- [ ] **Step 13: End-to-end verification**

  Stage a Go file and verify the pre-commit hook runs Go checks:

  ```bash
  cd backend
  touch internal/dummy_test.go
  echo 'package internal' > internal/dummy_test.go
  git add internal/dummy_test.go
  git commit --dry-run
  ```

  Then clean up:

  ```bash
  git reset HEAD internal/dummy_test.go
  rm internal/dummy_test.go
  ```

## Verification

- `cd backend && golangci-lint run ./...` — passes with zero findings.
- `cd backend && gofumpt -l .` — reports no unformatted files.
- `scripts/linting.bash --go` — runs Go linter and formatter successfully.
- `grep -q "golang.go" .vscode/extensions.json` — extension is recommended.
- `grep -q "gofumpt" .vscode/settings.json` — gofumpt is configured for format-on-save.
- `grep -q "golangci-lint" .vscode/settings.json` — golangci-lint is configured as the linter.
- Staging a `.go` file and running `scripts/linting.bash` (without flags) detects and runs Go
  checks.
- `grep -iE "ruff|pytest|uv-setup|flyway|uv run" .github/workflows/backend_linting.yaml` —
  no Python references remain in the CI workflow.
- `.github/workflows/backend_linting.yaml` contains `golangci-lint`, `gofumpt`, `go vet`, and
  `go test` jobs.
