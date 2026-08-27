# Scripts Refactor Implementation Plan

> **For Antigravity:** REQUIRED WORKFLOW: Use `.agent/workflows/execute-plan.md` to execute this plan in single-flow mode.

**Goal:** Refactor local and remote deployment/installation scripts to provide a single entry point `scripts/deploy.bash` that manages automated state-aware clean installs, service updates, and uninstallation.

**Architecture:** 
- The local script `deploy.bash` performs a pre-flight remote check to choose between clean install, update, or uninstall flows.
- Clean install renders config templates via the fixed `transform_templates.py`, generates a remote `.env`, uploads files, and runs `remote_installer.bash` as root.
- Service updates skip rendering and upload only `install_app.bash`, running it as user `arch-stats` under a service stop/start window.
- Uninstallation runs `remote_uninstaller.bash` as root to clean up all users, directories, services, and PostgreSQL configuration.

**Tech Stack:** Bash, Python 3.14, Jinja2, systemd, PostgreSQL, cloudflared, SSH/SCP.

---

### Task 1: Fix transform_templates.py Parser

**Files:**
- Modify: `scripts/transform_templates.py`

**Step 1: Write template parser updates**
Add the `--cloudflared-tunnel-id` argument to the argparse parser in `scripts/transform_templates.py` so it does not crash when rendering configurations:
```python
    parser.add_argument(
        "--cloudflared-tunnel-id", required=True, help="Cloudflared Tunnel ID"
    )
```

**Step 2: Run verification**
Run the script locally to render a test service and YAML config:
`uv run scripts/transform_templates.py --app-name arch-stats --app-name-label "Arch Stats" --prod-uvicorn-port 8000 --app-user-home-dir /opt/arch-stats --output-dir /tmp --cloudflared-tunnel-id 1234-abcd`
Expected: Files `/tmp/arch-stats.service` and `/tmp/cloudflared_config.yaml` are generated successfully.

**Step 3: Commit**
`git add scripts/transform_templates.py`

---

### Task 2: Delete Deprecated Installer Scripts

**Files:**
- Delete: `scripts/local_installer.bash`
- Delete: `scripts/install_cloudflared.bash`

**Step 1: Remove files**
Delete the old, deprecated scripts:
`rm -f scripts/local_installer.bash scripts/install_cloudflared.bash`

**Step 2: Verify deletion**
Run `ls scripts/` and ensure neither file remains.

**Step 3: Commit**
`git rm scripts/local_installer.bash scripts/install_cloudflared.bash`

---

### Task 3: Implement remote_uninstaller.bash

**Files:**
- Modify: `scripts/remote_uninstaller.bash`

**Step 1: Write remote uninstaller logic**
Rewrite `scripts/remote_uninstaller.bash` to stop and disable all systemd services, remove systemd configurations, drop PostgreSQL user/DB, purge the cloudflared package, and delete the system user and directories:
```bash
#!/usr/bin/env bash

set -euo pipefail

log_info() { echo "INFO: $*"; }
log_error() { echo "ERROR: $*" >&2; }

main() {
    if [[ $EUID -ne 0 ]]; then
        log_error "Please run as root."
        exit 1
    fi

    log_info "Stopping and disabling services..."
    for svc in arch-stats.service cloudflared.service cloudflared-update.timer cloudflared-update.service; do
        if systemctl is-active --quiet "$svc"; then
            systemctl stop "$svc" || true
        fi
        if systemctl is-enabled --quiet "$svc" 2>/dev/null; then
            systemctl disable "$svc" || true
        fi
    done

    log_info "Removing systemd unit files..."
    rm -f /etc/systemd/system/arch-stats.service \
          /etc/systemd/system/cloudflared.service \
          /etc/systemd/system/cloudflared-update.service \
          /etc/systemd/system/cloudflared-update.timer
    systemctl daemon-reload

    log_info "Dropping PostgreSQL database and user..."
    if command -v dropdb >/dev/null 2>&1; then
        sudo -u postgres dropdb --if-exists arch-stats || true
    fi
    if command -v dropuser >/dev/null 2>&1; then
        sudo -u postgres dropuser --if-exists arch-stats || true
    fi

    log_info "Purging cloudflared package..."
    if command -v cloudflared >/dev/null 2>&1; then
        apt-get purge -y cloudflared || true
    fi
    rm -rf /etc/cloudflared /usr/share/keyrings/cloudflare-main.gpg /etc/apt/sources.list.d/cloudflared.list

    log_info "Deleting user arch-stats and home directory..."
    if id -u arch-stats >/dev/null 2>&1; then
        userdel -r arch-stats || true
    fi

    log_info "Uninstallation completed successfully."
    exit 0
}

main "$@"
```

