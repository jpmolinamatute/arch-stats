# Task 034: Update Deployment Scripts + Systemd Service

## Git Branch

`refactor/034-deployment-scripts-and-systemd`

## Objective

Update the deployment scripts and systemd service file for the new single-binary Go deployment
model. The deployment simplifies drastically: download the binary, verify checksum, place it,
and run migrations — no venv, no `uv sync`, no tarball extraction.

Additionally, the deployment and installation scripts must be fully tested and verified against
the local Raspberry Pi emulator container configured in
[docker/docker-compose.yaml](file:///home/juanpa/Projects/arch-stats/docker/docker-compose.yaml) (profile `emulator`)
and built from [docker/Dockerfile.rpi](file:///home/juanpa/Projects/arch-stats/docker/Dockerfile.rpi).
The entire deployment and service startup lifecycle must execute cleanly end-to-end in the emulator environment.

## Dependencies

- Task 033 (CI produces the single Go binary as a release)
- Task 005 (Go binary can run migrations via `arch-stats migrate`)
- [docker/docker-compose.yaml](file:///home/juanpa/Projects/arch-stats/docker/docker-compose.yaml) (`emulator` profile service definition)
- [docker/Dockerfile.rpi](file:///home/juanpa/Projects/arch-stats/docker/Dockerfile.rpi) (Debian 12 + systemd + SSH emulator environment)

## Acceptance Criteria

- [x] `scripts/install_app.bash` is simplified:
    - Downloads the Go binary from GitHub Releases (not a tarball)
    - Verifies checksum
    - Places binary at `/opt/arch-stats/arch-stats`
    - Sets executable permissions
    - Runs migrations: `/opt/arch-stats/arch-stats migrate`
    - No venv creation, no `uv sync`, no pip
- [x] `scripts/remote_installer.bash` is updated:
    - Same stop/start flow but simpler internals
    - References the binary, not the venv
- [x] `scripts/deploy.bash` is updated for the new artifact type.
- [x] Systemd service file (referenced in scripts or `scripts/templates/arch-stats.service.j2`) is updated:
    - `ExecStart=/opt/arch-stats/arch-stats` (single binary)
    - No `WorkingDirectory` pointing to venv
    - Environment variables loaded from `/opt/arch-stats/.env`
- [x] `scripts/start_uvicorn.bash` is removed or replaced with a Go equivalent.
- [x] All scripts pass `shellcheck`:

    ```bash
    shellcheck scripts/install_app.bash scripts/remote_installer.bash scripts/deploy.bash
    ```
- [x] The deployment script fully runs and passes end-to-end testing in the local emulator:
    - Emulator container starts cleanly via `docker compose -f docker/docker-compose.yaml --profile emulator up -d --build`
    - SSH connection is established using `docker/ssh/arch_stats_dev` key on port 2222
    - The deployment / installer script executes to completion without errors inside the emulator
    - `arch-stats.service` systemd unit is active (`systemctl is-active arch-stats.service` returns `active`)
    - Embedded database migrations run successfully during installation
    - The server responds to HTTP requests inside the emulator

## Files to Modify/Delete

| Action | Path |
| ------ | ---- |
| Modify | `scripts/install_app.bash` |
| Modify | `scripts/remote_installer.bash` |
| Modify | `scripts/deploy.bash` |
| Delete | `scripts/start_uvicorn.bash` |
| Modify | `scripts/templates/arch-stats.service.j2` |
| Verify | `docker/docker-compose.yaml` |
| Verify | `docker/Dockerfile.rpi` |

## Reference

- Current installer: [install_app.bash](file:///home/juanpa/Projects/arch-stats/scripts/install_app.bash)
- Current remote installer: [remote_installer.bash](file:///home/juanpa/Projects/arch-stats/scripts/remote_installer.bash)
- Current deploy: [deploy.bash](file:///home/juanpa/Projects/arch-stats/scripts/deploy.bash)
- Emulator compose: [docker-compose.yaml](file:///home/juanpa/Projects/arch-stats/docker/docker-compose.yaml)
- Emulator Dockerfile: [Dockerfile.rpi](file:///home/juanpa/Projects/arch-stats/docker/Dockerfile.rpi)
- Emulator design doc: [2026-07-02-raspberry-pi-emulator-design.md](file:///home/juanpa/Projects/arch-stats/docs/plans/2026-07-02-raspberry-pi-emulator-design.md)
- Plan §12: single binary deployment model

## Steps

- [x] **Step 1: Update `install_app.bash`**

  Simplify to:
  1. Check if running as root
  2. Download binary from GitHub Releases
  3. Verify SHA-256 checksum
  4. Stop service if running
  5. Place binary at `/opt/arch-stats/arch-stats`
  6. `chmod +x /opt/arch-stats/arch-stats`
  7. Run migrations: `/opt/arch-stats/arch-stats migrate`
  8. Start/restart service

  Remove all Python/venv/uv references.

- [x] **Step 2: Update `remote_installer.bash`**

  Update to reference the binary instead of the tarball. Same SSH + stop/start flow.

- [x] **Step 3: Update `deploy.bash`**

  Update to download the binary (not tarball) and deploy.

- [x] **Step 4: Delete `scripts/start_uvicorn.bash`**

  ```bash
  git rm scripts/start_uvicorn.bash
  ```

- [x] **Step 5: Update systemd service template**

  ```ini
  [Unit]
  Description=Arch Stats Server
  After=network.target postgresql.service
  Requires=postgresql.service

  [Service]
  Type=simple
  User=arch-stats
  Group=arch-stats
  ExecStart=/opt/arch-stats/arch-stats
  EnvironmentFile=/opt/arch-stats/.env
  Restart=on-failure
  RestartSec=5

  [Install]
  WantedBy=multi-user.target
  ```

- [x] **Step 6: Run shellcheck**

  ```bash
  shellcheck scripts/install_app.bash scripts/remote_installer.bash scripts/deploy.bash
  ```

- [x] **Step 7: Verify no Python/venv references remain**

  ```bash
  grep -rn "venv\|uvicorn\|uv sync\|pip\|python" scripts/install_app.bash scripts/deploy.bash
  ```

  Expected: no results.

- [x] **Step 8: Test deployment against Raspberry Pi Docker emulator**

  1. Start the emulator container:
     ```bash
     docker compose -f docker/docker-compose.yaml --profile emulator up -d --build
     ```
  2. Verify SSH connectivity to the emulator:
     ```bash
     ssh -i docker/ssh/arch_stats_dev -p 2222 -o StrictHostKeyChecking=no root@localhost "systemctl is-system-running --wait || true"
     ```
  3. Run the deployment/installation script against the emulator environment (target `root@localhost:2222`).
  4. Verify service status and application health inside the emulator:
     ```bash
     ssh -i docker/ssh/arch_stats_dev -p 2222 root@localhost "systemctl status arch-stats.service"
     ssh -i docker/ssh/arch_stats_dev -p 2222 root@localhost "curl -s http://localhost:8001/api/v0/health"
     ```
  5. Tear down emulator container after verification:
     ```bash
     docker compose -f docker/docker-compose.yaml --profile emulator down
     ```

- [x] **Step 9: Commit**

  ```bash
  git add -A
  git commit -m "chore: update deployment scripts for Go single binary model and verify with emulator"
  ```

## Verification

- `shellcheck scripts/install_app.bash scripts/remote_installer.bash scripts/deploy.bash` — clean.
- `grep -rn "venv\|uvicorn\|uv sync" scripts/` — no Python references in deployment scripts.
- `scripts/start_uvicorn.bash` no longer exists.
- Systemd service references `/opt/arch-stats/arch-stats`.
- `docker compose -f docker/docker-compose.yaml --profile emulator up -d --build` builds and starts emulator container.
- Deployment script runs to completion without errors against the emulator target.
- `ssh -i docker/ssh/arch_stats_dev -p 2222 root@localhost "systemctl is-active arch-stats.service"` returns `active`.
- `ssh -i docker/ssh/arch_stats_dev -p 2222 root@localhost "curl -f http://localhost:8001/api/v0/health"` returns HTTP 200.
