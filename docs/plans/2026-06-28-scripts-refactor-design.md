# Scripts Refactor Design

This design document outlines the plan to refactor the maintenance and deployment scripts. The goal is to replace the existing fragmented/broken scripts with a clean, single-entry-point orchestration layer on the local machine and robust, state-aware installation/uninstallation/update flows on the remote machine.

## Goal

Provide a single, idempotent entry point script `scripts/deploy.bash` on the local machine to manage remote installation, updates, and uninstallation on the Raspberry Pi 5. The orchestrator dynamically checks the remote state to choose between a clean install and a lightweight service update.

---

## User Review Required

> [!IMPORTANT]
> - **Clean Install Requirements**: A local `.env` file containing `GITHUB_TOKEN`, `ARCH_STATS_GOOGLE_OAUTH_CLIENT_ID`, `ARCH_STATS_JWT_SECRET`, and `CLOUDFLARED_TUNNEL_ID` must be configured, and the local credentials file must exist at `${HOME}/.cloudflared/${CLOUDFLARED_TUNNEL_ID}.json`.
> - **Update Flow**: Does not require any environment variables besides `GITHUB_TOKEN`, nor does it require any credentials files or rendering. It operates on the assumption that the credentials and remote configuration files were successfully set up during the clean install.

---

## Proposed Changes

### Local Orchestration & Rendering

#### [NEW] [deploy.bash](file:///home/juanpa/Projects/arch-stats/scripts/deploy.bash)
- Single entry point: `./scripts/deploy.bash <remote-host> [<action>]`.
- If `<action>` is `uninstall`, execute the uninstall sequence.
- If `<action>` is `install` or omitted, execute a pre-flight remote check to determine the target action:
  - Check if the system user `arch-stats` exists, `cloudflared` is installed, `postgresql` is registered as a unit, and `/opt/arch-stats/backend` exists.
  - If all conditions are met, run the **Update Flow**.
  - Otherwise, run the **Clean Install Flow**.
- **Clean Install Flow**:
  1. Load local environment variables.
  2. Render configuration files (`arch-stats.service` and `cloudflared_config.yaml`) locally in `/tmp`.
  3. Generate the remote production `.env` file locally.
  4. Upload all scripts, credentials, services, and configs to `/tmp/` on the remote server.
  5. Run `remote_installer.bash` as root via `ssh`/`sudo`.
- **Update Flow**:
  1. Upload only `install_app.bash` to `/tmp/` on the remote server.
  2. Run the update sequence:
     ```bash
     ssh -t "<remote-host>" "
       sudo systemctl stop arch-stats.service && \
       sudo -u arch-stats GITHUB_TOKEN='${GITHUB_TOKEN}' /tmp/install_app.bash /opt/arch-stats && \
       sudo systemctl start arch-stats.service
     "
     ```

#### [MODIFY] [transform_templates.py](file:///home/juanpa/Projects/arch-stats/scripts/transform_templates.py)
- Add missing parser argument `--cloudflared-tunnel-id` to fix execution crashes.
- Map the variable correctly into the Jinja2 context directory.

#### [DELETE] [local_installer.bash](file:///home/juanpa/Projects/arch-stats/scripts/local_installer.bash)
- Retrospective file to be deleted. Replaced by `deploy.bash`.

---

### Remote Scripts

#### [MODIFY] [remote_installer.bash](file:///home/juanpa/Projects/arch-stats/scripts/remote_installer.bash)
- Script that runs as `root` for the clean installation flow.
- Installs PostgreSQL and configures peer-socket authentication for `arch-stats` user/database.
- Installs Cloudflared, copies credentials and configurations to system folders, and registers/starts the systemd service and auto-update timer.
- Registers the application service `arch-stats.service`.
- Executes `install_app.bash` as `arch-stats` user.
- Deletes temporary installation files under `/tmp`.

#### [MODIFY] [install_app.bash](file:///home/juanpa/Projects/arch-stats/scripts/install_app.bash)
- Script that runs as `arch-stats` user.
- Checks if `uv` is installed under the user path. If missing, installs it via `curl -LsSf https://astral.sh/uv/install.sh | sh`. If already present, runs `uv self update`.
- Downloads the latest release bundle from GitHub, verifies the SHA256 checksum, and extracts it to `/opt/arch-stats/backend`.
- Downloads database migrations from the private repository, extracts them, and applies them to the local database via peer socket authentication.
- Runs `uv sync --no-dev --frozen` to sync python dependencies.

#### [MODIFY] [remote_uninstaller.bash](file:///home/juanpa/Projects/arch-stats/scripts/remote_uninstaller.bash)
- Script that runs as `root` to completely uninstall the application.
- Stops and disables `arch-stats` and `cloudflared` services.
- Deletes system user `arch-stats` along with its home directory (`/opt/arch-stats`).
- Removes registered systemd unit files.
- Drops the PostgreSQL database and user `arch-stats`.
- Purges `cloudflared` packages, apt lists, and keyring configurations.

#### [DELETE] [install_cloudflared.bash](file:///home/juanpa/Projects/arch-stats/scripts/install_cloudflared.bash)
- Retrospective file to be deleted. Merged directly into `remote_installer.bash`.

---

## Verification Plan

### Automated Tests
- Lint all bash scripts using shellcheck and shfmt:
  ```bash
  ```bash
  ./scripts/linting.bash --scripts
  ```

### Manual Verification
- Dry-run verification on local loopback or a staging server to ensure the CLI option handling, template compilation, and SCP distribution run smoothly.