**Step 2: Commit**
`git add scripts/remote_uninstaller.bash`

---

### Task 4: Implement install_app.bash

**Files:**
- Modify: `scripts/install_app.bash`

**Step 1: Add uv installation/updater checks**
Update `scripts/install_app.bash` to check if `uv` is installed. If not, install it via the official installation script. Otherwise, run `uv self update`:
```bash
# In scripts/install_app.bash:
install_dependencies() {
    local app_home_dir="${1}/backend"

    # Add uv path to path variable
    export PATH="${HOME}/.local/bin:${PATH}"

    if ! command -v uv >/dev/null 2>&1; then
        log_info "uv not found, installing uv..."
        curl -LsSf https://astral.sh/uv/install.sh | sh
    else
        log_info "Running 'uv self update'"
        if ! uv self update; then
            log_error "uv self update failed"
            exit 5
        fi
    fi

    cd "$app_home_dir" || {
        log_error "backend directory not found: $app_home_dir"
        exit 4
    }

    log_info "Syncing production dependencies (no dev, frozen)"
    if ! uv sync --no-dev --frozen --python "$(cat .python-version)"; then
        log_error "uv sync failed"
        exit 6
    fi

    log_info "Dependency installation completed successfully"
}
```

**Step 2: Commit**
`git add scripts/install_app.bash`

---

### Task 5: Implement remote_installer.bash

**Files:**
- Modify: `scripts/remote_installer.bash`

**Step 1: Write clean installer logic**
Rewrite `scripts/remote_installer.bash` to create the system user `arch-stats`, set up PostgreSQL (install + peer socket access), install and register Cloudflared, register the systemd service, and call `install_app.bash`:
```bash
#!/usr/bin/env bash

set -Eeuo pipefail

APP="arch-stats"
SYSTEM_SERVICE="${APP}.service"
ROOT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd -P)"

log_info() { echo "INFO: $*"; }
log_error() { echo "ERROR: $*" >&2; }

setup_app_user() {
    if ! id -u "${APP}" >/dev/null 2>&1; then
        log_info "Creating system user '${APP}'..."
        useradd -r -m -d "/opt/${APP}" -s "/usr/sbin/nologin" "${APP}"
    fi
    local user_dir
    user_dir="$(getent passwd "$APP" | cut -d: -f6)"
    
    if [[ -f "/tmp/.env" ]]; then
        mv "/tmp/.env" "${user_dir}/.env"
        chown "${APP}:${APP}" "${user_dir}/.env"
        chmod 600 "${user_dir}/.env"
    fi
}

setup_postgres() {
    if ! systemctl list-units --type=service | grep -q "postgresql"; then
        log_info "Installing PostgreSQL..."
        apt-get update -y && apt-get install -y postgresql postgresql-contrib
    fi

    log_info "Configuring PostgreSQL user and database..."
    # Ensure service is running
    systemctl start postgresql
    
    if ! sudo -u postgres psql -t -c '\du' | cut -d \| -f 1 | grep -q "${APP}"; then
        sudo -u postgres createuser "${APP}"
    fi

    if ! sudo -u postgres psql -lqt | cut -d \| -f 1 | grep -qw "${APP}"; then
        sudo -u postgres createdb -O "${APP}" "${APP}"
    fi
}

setup_cloudflared() {
    local cred_file
    cred_file="$(find /tmp -maxdepth 1 -name "*.json" | head -n 1)"
    if [[ ! -f "${cred_file}" ]]; then
        log_error "Cloudflared credentials file not found in /tmp. Aborting."
        exit 23
    fi

    local tunnel_id
    tunnel_id="$(basename "${cred_file}" .json)"

    if ! command -v cloudflared >/dev/null 2>&1; then
        log_info "Installing cloudflared..."
        curl -fsSL https://pkg.cloudflare.com/cloudflare-main.gpg | tee /usr/share/keyrings/cloudflare-main.gpg >/dev/null
        local codename="bookworm"
        if [[ -f /etc/os-release ]]; then
            codename="$(. /etc/os-release && echo "${VERSION_CODENAME:-bookworm}")"
        fi
        echo "deb [signed-by=/usr/share/keyrings/cloudflare-main.gpg] https://pkg.cloudflare.com/cloudflared ${codename} main" | tee /etc/apt/sources.list.d/cloudflared.list >/dev/null
        apt-get update -y && apt-get install -y cloudflared
    fi

    local user_dir
    user_dir="$(getent passwd "$APP" | cut -d: -f6)"
    local cf_dir="${user_dir}/.cloudflared"
    mkdir -p "${cf_dir}"
    chmod 700 "${cf_dir}"
    cp "${cred_file}" "${cf_dir}/${tunnel_id}.json"
    chmod 600 "${cf_dir}/${tunnel_id}.json"
    chown -R "${APP}:${APP}" "${cf_dir}"

    mkdir -p "/etc/cloudflared"
    cp "/tmp/cloudflared_config.yaml" "/etc/cloudflared/config.yml"
    chmod 644 "/etc/cloudflared/config.yml"

    cp /tmp/cloudflared.service /tmp/cloudflared-update.service /tmp/cloudflared-update.timer /etc/systemd/system/
    chmod 644 /etc/systemd/system/cloudflared.service /etc/systemd/system/cloudflared-update.service /etc/systemd/system/cloudflared-update.timer

    systemctl daemon-reload
    systemctl enable cloudflared.service cloudflared-update.timer
    systemctl restart cloudflared.service cloudflared-update.timer
}

register_app_service() {
    cp "/tmp/arch-stats.service" "/etc/systemd/system/"
    chmod 644 "/etc/systemd/system/arch-stats.service"
    systemctl daemon-reload
    systemctl enable arch-stats.service
}

install_app_as_user() {
    local user_dir
    user_dir="$(getent passwd "$APP" | cut -d: -f6)"
    local script_path="${ROOT_DIR}/install_app.bash"
    chmod +x "$script_path"
    log_info "Running application installer as ${APP}..."
    if ! runuser -u "${APP}" -- "${script_path}" "${user_dir}"; then
        log_error "Application installation failed."
        exit 14
    fi
}

main() {
    if [[ $EUID -ne 0 ]]; then
        log_error "Please run as root."
        exit 1
    fi

    setup_app_user
    setup_postgres
    setup_cloudflared
    register_app_service
    install_app_as_user
    
    log_info "Starting arch-stats service..."
    systemctl start arch-stats.service
    
    if ! systemctl is-active --quiet arch-stats.service; then
        log_error "Service arch-stats.service failed to start."
        exit 22
    fi

    log_info "Cleaning up temporary installation files..."
    rm -f /tmp/remote_installer.bash /tmp/install_app.bash \
          /tmp/cloudflared_config.yaml /tmp/arch-stats.service \
          /tmp/cloudflared.service /tmp/cloudflared-update.service /tmp/cloudflared-update.timer \
          /tmp/*.json

    log_info "Remote installation completed successfully."
    exit 0
}

main "$@"
```

