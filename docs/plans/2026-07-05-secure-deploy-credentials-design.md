# Secure Deploy Credentials Design

**Topic:** Secure transmission and storage of deployment credentials (like `GITHUB_TOKEN`) during remote installations and updates.

## Problem Statement
The current deployment process passes the `GITHUB_TOKEN` directly via command-line arguments on the remote host (e.g., `sudo GITHUB_TOKEN='...'` and `sudo -u arch-stats GITHUB_TOKEN='...'`). This exposes sensitive tokens to any local user running process monitoring tools like `ps` or accessing `/proc`.

## Proposed Design (Approach 1: Secure Temporary Directory)

We will upload all deployment files (including a temporary environment file containing `GITHUB_TOKEN`) to a secure remote directory, configure correct permissions, source the credentials within the scripts, and clean up securely on exit/error.

### 1. Secure Remote Directory Creation
Instead of uploading files directly to `/tmp/`, `deploy.bash` will:
- Create a directory `/tmp/deploy_assets` on the remote host.
- Restrict its permissions to `700` (read, write, and execute permissions restricted to the SSH user).

### 2. Unified Asset Uploads
All configuration files, service units, scripts, credentials, and a temporary `env` file containing `export GITHUB_TOKEN='...'` will be uploaded to `/tmp/deploy_assets/`.

### 3. Remote Installer Adaptations (`remote_installer.bash`)
- Deduce configuration paths dynamically relative to `ROOT_DIR` (which resolves to `/tmp/deploy_assets`) instead of hardcoding `/tmp/`.
- Register an `EXIT` trap to ensure the `/tmp/deploy_assets` directory is recursively deleted upon script completion or premature exit (on errors).

### 4. Deploy Actions Execution
- **Install Flow:**
  - SSH to host and execute:
    ```bash
    sudo bash -c 'source /tmp/deploy_assets/env && rm -f /tmp/deploy_assets/env && bash /tmp/deploy_assets/remote_installer.bash "arch-stats"'
    ```
- **Update Flow:**
  - Change ownership of `/tmp/deploy_assets` to the `arch-stats` user.
  - Run the update script as the `arch-stats` user, then clean up the directory unconditionally:
    ```bash
    sudo chown -R arch-stats:arch-stats /tmp/deploy_assets && \
    sudo systemctl stop arch-stats.service && \
    (
        sudo -u arch-stats bash -c '
            source /tmp/deploy_assets/env
            bash /tmp/deploy_assets/install_app.bash /opt/arch-stats
        ' && \
        sudo systemctl start arch-stats.service
    )
    ec=$?
    sudo rm -rf /tmp/deploy_assets
    exit $ec
    ```

## Alternatives Considered
- **Direct Command Line Injection:** Rejected due to `ps` command line exposure.
- **Named Pipes (FIFOs):** Rejected due to operational complexity and the risk of hanging.

## Verification
- Validate the scripts using `shellcheck` and `shfmt` via `./scripts/linting.bash --scripts`.
