# Arch-Stats DevOps Scripts

*Who this is for:* DevOps engineers maintaining Arch-Stats CI/CD, deployments, and releases.

- [Arch-Stats DevOps Scripts](#arch-stats-devops-scripts)
  - [Overview](#overview)
  - [Scripts Overview](#scripts-overview)
  - [CI/CD Workflows (.github/workflows)](#cicd-workflows-githubworkflows)
  - [VS Code Integration](#vs-code-integration)
  - [Platform Assumptions](#platform-assumptions)

## Overview

This directory contains Bash scripts and related tools to automate setup, installation, and workflow tasks for Arch-Stats. All scripts assume a **Linux environment** (they are tested on Linux and in CI containers; Windows users should use WSL or a Linux VM). They also rely on Docker for certain tasks (like running a Postgres database for development and testing).

## Scripts Overview

- [`install.bash`](./install.bash): **Install Arch-Stats to local system**. Downloads the latest build artifact and sets up the backend in a Python virtual environment. It creates a `backend/.env` file with required environment variables. **Usage:** `install.bash <db_user> <db_name> [db_socket_dir] [server_port]`. You must provide a Postgres role and database name; optionally a socket directory (default `/var/run/postgresql`) and server port (default `8000`). The script will:
  - Fetch the `arch-stats.tar.xz` release from GitHub (assuming a release tagged "latest").
  - Extract it into the current user's home directory (placing the code in `~/backend`, etc.).
  - Install Python dependencies using the **uv** tool (a CLI for managing virtualenv + pip, installed via a one-line script inside).
  - Create a `backend/.env` file populated with environment variables needed to run the server (database credentials, port, etc.), with secure permissions.
- [`remote_installer.bash`](./remote_installer.bash): **Remote installation and service setup**. Intended to run on a target Linux machine (as root). It creates a dedicated system user (`archy` by default), then uses `install.bash` (as that user) to set up the application, and registers a systemd service. **Usage:** `remote_installer.bash <db_user> <db_name> [socket_dir] [server_port]` (same parameters passed through to install.bash). After running, it writes a systemd service file at `/etc/systemd/system/arch-stats.service` configured to run the FastAPI server via Uvicorn in the background on boot, and starts/enables that service. This is used for deploying Arch-Stats onto a server (for example, a Raspberry Pi running Linux). It expects to be run with root privileges.
- [`remote_uninstaller.bash`](./remote_uninstaller.bash): **Remote removal script**. The counterpart to the installer, this script (run on the target as root) will attempt to stop services and remove the `archy` user and its home directory, cleaning up an Arch-Stats installation. *(Note: Stopping the running service is not fully implemented in the script as of now, but it does remove the system user and any downloaded artifact.)*
- [`local_installer.bash`](./local_installer.bash): **Local orchestrator for remote install/uninstall.** This is a convenience script to run an install or uninstall on a remote host via SSH. **Usage:** `local_installer.bash <remote-host> <action>` where `<action>` is `install` or `uninstall`. It will:
  - Check SSH connectivity to `<remote-host>` (which should be configured, e.g. an entry in `~/.ssh/config` or a reachable hostname).
  - Upload the required script(s) to the remote (for install: `remote_installer.bash` and `install.bash`; for uninstall: `remote_uninstaller.bash`).
  - Execute the remote script on the target via SSH.
This wraps the entire remote deployment process into a single command for developer convenience.
- [`create_pr.bash`](./create_pr.bash): **Pull Request automation**. Developers can run this script after pushing a feature branch to quickly open a GitHub PR with proper labels and links. It uses the GitHub CLI (`gh`) to create a PR from the current branch. The script automatically determines which parts of the code were changed and adds corresponding labels (e.g., changes under `frontend/` get a "frontend" label, changes in `backend/src/server/` get a "server" label, etc.). It also assigns the PR to the "Arch Stats" project and a preset milestone, and if no specific labels apply, it labels it as an enhancement by default. **Usage:** simply run `./scripts/create_pr.bash` from a feature branch. The script will safely exit without doing anything if run on the `main` branch or if a PR for that branch already exists.
- [`linting.bash`](./linting.bash): **All-in-one pre-commit linter/test runner**. This script can be used to run all linting and formatting checks (and backend tests) before committing or in CI. It will:
  - Detect which parts of the repo have changes staged (frontend, backend, or scripts).
  - For frontend changes: run `npm run lint` (ESLint) and `npm run format` (Prettier) in the `frontend/` directory.
  - For backend changes: activate the Python venv and run code formatters (`isort`, `black`), static type check (`mypy`), lint (`pylint`), **and tests** (`pytest`). It uses a helper to ensure the Docker-based Postgres test database is running (see below).
  - For script changes: run ShellCheck and shfmt to lint/format Bash scripts.
  - Return a non-zero exit code if any check fails, so it can block a commit or CI if issues are found.
- [`start_uvicorn.bash`](./start_uvicorn.bash): **Launch the backend dev server**. This is used during development to start the FastAPI/Uvicorn server with live reload. It will spin up the Docker services (ensuring the Postgres database from [`docker/docker-compose.yaml`](../docker/docker-compose.yaml) is running), then activate the Python virtualenv and launch Uvicorn with the app factory (`server.app:run`) in debug mode (with `--reload` for code changes). The script sets Uvicorn to use uvloop, websockets, and other production-similar settings but in debug log level. (This script is invoked by the VS Code task **Start Uvicorn Dev Server**.)
- `start_archy_bot.bash`: **Simulated data feeder** (Archery bot). *If present*, this script would start a small bot process that simulates an archer shooting arrows, feeding test data to the system (likely via WebSocket or API). This is referenced in the VS Code tasks as **Start Archy bot**. (Check `scripts/` for this file; ensure the backend and database are running before using it.) The Archy bot is for development/demo purposes to generate dummy shooting data.

**Docker & Database:** Many scripts use Docker to manage the Postgres database needed by the backend:

- The Docker Compose file ([`docker-compose.yaml`](../docker/docker-compose.yaml)) defines a **postgres:15** database container. The `scripts/lib/manage_docker` helper provides `start_docker` and `stop_docker` Bash functions to manage this container. For example, when you run tests or start the dev server, the scripts will bring up the DB container automatically (if not already running) and tear it down when done (in tests).
- The Compose config mounts the host's `/var/run/postgresql` socket directory into the container so that the database can be accessed via a UNIX domain socket on the host. This is a Linux-specific strategy (for performance and avoiding TCP setup); on Linux/WSL it "just works." If you run the backend on a different OS without WSL, you may need to adjust the `POSTGRES_HOST` to `localhost` and use TCP (`localhost:5432`).

## CI/CD Workflows ([.github/workflows](../.github/workflows/))

The repository uses GitHub Actions to enforce code quality and build releases. Key workflows include:

- **Frontend Linting & Formatting** - ([`frontend_linting.yaml`](../../.github/workflows/frontend_linting.yaml)): Runs on pull requests when files in `frontend/**` change. It sets up Node (via our composite action, see below) and then runs Prettier and ESLint in check mode (using `npm run format -- --check` and `npm run lint`). This ensures any front-end code meets our formatting and linting standards before merging.
- **Backend Type Check & Linting** - ([`backend_linting.yaml`](../../.github/workflows/backend_linting.yaml)): Runs on pull requests for changes under `backend/**`. It sets up the Python environment (via our `uv-setup` action) then runs Black, Isort, MyPy, and Pylint on the backend code (all configured via `backend/pyproject.toml`) to ensure style and type correctness. Each tool is run in a separate job for clarity.
- **Bash Linting & Formatting** - ([`bash_linting.yaml`](../../.github/workflows/bash_linting.yaml)): Runs on pull requests that include changes to any `scripts/*.bash` files. It installs ShellCheck and shfmt on an Ubuntu runner and checks that all shell scripts are linted (no warnings/errors) and properly formatted. This workflow prevents bad bash syntax or style from being merged.
- **Build Artifact (Release)** - ([`build_artifact.yaml`](../../.github/workflows/build_artifact.yaml)): Triggers on every push to `main`. This is the continuous deployment build. It performs a full project build:
  - Checks out the code and installs both backend and frontend dependencies (using local composite actions: `uv-setup` for Python, `npm-setup` for Node).
  - Generates the OpenAPI spec by running the FastAPI app (`uv run scripts/generate_openapi.py`) and then runs `openapi-typescript` to produce the latest `types.generated.ts` for the frontend.
  - Builds the frontend (production build) and packages the entire backend (including the new frontend assets) into a tarball (`arch-stats.tar.xz`).
  - Uploads this artifact to a GitHub Release tagged "latest", effectively updating the deployable package that `install.bash` will download.
  
These workflows use local composite actions defined under `.github/actions/`:

- **uv-setup**: Installs the **uv** tool and uses it to create the Python virtualenv and install backend dependencies (as defined in `pyproject.toml`). This abstracts the environment setup so all workflows use a consistent Python setup.
- **npm-setup**: Sets up Node.js (using actions/setup-node under the hood) and installs `frontend/package.json` dependencies. This ensures the correct Node version and packages are present for linting or building.

**Extending CI:** If you add new languages or directories, it's important to extend the CI to cover them. The current setup filters workflows by path:

- Frontend changes trigger only frontend linting, backend changes trigger backend linting, etc. This keeps CI efficient. If your change spans multiple areas, multiple workflows run in parallel.
- To **add tests to CI** (e.g., run `pytest` on pull requests), you could create a new workflow or integrate it into the backend workflow. For example, you might add a job in `backend_linting.yaml` or a separate `backend_tests.yaml` triggered on the same path that runs `pytest` using the Postgres service (similar to how `scripts/linting.bash` does locally with Docker). Ensure you use the `uv-setup` action and perhaps add a service in the workflow for Postgres.
- When adding new script files or changing their extension, update the path filter in **Bash Linting** workflow (currently it watches `scripts/*.bash`). Consistency in naming (using `.bash` extension for shell scripts) helps keep this simple.
- Always test changes to workflows on a branch (in a fork or with a dummy PR) before relying on them in `main`, to avoid breaking the CI for everyone.

## VS Code Integration

For convenience, the repository provides a [.vscode/tasks.json](../.vscode/tasks.json) with common tasks:

- **Start Docker Compose** / **Stop Docker Compose**: to manage the dev database container.
- **Start Uvicorn Dev Server**: runs `scripts/start_uvicorn.bash` (as described above) to launch the backend.
- **Start Vite Frontend**: runs `npm run dev` for the frontend.
- **Start Archy bot**: runs the archery simulation bot (if needed for testing).
These tasks can be run via the VS Code interface (Tasks: Run Task) and are set up to appear in the IDE's debugger/run menu as well. They streamline the development workflow by ensuring you don't forget to start dependencies (the tasks handle Docker etc.).

## Platform Assumptions

All scripts and automation assume a **Unix-like environment**:

- **Linux or WSL:** The installation and management scripts use commands like `useradd`, `systemctl`, and paths like `/var/run/postgresql` which are available on Linux. For Windows development, use WSL2 to host Docker and a Linux environment.
- **Docker:** Ensure Docker is installed and the daemon is running if you intend to run tests or the dev server via the provided scripts. The `manage_docker` utility will call `docker compose up/down` as needed. The CI expects Docker to be available (the GitHub runners provide it by default).
- **Python & Node Versions:** The backend uses Python (check `backend/.python-version` for the exact version, e.g., Python 3.11.x) and the frontend uses Node (see `frontend/package.json` for engine recommendations). The `uv` tool will auto-install the correct Python and create an isolated environment. For Node, use an LTS release; minor version differences in npm shouldn't affect scripts, but using the same version as in CI (as set in `npm-setup` action or `.nvmrc` if provided) is recommended.

By using these scripts and workflows, you can reliably deploy Arch-Stats and ensure code quality. Always keep the scripts updated with any configuration changes (e.g., new environment variables or dependencies) and test deployment changes in a staging environment when possible. The separation of concerns (scripts for env setup, CI for quality checks, tasks for dev workflow) is designed to make maintaining and extending the DevOps pipeline straightforward and predictable.