**Step 2: Commit**
`git add scripts/remote_installer.bash`

---

### Task 6: Implement deploy.bash

**Files:**
- Create: `scripts/deploy.bash`

**Step 1: Write local orchestrator script**
Write `scripts/deploy.bash` to parse arguments, execute remote capability checks, compile configurations, distribute files, and trigger the remote scripts:
```bash
#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")"

usage() {
    cat <<EOF
Usage: $(basename "${0}") <remote-host> [<action>]
Description: Deploy or uninstall arch-stats remotely.

Arguments:
  <remote-host>   SSH host target (required)
  [<action>]      Option: "install" or "uninstall". Defaults to auto-resolving install/update.
EOF
    exit 1
}

# Pre-flight checklist to verify remote state
check_remote_action() {
    local host="${1}"
    if ssh -q -o BatchMode=yes -o ConnectTimeout=5 "${host}" "
      id -u arch-stats >/dev/null 2>&1 && \
      command -v cloudflared >/dev/null 2>&1 && \
      systemctl list-units --type=service | grep -q 'postgresql' && \
      [ -d '/opt/arch-stats/backend' ]
    "; then
        echo "update"
    else
        echo "install"
    fi
}

main() {
    if [[ $# -lt 1 ]]; then
        usage
    fi

    local host="${1}"
    local action="${2:-}"

    # Load environment variables
    if [[ -f "${SCRIPT_DIR}/.env" ]]; then
        # shellcheck source=/dev/null
        source "${SCRIPT_DIR}/.env"
    fi

    # Check SSH connectivity
    if ! ssh -o BatchMode=yes -o ConnectTimeout=5 "${host}" exit 2>/dev/null; then
        echo "ERROR: Cannot connect to remote host: ${host}" >&2
        exit 1
    fi

    # Resolve action
    if [[ -z "${action}" ]]; then
        action=$(check_remote_action "${host}")
        echo "Resolved deployment action: ${action}"
    elif [[ "${action}" == "install" ]]; then
        action=$(check_remote_action "${host}")
        echo "Resolved install parameter to: ${action}"
    elif [[ "${action}" != "uninstall" ]]; then
        echo "ERROR: Invalid action: ${action}" >&2
        usage
    fi

    if [[ "${action}" == "uninstall" ]]; then
        echo "Starting remote uninstallation on '${host}'..."
        scp -o BatchMode=yes "${SCRIPT_DIR}/remote_uninstaller.bash" "${host}":/tmp/
        ssh -t "${host}" "sudo chmod +x /tmp/remote_uninstaller.bash && sudo /tmp/remote_uninstaller.bash"
        exit 0
    fi

    # Require GITHUB_TOKEN for install/update
    : "${GITHUB_TOKEN:?Environment variable GITHUB_TOKEN must be set}"

    if [[ "${action}" == "install" ]]; then
        echo "Starting clean installation on '${host}'..."
        
        # Load sensitive parameters
        : "${ARCH_STATS_GOOGLE_OAUTH_CLIENT_ID:?Google OAuth Client ID is required for clean install}"
        : "${ARCH_STATS_JWT_SECRET:?JWT Secret is required for clean install}"
        : "${CLOUDFLARED_TUNNEL_ID:?Cloudflared Tunnel ID is required for clean install}"

        local cred_file="${HOME}/.cloudflared/${CLOUDFLARED_TUNNEL_ID}.json"
        if [[ ! -f "${cred_file}" ]]; then
            echo "ERROR: Local Cloudflared credential file not found at: ${cred_file}" >&2
            exit 1
        fi

        # Create temporary working dir
        local temp_dir
        temp_dir="$(mktemp -d)"

        # Render configurations
        uv run "${SCRIPT_DIR}/transform_templates.py" \
            --app-name "arch-stats" \
            --app-name-label "Arch Stats" \
            --prod-uvicorn-port 8000 \
            --app-user-home-dir "/opt/arch-stats" \
            --output-dir "${temp_dir}" \
            --cloudflared-tunnel-id "${CLOUDFLARED_TUNNEL_ID}"

        # Generate remote .env
        cat <<EOF > "${temp_dir}/.env"
POSTGRES_USER=arch-stats
POSTGRES_DB=arch-stats
POSTGRES_PORT=5432
POSTGRES_SOCKET_DIR=/var/run/postgresql
ARCH_STATS_DEV_MODE=False
ARCH_STATS_GOOGLE_OAUTH_CLIENT_ID=${ARCH_STATS_GOOGLE_OAUTH_CLIENT_ID}
ARCH_STATS_JWT_SECRET=${ARCH_STATS_JWT_SECRET}
EOF

        # Upload everything
        scp -o BatchMode=yes \
            "${SCRIPT_DIR}/remote_installer.bash" \
            "${SCRIPT_DIR}/install_app.bash" \
            "${SCRIPT_DIR}/../OS/cloudflared/cloudflared.service" \
            "${SCRIPT_DIR}/../OS/cloudflared/cloudflared-update.service" \
            "${SCRIPT_DIR}/../OS/cloudflared/cloudflared-update.timer" \
            "${temp_dir}/cloudflared_config.yaml" \
            "${temp_dir}/arch-stats.service" \
            "${temp_dir}/.env" \
            "${cred_file}" \
            "${host}":/tmp/

        rm -rf "${temp_dir}"

        ssh -t "${host}" "sudo chmod +x /tmp/remote_installer.bash && sudo GITHUB_TOKEN='${GITHUB_TOKEN}' /tmp/remote_installer.bash"

    elif [[ "${action}" == "update" ]]; then
        echo "Starting update on '${host}'..."
        scp -o BatchMode=yes "${SCRIPT_DIR}/install_app.bash" "${host}":/tmp/
        ssh -t "${host}" "
            sudo systemctl stop arch-stats.service && \
            sudo -u arch-stats GITHUB_TOKEN='${GITHUB_TOKEN}' bash /tmp/install_app.bash /opt/arch-stats && \
            sudo systemctl start arch-stats.service
        "
    fi

    exit 0
}

main "$@"
```

**Step 2: Commit**
`git add scripts/deploy.bash`

---

### Task 7: Linting and Validation

**Files:**
- Test: `scripts/deploy.bash`
- Test: `scripts/remote_installer.bash`
- Test: `scripts/remote_uninstaller.bash`
- Test: `scripts/install_app.bash`

**Step 1: Run linting**
Execute bash linting checks:
`./scripts/linting.bash --scripts`
Expected: Zero ShellCheck/shfmt violations.
