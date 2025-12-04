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
| :--- | :--- |
| [`install.bash`](./install.bash) | Installs Arch-Stats to the local system (downloads artifact, sets up venv). |
| [`start_uvicorn.bash`](./start_uvicorn.bash) | Starts the backend dev server with hot reload and Docker dependencies. |
| [`linting.bash`](./linting.bash) | All-in-one runner for backend, frontend, and bash linting/testing. |
| [`create_pr.bash`](./create_pr.bash) | Automates PR creation with labels based on changed files. |
| [`remote_installer.bash`](./remote_installer.bash) | Sets up the application and systemd service on a remote Linux server. |
| [`local_installer.bash`](./local_installer.bash) | Orchestrates remote installation/uninstallation via SSH. |

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

- **`scripts/*.bash`**: executable scripts for development and deployment tasks.
- **`scripts/lib/`**: Shared libraries and helper functions (e.g., `manage_docker`).
- **`.github/workflows/`**: CI/CD pipeline definitions.
- **`.github/actions/`**: Local composite actions (`uv-setup`, `npm-setup`).

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
./scripts/linting.bash --lint-scripts
```

This runs **ShellCheck** and **shfmt** to ensure code correctness and consistent style.

## CI/CD Workflows

The repository uses GitHub Actions for continuous integration and deployment.

- **Frontend**: Linting & Formatting ([`frontend_linting.yaml`](../.github/workflows/frontend_linting.yaml))
- **Backend**: Type Check, Linting & Tests ([`backend_linting.yaml`](../.github/workflows/backend_linting.yaml))
- **Scripts**: Bash Linting ([`bash_linting.yaml`](../.github/workflows/bash_linting.yaml))
- **Release**: Build Artifact ([`build_artifact.yaml`](../.github/workflows/build_artifact.yaml))

### Merge Requirements

For a PR to be mergeable, the following workflows must pass if triggered:

| Workflow | Triggers on Changes In | Must Pass |
| :--- | :--- | :--- |
| **Frontend** | `frontend/**` | Formatting, Linting, Tests |
| **Backend** | `backend/**` | Black, Isort, MyPy, Pylint, Tests |
| **Scripts** | `scripts/*.bash` | ShellCheck, shfmt |

> [!NOTE]
> The **Release** workflow runs only on `push` to `main` and is not a PR requirement.

## Environment Variables

Deployment and CI jobs must surface required runtime variables explicitly.

| Variable | Purpose | Required | Default (dev) |
| :--- | :--- | :--- | :--- |
| `PGHOST` | Postgres host | yes | `localhost` |
| `PGPORT` | Postgres port | no | `5432` |
| `PGUSER` | Postgres user | yes | `postgres` |
| `PGPASSWORD` | Postgres password | yes | *(secret)* |
| `PGDATABASE` | DB name | yes | `arch_stats` |
| `POSTGRES_POOL_MIN_SIZE` | Min pool connections | no | `1` |
| `POSTGRES_POOL_MAX_SIZE` | Max pool connections | no | `10` |

## Platform Assumptions

- **Linux**: Scripts rely on Linux-specific paths (e.g., `/var/run/postgresql`) and commands
  (`systemctl`).
- **Docker**: Expected to be available for running the database and tests.
- **Docker Compose**: Expected to be available for running the database and tests.
- **Dependencies**: `uv` for Python and `npm` for Node.js are expected to be managed via the
  provided setup actions/scripts.

## References

- **Backend**: [backend/README.md](../backend/README.md)
- **Frontend**: [frontend/README.md](../frontend/README.md)
