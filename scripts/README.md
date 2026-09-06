# Arch Stats Scripts

## Description

The **Arch Stats scripts** directory contains the automation and tooling that powers the
development, CI/CD, and deployment workflows of the Arch Stats platform. These scripts abstract
complex operations like environment setup, database management, and release orchestration into
simple, reusable commands.

## Development Guidelines

**Audience:** DevOps engineers and developers working in the `scripts/` directory.

## Architecture

The scripting ecosystem is built on standard, portable tools:

- **Language**: [Bash](https://www.gnu.org/software/bash/) (Strict mode: `set -Eeuo pipefail`)
- **Linting**: [ShellCheck](https://www.shellcheck.net/)
- **Formatting**: [shfmt](https://github.com/mvdan/sh)
- **Containerization**: [Docker](https://www.docker.com/) & Docker Compose
- **CI/CD**: GitHub Actions

## Prerequisites

Ensure your environment meets these requirements to run and develop these scripts:

- **OS**: Linux.
- **Shell**: Bash v5.3+.
- **Docker**: Required for database management and testing.
- **Tools**: `curl`, `jq`, `git`, `gh` (GitHub CLI).

## Setup

No specific installation is required for the scripts themselves, but you can set up the development
environment for script editing:

1. **Install ShellCheck & shfmt**:

   Most Linux distributions include these:

   ```bash
   sudo apt install shellcheck
   # shfmt usually requires a separate install or go install mvdan.cc/sh/v3/cmd/shfmt@latest
   ```

2. **VS Code Extensions**:

   Install the **ShellCheck** and **shfmt** extensions for real-time feedback.

## Scripts

Here are the primary scripts used in the project lifecycle:

| Script | Description |
| :------ | :--------- |
| [`create_pr.bash`](./create_pr.bash) | Automates PR creation with labels (`frontend`, `backend`, `documentation`) based on changed files vs `origin/main`. |
| [`deploy.bash`](./deploy.bash) | Orchestrates remote deployment via SSH — auto-detects install vs update, renders templates, uploads assets, and runs the remote installer. Also supports uninstalling. |
| [`enrich_openapi.py`](./enrich_openapi.py) | Post-processes an OpenAPI 3.0 JSON spec to add schema aliases, validation-error schemas, and `nullable` annotations for frontend TypeScript compatibility. |
| [`generate_fe_types.bash`](./generate_fe_types.bash) | Generates frontend TypeScript types from the backend OpenAPI schema. Fetches from the live server or regenerates via `swag` when offline. |
| [`install_app.bash`](./install_app.bash) | Downloads the latest release binary and checksum from GitHub, verifies SHA-256, installs the binary, runs database migrations, and restarts the systemd service. Runs on the remote server as root. |
| [`linting.bash`](./linting.bash) | All-in-one linting runner for frontend (ESLint, Prettier, Vitest, vue-tsc build), bash (ShellCheck, shfmt), and Go (gofumpt, golangci-lint, go test). Auto-detects staged files when used as a pre-commit hook. |
| [`remote_installer.bash`](./remote_installer.bash) | Full first-time server provisioning: installs OS packages (PostgreSQL, cloudflared), creates the app user, generates the `.env` file, configures PostgreSQL and Cloudflare Tunnel, registers systemd services, and delegates binary installation to `install_app.bash`. |
| [`remote_uninstaller.bash`](./remote_uninstaller.bash) | Reverses a remote installation: stops and removes systemd services, drops the PostgreSQL database and user, purges cloudflared, and deletes the app user and home directory. |

## Type Generation

To keep frontend TypeScript types in sync with the backend Go models, use the generation script:

```bash
./scripts/generate_fe_types.bash
```

The script intelligently determines the source of the OpenAPI schema:

1. **Server Running**: If the backend is running, it fetches the schema directly from
   `http://localhost:<ARCH_STATS_SERVER_PORT>/api/openapi.json` (defaults to port `8000`).
2. **Server Stopped**: If the backend is not running, it regenerates Swagger 2.0 specs via `swag`
   and converts them to OpenAPI 3.0 using `swagger2openapi`.
3. **Enrichment**: If `enrich_openapi.py` is present, it post-processes the OpenAPI 3.0 spec to add
   schema aliases, validation error schemas, and nullable field annotations for full frontend
   compatibility.

## Git Hooks & Safety Net

To prevent committing broken code, "activate" the git pre-commit hook by creating a symlink to the
linting script:

```bash
ln -sfr ./scripts/linting.bash ./.git/hooks/pre-commit
```

This ensures that all linters and tests pass before a commit is allowed.

### Pull Requests

After pushing your changes, use the helper script to create a Pull Request:

```bash
./scripts/create_pr.bash
```

> [!IMPORTANT]
> Workflows are triggered selectively based on the files changed (e.g., frontend changes do not
> trigger backend tests). All **triggered** workflows must pass for a PR to be mergeable.

## Structure

The directory is organized by function:

- **`scripts/*.bash`**: Executable scripts for development and deployment tasks.
- **`scripts/enrich_openapi.py`**: Python post-processor for OpenAPI spec enrichment.
- **`scripts/lib/`**: Shared Bash libraries sourced by other scripts.
    - `logging` Colored `log_info` / `log_error` helpers.
    - `manage_docker` Docker Compose lifecycle helpers (`start_docker`, `stop_docker`, `is_docker_running`).
- **`scripts/cloudflared/`**: Cloudflare Tunnel systemd service unit file.
- **`scripts/pg_conf/`**: Custom PostgreSQL configuration files (`postgresql.conf`, `secondary.conf`).
- **`scripts/templates/`**: Jinja2-style templates rendered by `deploy.bash` at deploy time.
    - `arch-stats.service.j2` Systemd unit for the Arch Stats application.
    - `cloudflared_config.yaml.j2` Cloudflare Tunnel configuration.
    - `pg_hba.conf.j2` PostgreSQL host-based authentication.
- **`.github/workflows/`**: CI/CD pipeline definitions.
- **`.github/actions/`**: Local composite actions (`npm-setup`).

### Style Guidelines

> [!IMPORTANT]
> All scripts must adhere to the **Development Guidelines** below to ensure safety and portability.

### Coding Standards

1. **Shebang**: Always use `#!/usr/bin/env bash`.
2. **Strict Mode**: Start every script with `set -Eeuo pipefail`.
3. **Indentation**: Use **4 spaces**.
4. **Output**: Use `printf` over `echo -e`. Errors must go to `stderr` (`>&2`).
5. **Usage**: Provide a `usage()` function for scripts with arguments.
6. **Cleanup**: Use `trap` for cleaning up temporary files/directories.

### Example Script Pattern

```bash
#!/usr/bin/env bash

set -Eeuo pipefail

usage() {
    cat <<EOF
Usage: $(basename "${0}") [-f FILE] [-n NUM] [-h]
Description: Example script pattern with argument parsing.

Options:
  -f FILE   Input file (required)
  -n NUM    Sample size (default: 10)
  -h        Show this help message
EOF
    exit 1
}

main() {
    local file=""
    local num="10"

    while getopts ":f:n:h" opt; do
        case "$opt" in
            f) file="$OPTARG" ;;
            n) num="$OPTARG" ;;
            h) usage ;;
            \?) echo "ERROR: invalid option: -$OPTARG" >&2; usage ;;
            :)  echo "ERROR: option -$OPTARG requires an argument" >&2; usage ;;
        esac
    done
    shift $((OPTIND - 1))

    # Validation
    [[ -n "$file" ]] || { echo "ERROR: -f FILE is required" >&2; usage; }
    [[ -r "$file" ]] || { echo "ERROR: cannot read file: $file" >&2; exit 3; }

    echo "Processing $file with size $num..."
}

main "$@"
```

### Code Quality

All scripts must pass the strict linting suite:

```bash
./scripts/linting.bash --scripts
```

This runs **ShellCheck** and **shfmt** to ensure code correctness and consistent style.

## CI/CD Workflows

The repository uses GitHub Actions for continuous integration and deployment.

- **Frontend**: Linting & Formatting ([`frontend_linting.yaml`](../.github/workflows/frontend_linting.yaml))
- **Backend**: Linting & Tests ([`backend_linting.yaml`](../.github/workflows/backend_linting.yaml))
- **Scripts**: Bash Linting ([`bash_linting.yaml`](../.github/workflows/bash_linting.yaml))
- **Release**: Build Artifact ([`build_artifact.yaml`](../.github/workflows/build_artifact.yaml))

### Merge Requirements

For a PR to be mergeable, the following workflows must pass if triggered:

| Workflow | Triggers on Changes In | Must Pass |
| :------ | :--------------------- | :-------- |
| **Frontend** | `frontend/**` | Formatting, Linting, Tests |
| **Backend** | `backend/**` | golangci-lint, gofumpt, go test |
| **Scripts** | `scripts/*.bash` | ShellCheck, shfmt |

> [!NOTE]
> The **Release** workflow runs only on `push` to `main` and is not a PR requirement.

## Environment Variables

Deployment and CI jobs must surface required runtime variables explicitly.

| Variable | Purpose | Required | Default |
| :------- | :------ | :------- | :------ |
| `GITHUB_TOKEN` | GitHub API token for release downloads | yes | *(secret)* |
| `POSTGRES_USER` | Postgres user | yes | `arch-stats` |
| `POSTGRES_PASSWORD` | Postgres password | yes | *(auto-generated)* |
| `POSTGRES_DB` | Database name | yes | `arch-stats` |
| `POSTGRES_HOST` | Postgres host | no | *(empty,  uses unix socket)* |
| `POSTGRES_PORT` | Postgres port | no | `5432` |
| `POSTGRES_SOCKET_DIR` | Postgres unix socket directory | no | `/var/run/postgresql` |
| `POSTGRES_POOL_MIN_SIZE` | Min pool connections | no | `1` |
| `POSTGRES_POOL_MAX_SIZE` | Max pool connections | no | `10` |
| `ARCH_STATS_SERVER_PORT` | Application HTTP port | no | `8001` |
| `ARCH_STATS_DEV_MODE` | Enable development mode | no | `false` |
| `ARCH_STATS_GOOGLE_OAUTH_CLIENT_ID` | Google OAuth Client ID | yes (install) | -W |
| `ARCH_STATS_JWT_SECRET` | JWT signing secret | yes (install) | *(auto-generated)* |
| `ARCH_STATS_JWT_ALGORITHM` | JWT algorithm | no | `HS256` |
| `ARCH_STATS_JWT_TTL_MINUTES` | JWT token TTL | no | `60` |
| `CLOUDFLARED_TUNNEL_ID` | Cloudflare Tunnel UUID | yes (install) | - |

## Platform Assumptions

- **Linux**: Scripts rely on Linux-specific paths (e.g., `/var/run/postgresql`) and commands
  (`systemctl`).
- **Docker**: Expected to be available for running the database in development.
- **Docker Compose**: Expected to be available for running the database in development.
- **Dependencies**: `npm` for Node.js frontend and Go toolchain for the backend are expected to be
  available in the development environment.

## References

- **Backend**: [backend/README.md](file://../backend/README.md)
- **Frontend**: [frontend/README.md](file://../frontend/README.md)
