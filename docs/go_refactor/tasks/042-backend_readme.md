# Task 042: Update Backend README for Go Ecosystem and Deployment

## Git Branch

`refactor/042-backend-readme`

## Objective

Rewrite `backend/README.md` to reflect the Go backend architecture, replacing all outdated
Python/Uvicorn/uv/Pydantic references. The new documentation must provide comprehensive guidance
for new Go developers onboarding to the project, explain the updated GitHub Actions CI/CD workflows,
and document how to install the single-binary application on a Raspberry Pi 5 (including testing the
installation scripts locally using the Docker Raspberry Pi emulator).

## Dependencies

- Task 031 (air hot reload and local dev environment)
- Task 032 (CI backend linting workflow)
- Task 033 (CI build artifact and cross-compilation workflow)
- Task 034 (deployment scripts and systemd service)

## Acceptance Criteria

- [ ] `backend/README.md` is updated and completely free of Python, uv, Uvicorn, FastAPI, Pydantic,
  or Flyway references.
- [ ] **Development Environment Guide for Go Developers**:
    - Prerequisites listed: Go 1.27.0+, Docker & Docker Compose, `air` (live reload), `swag`
      (OpenAPI generation), `golangci-lint`.
    - Local setup instructions: cloning, module download (`go mod download`), database startup via
      `docker compose -f ./docker/docker-compose.yaml up -d`.
    - How to run database migrations with goose / binary CLI (`go run ./cmd/arch-stats migrate`).
    - How to run the dev server with hot reload (`air`) or standalone (`go run ./cmd/arch-stats`).
    - How to run tests (`go test ./...` and `go test -v ./...`) and linting (`golangci-lint run`).
    - How to regenerate OpenAPI annotations (`swag init`) and sync TypeScript frontend types
      (`npm run generate:types`).
    - VS Code task integration (`Start Go Server (air)`, `Run Go Tests`, etc.).
- [ ] **GitHub Workflows Explanation**:
    - Explains `backend_linting.yaml` (linting, vetting, unit & integration tests).
    - Explains `frontend_linting.yaml` and `bash_linting.yaml`.
    - Explains `build_artifact.yaml` release pipeline: frontend build, embedding static assets via
      `//go:embed`, cross-compiling for `linux/arm64` (Raspberry Pi 5), generating SHA256 checksums,
      and publishing GitHub Releases with the single binary.
- [ ] **Raspberry Pi 5 Installation & Emulator Testing**:
    - Describes target deployment model: single `/opt/arch-stats/arch-stats` binary, `.env`
      configuration, and `arch-stats.service` systemd unit.
    - Instructions on deploying / installing to a physical Raspberry Pi 5 using `scripts/deploy.bash`,
      `scripts/install_app.bash`, and `scripts/remote_installer.bash`.
    - Step-by-step guide for testing installation scripts locally using the Docker Raspberry Pi
      emulator (`docker compose --profile emulator up -d`), connecting via port 2222 with SSH keys,
      and verifying systemd service execution.
- [ ] Markdown complies with markdownlint standards (including MD007 4-space list indentation).

## Files to Create/Modify

| Action | Path |
| ------ | ---- |
| Modify | `backend/README.md` |

## Reference

- Plan §11: CI/CD Workflows (`backend_linting.yaml`, `build_artifact.yaml`)
- Plan §12: Deployment & Raspberry Pi 5 Model
- Emulator design:
  [rpi emulator](file:///home/juanpa/Projects/arch-stats/docs/plans/2026-07-02-raspberry-pi-emulator-design.md)
- Deployment scripts task:
  [task 034](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/034-deployment_scripts_and_systemd.md)
- CI workflow tasks:
  [task 032](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/032-ci_backend_linting_workflow.md)
  and [task 033](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/033-ci_build_artifact_workflow.md)

## Steps

- [ ] **Step 1: Draft the Architecture & Getting Started Section**

  Define the Go stack: Chi router, pgx pool, squirrel SQL builder, envconfig, goose migrations,
  and embedded frontend serving. Detail prerequisites and local environment bootstrap steps
  (Docker database, `air` live reload, VS Code tasks).

- [ ] **Step 2: Document Code Quality, Testing, and OpenAPI Workflows**

  Document commands for:
    - Unit and integration tests: `go test -v ./...`
    - Linting and static analysis: `golangci-lint run` and `go vet ./...`
    - OpenAPI spec generation with `swag init -g cmd/arch-stats/main.go -o specs/`
    - Syncing frontend types with `npm run generate:types`

- [ ] **Step 3: Document GitHub Actions CI/CD Pipeline**

  Document the CI/CD matrix and workflow triggers:
    - `backend_linting.yaml`: Trigger conditions, steps executed.
    - `build_artifact.yaml`: Cross-compilation target (`linux/arm64`), asset embedding, release
      artifact generation.

- [ ] **Step 4: Document Raspberry Pi 5 Installation & Docker Emulator Testing**

  Document:
    - How `scripts/install_app.bash` downloads and installs the standalone Go binary on
      Raspberry Pi 5.
    - How systemd manages `/opt/arch-stats/arch-stats`.
    - How to spin up the local emulator container (`docker compose --profile emulator up -d`).
    - How to run `scripts/deploy.bash` or `scripts/remote_installer.bash` against `localhost:2222`.
    - How to verify systemctl status and API response inside the emulator.

## Verification

- [ ] Verify `backend/README.md` contains accurate, executable commands for Go developers.
- [ ] Run markdownlint on `docs/go_refactor/tasks/042-backend_readme.md` and `backend/README.md`.
