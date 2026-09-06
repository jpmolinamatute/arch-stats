# Deployment Scripts & Systemd Service Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Update deployment scripts (`install_app.bash`, `remote_installer.bash`, `deploy.bash`) and systemd unit configuration for the single-binary Go deployment model, eliminate legacy Python/uv/uvicorn references, verify end-to-end execution against the Raspberry Pi emulator, and mark Task 034 as done.

**Architecture:** Transition from Python virtualenv + tarball extraction + manual SQL migrations to a single statically compiled Go binary (`arch-stats`) downloaded from GitHub Releases, verified via SHA-256 checksum, placed at `/opt/arch-stats/arch-stats`, managed via systemd unit with `/opt/arch-stats/.env`, running migrations via `/opt/arch-stats/arch-stats migrate`.

**Tech Stack:** Bash, systemd, OpenSSH, PostgreSQL 15/17, Docker, Docker Compose, Go 1.24.

**Spec:** [docs/go_refactor/tasks/034-deployment_scripts_and_systemd.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/034-deployment_scripts_and_systemd.md)

## Global Constraints

- Target branch: `refactor/034-deployment-scripts-and-systemd`
- Binary destination: `/opt/arch-stats/arch-stats`
- Checksum verification: SHA-256 against release asset `arch-stats.sha256`
- Systemd service unit: `ExecStart=/opt/arch-stats/arch-stats`, `EnvironmentFile=/opt/arch-stats/.env`
- Zero Python/venv/uv references in `scripts/install_app.bash` and `scripts/deploy.bash`
- All bash scripts must pass `shellcheck` and `shfmt`
- End-to-end verification must pass in `arch-stats-emulator` Docker container on port 2222 with HTTP 200 response on `http://localhost:8001/api/v0/health`
- Live tracking checklist in [docs/plans/task.md](file:///home/juanpa/Projects/arch-stats/docs/plans/task.md) and task file checklist in [docs/go_refactor/tasks/034-deployment_scripts_and_systemd.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/034-deployment_scripts_and_systemd.md) must be marked completed at the end of implementation

---

### Task 1: Switch Git Branch & Delete Legacy Script

**Files:**
- Delete: `scripts/start_uvicorn.bash`

**Interfaces:**
- Consumes: Existing git repository on branch `main`
- Produces: Working branch `refactor/034-deployment-scripts-and-systemd` without `scripts/start_uvicorn.bash`

- [ ] **Step 1: Check out feature branch**

```bash
git checkout -b refactor/034-deployment-scripts-and-systemd
```

- [ ] **Step 2: Delete `scripts/start_uvicorn.bash`**

```bash
git rm scripts/start_uvicorn.bash
```

- [ ] **Step 3: Verify script removal and status**

Run: `git status`
Expected: `deleted: scripts/start_uvicorn.bash`

- [ ] **Step 4: Commit**

```bash
git commit -m "chore: remove legacy scripts/start_uvicorn.bash"
```

---

### Task 2: Update Systemd Service Template

**Files:**
- Modify: `scripts/templates/arch-stats.service.j2`

**Interfaces:**
- Consumes: Application name `{{ app_name }}`
- Produces: Systemd unit file template pointing directly to single binary at `/opt/{{ app_name }}/{{ app_name }}`

- [ ] **Step 1: Update `scripts/templates/arch-stats.service.j2`**

Replace `scripts/templates/arch-stats.service.j2` with:

```ini
[Unit]
Description=Arch Stats Server
After=network.target postgresql.service
Requires=postgresql.service

[Service]
Type=simple
User={{ app_name }}
Group={{ app_name }}
ExecStart=/opt/{{ app_name }}/{{ app_name }}
EnvironmentFile=/opt/{{ app_name }}/.env
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

- [ ] **Step 2: Verify no Python/uvicorn references in the template**

Run:
```bash
grep -rn "uvicorn\|backend-old\|\.venv" scripts/templates/arch-stats.service.j2 || true
```
Expected: No output.

- [ ] **Step 3: Commit**

```bash
git add scripts/templates/arch-stats.service.j2
git commit -m "chore: update systemd service template for single Go binary"
```

---

### Task 3: Refactor `scripts/install_app.bash`

**Files:**
- Modify: `scripts/install_app.bash`

**Interfaces:**
- Consumes: `GITHUB_TOKEN` environment variable, optional target user `$1` (defaults to `arch-stats`)
- Produces: `/opt/arch-stats/arch-stats` executable binary, runs migrations, manages service lifecycle

- [ ] **Step 1: Update `scripts/install_app.bash`**

Rewrite `scripts/install_app.bash` to implement the simplified Go single binary installation lifecycle:
1. Validate running as root (`$EUID -eq 0`)
2. Validate `GITHUB_TOKEN` presence
3. Resolve latest release metadata from `https://api.github.com/repos/jpmolinamatute/arch-stats/releases/latest`
4. Locate asset `arch-stats` and checksum `arch-stats.sha256`
5. Download binary and checksum to secure temp directory
6. Verify SHA-256 hash using `sha256sum`
7. Stop `arch-stats.service` if running (`systemctl stop arch-stats.service || true`)
8. Place binary at `/opt/arch-stats/arch-stats`, set permissions to `755` and ownership `arch-stats:arch-stats`
9. Run migrations as the `arch-stats` user via `(cd /opt/arch-stats && runuser -u arch-stats -- /opt/arch-stats/arch-stats migrate)`
10. If systemd service exists (`/etc/systemd/system/arch-stats.service`), restart it (`systemctl restart arch-stats.service`)

Implementation:

```bash
#!/usr/bin/env bash

set -Eeuo pipefail

: "${GITHUB_TOKEN:?Environment variable GITHUB_TOKEN is not set}"

export PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

log_info() { echo "INFO: $*"; }
log_error() { echo "ERROR: $*" >&2; }

# shellcheck disable=SC2329
cleanup_tmp_workspace() {
    local tmp_dir="${1}"
    if [[ -d "${tmp_dir}" ]]; then
        rm -rf "${tmp_dir}"
        log_info "Removed temp workspace: ${tmp_dir}"
    fi
}

gh_download() {
    local url="${1}"
    local out="${2}"
    local api_call="${3:-false}"
    local curl_ec
    local curl_opts=(
        -fsSL
        --max-time 120
        --connect-timeout 15
        --retry 3
        --retry-delay 2
        --retry-all-errors
        --user-agent "arch-stats-installer"
        -H "Authorization: Bearer ${GITHUB_TOKEN}"
        --output "${out}"
    )
    if [[ "${api_call}" == "true" ]]; then
        curl_opts+=(
            -H "Accept: application/vnd.github+json"
            -H "X-GitHub-Api-Version: 2022-11-28"
        )
    fi
    if ! curl "${curl_opts[@]}" "${url}"; then
        curl_ec=$?
        log_error "Download failed (curl exit=${curl_ec}) url=${url}"
        exit 15
    fi
}

get_release_metadata() {
    local base_url="${1}"
    local release_json_file="${2}"
    local api_url="${base_url}/releases/latest"

    log_info "Resolving latest release metadata from GitHub API"
    gh_download "${api_url}" "${release_json_file}" true
}

json_get_asset_url() {
    local asset_name="${1}"
    local release_json_file="${2}"
    local url
    url="$(jq -r --arg name "${asset_name}" '.assets[] | select(.name==$name) | .browser_download_url' "${release_json_file}")"
    if [[ -z "${url}" || "${url}" == "null" ]]; then
        log_error "Could not find asset '${asset_name}' in latest release."
        exit 1
    fi
    echo "${url}"
}

get_expected_sha256() {
    local checksum_url="${1}"
    local checksum_file="${2}"
    local sha

    gh_download "${checksum_url}" "${checksum_file}" false
    sha="$(grep -Eoi '^[0-9a-f]{64}' "${checksum_file}" | head -n 1 || true)"
    if [[ -z "${sha}" ]]; then
        log_error "Checksum file did not contain a valid 64-hex SHA-256 hash."
        exit 1
    fi
    echo "${sha}"
}

verify_sha256() {
    local file="${1}"
    local expected="${2}"
    local actual

    actual="$(sha256sum "${file}" | awk '{print $1}')"
    if [[ "${actual}" != "${expected}" ]]; then
        log_error "SHA-256 mismatch. Expected='${expected}' Actual='${actual}'"
        exit 1
    fi
    log_info "SHA-256 OK: ${actual}"
}

assert_postgres_socket() {
    local socket_path="/var/run/postgresql/.s.PGSQL.5432"
    if [[ ! -d /var/run/postgresql/ || ! -S "${socket_path}" ]]; then
        log_error "PostgreSQL socket not found at ${socket_path}. Is PostgreSQL running?"
        exit 10
    fi
    log_info "Detected PostgreSQL socket: ${socket_path}"
}

main() {
    local app_user="${1:-arch-stats}"
    local app_name="arch-stats"
    local repo="jpmolinamatute/arch-stats"
    local base_url="https://api.github.com/repos/${repo}"
    local install_dir="/opt/${app_user}"
    local target_bin="${install_dir}/${app_name}"

    if [[ ${EUID} -ne 0 ]]; then
        log_error "This script must be run as root."
        exit 1
    fi

    local tmp_dir
    tmp_dir="$(mktemp -d -t "arch-stats-installer.XXXXXX")"
    trap 'cleanup_tmp_workspace "${tmp_dir}"' EXIT

    assert_postgres_socket

    local release_json="${tmp_dir}/release.json"
    get_release_metadata "${base_url}" "${release_json}"

    local bin_url checksum_url expected_sha
    bin_url="$(json_get_asset_url "${app_name}" "${release_json}")"
    checksum_url="$(json_get_asset_url "${app_name}.sha256" "${release_json}")"

    local dl_bin="${tmp_dir}/${app_name}"
    local dl_checksum="${tmp_dir}/${app_name}.sha256"

    log_info "Downloading ${app_name} binary..."
    gh_download "${bin_url}" "${dl_bin}" false

    log_info "Downloading ${app_name}.sha256 checksum..."
    expected_sha="$(get_expected_sha256 "${checksum_url}" "${dl_checksum}")"

    log_info "Verifying SHA-256 checksum..."
    verify_sha256 "${dl_bin}" "${expected_sha}"

    if systemctl is-active --quiet "${app_user}.service" 2>/dev/null; then
        log_info "Stopping active ${app_user}.service before updating binary..."
        systemctl stop "${app_user}.service" || true
    fi

    mkdir -p "${install_dir}"
    log_info "Installing binary to ${target_bin}..."
    install -m 755 -o "${app_user}" -g "${app_user}" "${dl_bin}" "${target_bin}"

    log_info "Running database migrations as user ${app_user}..."
    if ! runuser -u "${app_user}" -- env -C "${install_dir}" "${target_bin}" migrate; then
        log_error "Database migrations failed."
        exit 13
    fi
    log_info "Database migrations completed successfully."

    if [[ -f "/etc/systemd/system/${app_user}.service" ]]; then
        log_info "Starting ${app_user}.service..."
        systemctl restart "${app_user}.service"
    fi

    log_info "Installation completed successfully."
    exit 0
}

main "$@"
```

- [ ] **Step 2: Check formatting and syntax**

Run:
```bash
shellcheck --shell=bash -x scripts/install_app.bash
shfmt --language-dialect bash -i 4 -w scripts/install_app.bash
```
Expected: Zero lint or syntax warnings.

- [ ] **Step 3: Commit**

```bash
git add scripts/install_app.bash
git commit -m "refactor(scripts): simplify install_app.bash for Go single binary release"
```

---

### Task 4: Refactor `scripts/remote_installer.bash`

**Files:**
- Modify: `scripts/remote_installer.bash`

**Interfaces:**
- Consumes: Target user `$1`, staged deployment files in `$ROOT_DIR`
- Produces: System user, `.env` file with complete Go configuration, PostgreSQL setup, cloudflared setup, systemd registration, and clean binary deployment

- [ ] **Step 1: Update `scripts/remote_installer.bash`**

Update `generate_env_file` to write all configuration keys required by the Go backend in production mode (`ARCH_STATS_DEV_MODE="false"`, socket directory, pool sizes, query timeouts, websocket channel, JWT secret, session configs, etc.).
Update `setup_cloudflared` to safely enable and optionally start without failing if disconnected.
Update `install_app_as_user` to run `install_app.bash` as root with `GITHUB_TOKEN` passed through.

Implementation details:
```bash
#!/usr/bin/env bash

set -Eeuo pipefail

ROOT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd -P)"

log_info() { echo "INFO: $*"; }
log_error() { echo "ERROR: $*" >&2; }

# shellcheck disable=SC2329
cleanup() {
    log_info "Cleaning up temporary installation files..."
    if [[ -n "${ROOT_DIR}" && "${ROOT_DIR}" != "/tmp" && "${ROOT_DIR}" != "/" ]]; then
        rm -rf "${ROOT_DIR}"
    else
        log_error "Safety check failed: ROOT_DIR is '${ROOT_DIR}', skipping deletion."
    fi
}

create_app_user() {
    local app_user="${1}"
    if ! id -u "${app_user}" >/dev/null 2>&1; then
        log_info "Creating system user '${app_user}'..."
        useradd -r -m -d "/opt/${app_user}" -s "/usr/sbin/nologin" "${app_user}"
    fi
}

generate_env_file() {
    local app_user="${1}"
    local postgres_password="${2}"
    local user_dir
    local jwt_secret
    user_dir="$(getent passwd "${app_user}" | cut -d: -f6)"

    jwt_secret="$(openssl rand -hex 32)"

    cat <<EOF >"${user_dir}/.env"
POSTGRES_USER="${app_user}"
POSTGRES_PASSWORD="${postgres_password}"
POSTGRES_DB="${app_user}"
POSTGRES_HOST=""
POSTGRES_PORT="5432"
POSTGRES_SOCKET_DIR="/var/run/postgresql"
POSTGRES_POOL_MIN_SIZE="1"
POSTGRES_POOL_MAX_SIZE="10"
POSTGRES_MAX_QUERIES="50000"
POSTGRES_MAX_INACTIVE_CONNECTION_LIFETIME="300.0"
POSTGRES_COMMAND_TIMEOUT="15.0"
POSTGRES_STATEMENT_CACHE_SIZE="200"

ARCH_STATS_SERVER_PORT="8001"
ARCH_STATS_DEV_MODE="false"
ARCH_STATS_WS_CHANNEL="archy"
APPLY_DB_MIGRATIONS_ON_START="true"

SESSION_TTL_HOURS="24"
SESSION_TOKEN_BYTES="32"

ARCH_STATS_GOOGLE_OAUTH_CLIENT_ID="${ARCH_STATS_GOOGLE_OAUTH_CLIENT_ID:-}"
VITE_GOOGLE_CLIENT_ID="${ARCH_STATS_GOOGLE_OAUTH_CLIENT_ID:-}"
ARCH_STATS_JWT_SECRET="${jwt_secret}"
ARCH_STATS_JWT_ALGORITHM="HS256"
ARCH_STATS_JWT_TTL_MINUTES="60"
EOF

    chown "${app_user}:${app_user}" "${user_dir}/.env"
    chmod 600 "${user_dir}/.env"
}

install_os_packages() {
    log_info "Configuring Cloudflare repository..."
    local codename="bookworm"
    if [[ -f /etc/os-release ]]; then
        codename="$(. /etc/os-release && echo "${VERSION_CODENAME:-bookworm}")"
    fi
    curl -fsSL https://pkg.cloudflare.com/cloudflare-main.gpg | tee /usr/share/keyrings/cloudflare-main.gpg >/dev/null
    echo "deb [signed-by=/usr/share/keyrings/cloudflare-main.gpg] https://pkg.cloudflare.com/cloudflared ${codename} main" | tee /etc/apt/sources.list.d/cloudflared.list >/dev/null

    log_info "Installing OS packages (cloudflared, postgresql, postgresql-contrib, openssl, jq, curl)..."
    apt-get update -y
    apt-get install -y \
        cloudflared \
        postgresql \
        postgresql-contrib \
        openssl \
        jq \
        curl
}

setup_postgres() {
    local app_user="${1}"
    local postgres_password="${2}"
    local pg_path="/etc/postgresql/15/main"
    if [[ ! -d "${pg_path}" ]]; then
        # Find active postgresql conf directory if version differs
        pg_path="$(find /etc/postgresql -mindepth 2 -maxdepth 2 -type d | head -n 1)"
    fi
    log_info "Setting up PostgreSQL user and database in ${pg_path}..."

    log_info "Stopping postgresql service if running to apply custom configurations..."
    systemctl stop postgresql || true

    log_info "Applying custom PostgreSQL configurations..."
    mv "${ROOT_DIR}/postgresql.conf" "${ROOT_DIR}/secondary.conf" "${ROOT_DIR}/pg_hba.conf" "${pg_path}/"
    chown postgres:postgres "${pg_path}/postgresql.conf" "${pg_path}/secondary.conf" "${pg_path}/pg_hba.conf"
    chmod 644 "${pg_path}/postgresql.conf" "${pg_path}/secondary.conf" "${pg_path}/pg_hba.conf"

    systemctl enable --now postgresql

    (
        cd ~postgres
        if ! sudo -u postgres psql -t -c '\du' | cut -d \| -f 1 | grep -qw "${app_user}"; then
            log_info "Creating PostgreSQL user '${app_user}'..."
            sudo -u postgres psql -c "CREATE USER \"${app_user}\" WITH PASSWORD '${postgres_password}';"
        fi

        if ! sudo -u postgres psql -lqt | cut -d \| -f 1 | grep -qw "${app_user}"; then
            log_info "Creating PostgreSQL database '${app_user}'..."
            sudo -u postgres createdb -O "${app_user}" "${app_user}"
        fi
    )
    log_info "PostgreSQL setup complete."
}

setup_cloudflared() {
    local app_user="${1}"
    local src_cred_file
    local dest_cred_file
    local src_cert_file="${ROOT_DIR}/cert.pem"
    local dest_cert_file
    local user_dir
    local cf_dir
    log_info "Setting up cloudflared..."
    src_cred_file="$(find "${ROOT_DIR}" -maxdepth 1 -name "*.json" | head -n 1)"
    if [[ ! -f "${src_cred_file}" || ! -f "${src_cert_file}" ]]; then
        log_error "Cloudflared credentials file or cert file not found in ${ROOT_DIR}. Aborting."
        exit 23
    fi

    user_dir="$(getent passwd "${app_user}" | cut -d: -f6)"
    cf_dir="${user_dir}/.cloudflared"
    dest_cred_file="${cf_dir}/$(basename "${src_cred_file}")"
    dest_cert_file="${cf_dir}/cert.pem"
    mkdir -p "${cf_dir}"
    mv "${src_cred_file}" "${dest_cred_file}"
    mv "${src_cert_file}" "${dest_cert_file}"
    chmod 700 "${cf_dir}"
    chmod 600 "${dest_cred_file}" "${dest_cert_file}"
    chown -R "${app_user}:${app_user}" "${cf_dir}"

    mkdir -p "/etc/cloudflared"
    mv "${ROOT_DIR}/cloudflared_config.yaml" /etc/cloudflared/cloudflared_config.yaml
    mv "${ROOT_DIR}/cloudflared.service" /etc/systemd/system/
    chmod 644 /etc/systemd/system/cloudflared.service /etc/cloudflared/cloudflared_config.yaml

    systemctl daemon-reload
    systemctl enable cloudflared.service
    systemctl start cloudflared.service || true
    log_info "Cloudflared setup complete."
}

register_app_service() {
    local app_user="${1}"
    mv "${ROOT_DIR}/${app_user}.service" "/etc/systemd/system/"
    chmod 644 "/etc/systemd/system/${app_user}.service"
    systemctl daemon-reload
    systemctl enable "${app_user}.service"
}

run_install_app() {
    local app_user="${1}"
    local env_file="${ROOT_DIR}/env"
    if [[ -f "${env_file}" ]]; then
        # shellcheck disable=SC1090
        source "${env_file}"
    else
        log_error "Environment file not found. Aborting."
        exit 15
    fi

    log_info "Executing application installer for ${app_user}..."
    chmod 755 "${ROOT_DIR}/install_app.bash"
    GITHUB_TOKEN="${GITHUB_TOKEN}" "${ROOT_DIR}/install_app.bash" "${app_user}"
}

main() {
    local app_user="${1}"
    local postgres_password

    trap cleanup EXIT
    if [[ $EUID -ne 0 ]]; then
        log_error "Please run as root."
        exit 1
    fi

    install_os_packages
    postgres_password="$(openssl rand -hex 16)"
    create_app_user "${app_user}"
    generate_env_file "${app_user}" "${postgres_password}"
    setup_postgres "${app_user}" "${postgres_password}"
    setup_cloudflared "${app_user}"
    register_app_service "${app_user}"
    run_install_app "${app_user}"

    log_info "Checking ${app_user} service status..."
    if ! systemctl is-active --quiet "${app_user}.service"; then
        log_error "Service ${app_user}.service failed to start."
        journalctl -u "${app_user}.service" -n 50 --no-pager || true
        exit 22
    fi

    log_info "Remote installation completed successfully."
    exit 0
}

main "$@"
```

- [ ] **Step 2: Run shellcheck and shfmt**

Run:
```bash
shellcheck --shell=bash -x scripts/remote_installer.bash
shfmt --language-dialect bash -i 4 -w scripts/remote_installer.bash
```
Expected: Clean output.

- [ ] **Step 3: Commit**

```bash
git add scripts/remote_installer.bash
git commit -m "refactor(scripts): update remote_installer.bash for Go binary deployment"
```

---

### Task 5: Refactor `scripts/deploy.bash`

**Files:**
- Modify: `scripts/deploy.bash`

**Interfaces:**
- Consumes: Target SSH host and optional flags (`-p <port>`, `-i <identity_file>`), local configuration from `.env`
- Produces: Renders service templates via native bash/sed (removing `uv`/Python requirements), uploads assets, triggers `remote_installer.bash` or `install_app.bash`

- [ ] **Step 1: Update `scripts/deploy.bash`**

Update `check_remote_action`:
Check for `/opt/${APP_NAME}/${APP_NAME}` instead of `uv` and `/opt/${APP_NAME}/backend`.

Add native bash template rendering `render_templates` function replacing `transform_templates.py`:
- `arch-stats.service.j2` -> `arch-stats.service`
- `cloudflared_config.yaml.j2` -> `cloudflared_config.yaml`
- `pg_hba.conf.j2` -> `pg_hba.conf`

Support SSH port and key flags (`-p`, `-i`, or `host:port`) for seamless testing with the emulator.

Update `update()` to invoke `install_app.bash` cleanly with root permissions.

Verify no Python, venv, or uv references remain in `scripts/deploy.bash`.

Implementation:
```bash
#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")"
UNINSTALLER_SCRIPT="remote_uninstaller.bash"
APP_NAME="arch-stats"
REMOTE_SECURE_DIR="/tmp/deploy_assets"
# shellcheck source=./lib/logging
. "${SCRIPT_DIR}/lib/logging"

SSH_OPTS=(-o BatchMode=yes -o StrictHostKeyChecking=no)
SCP_OPTS=(-o BatchMode=yes -o StrictHostKeyChecking=no)

usage() {
    cat <<EOF
Usage: $(basename "${0}") [-p <port>] [-i <identity_file>] <remote-host> [<action>]
Description: Deploy or uninstall ${APP_NAME} remotely.

Options:
  -p <port>           SSH/SCP port
  -i <identity_file>  SSH/SCP private key file

Arguments:
  <remote-host>       SSH host target (required)
  [<action>]          Option: "install" or "uninstall". Defaults to auto-resolving install/update.
EOF
}

render_templates() {
    local temp_dir="${1}"
    local app_name="${2}"
    local app_user_home_dir="/opt/${app_name}"
    local server_port="${ARCH_STATS_SERVER_PORT:-8001}"
    local tunnel_id="${CLOUDFLARED_TUNNEL_ID}"

    log_info "Rendering configuration templates..."

    sed \
        -e "s|{{ app_name }}|${app_name}|g" \
        "${SCRIPT_DIR}/templates/arch-stats.service.j2" >"${temp_dir}/${app_name}.service"

    sed \
        -e "s|{{ cloudflared_tunnel_id }}|${tunnel_id}|g" \
        -e "s|{{ app_user_home_dir }}|${app_user_home_dir}|g" \
        -e "s|{{ app_name }}|${app_name}|g" \
        -e "s|{{ prod_uvicorn_port }}|${server_port}|g" \
        "${SCRIPT_DIR}/templates/cloudflared_config.yaml.j2" >"${temp_dir}/cloudflared_config.yaml"

    sed \
        -e "s|{{ app_name }}|${app_name}|g" \
        "${SCRIPT_DIR}/templates/pg_hba.conf.j2" >"${temp_dir}/pg_hba.conf"
}

check_remote_action() {
    local host="${1}"
    if ssh -q "${SSH_OPTS[@]}" -o ConnectTimeout=5 "${host}" "
      id -u ${APP_NAME} >/dev/null 2>&1 && \
      command -v cloudflared >/dev/null 2>&1 && \
      systemctl list-units --type=service | grep -q 'postgresql' && \
      [ -f '/opt/${APP_NAME}/${APP_NAME}' ]
    "; then
        echo "update"
    else
        echo "install"
    fi
}

uninstall() {
    local host="${1}"
    log_info "Starting remote uninstallation on '${host}'..."
    log_info "Uploading uninstaller script to '${host}:/tmp'"
    scp "${SCP_OPTS[@]}" "${SCRIPT_DIR}/${UNINSTALLER_SCRIPT}" "${host}":/tmp/
    ssh -t "${SSH_OPTS[@]}" "${host}" "sudo bash /tmp/${UNINSTALLER_SCRIPT}"
}

upload_install_assets() {
    local host="${1}"
    local temp_dir="${2}"
    local cred_file="${3}"
    local cert_file="${4}"
    local env_temp_file="${temp_dir}/env"

    log_info "Uploading installation assets to '${host}:${REMOTE_SECURE_DIR}'..."
    scp "${SCP_OPTS[@]}" \
        "${SCRIPT_DIR}/remote_installer.bash" \
        "${SCRIPT_DIR}/install_app.bash" \
        "${SCRIPT_DIR}/cloudflared/cloudflared.service" \
        "${SCRIPT_DIR}/pg_conf/postgresql.conf" \
        "${SCRIPT_DIR}/pg_conf/secondary.conf" \
        "${temp_dir}/pg_hba.conf" \
        "${temp_dir}/cloudflared_config.yaml" \
        "${temp_dir}/${APP_NAME}.service" \
        "${cred_file}" \
        "${cert_file}" \
        "${env_temp_file}" \
        "${host}:${REMOTE_SECURE_DIR}/"
}

upload_update_assets() {
    local host="${1}"
    local temp_dir="${2}"
    local env_temp_file="${temp_dir}/env"

    log_info "Uploading update assets to '${host}:${REMOTE_SECURE_DIR}'..."
    scp "${SCP_OPTS[@]}" \
        "${SCRIPT_DIR}/install_app.bash" \
        "${env_temp_file}" \
        "${host}:${REMOTE_SECURE_DIR}/"
}

install() {
    local host="${1}"
    local temp_dir="${2}"
    local cred_file
    local cert_file="${HOME}/.cloudflared/cert.pem"
    log_info "Starting clean installation on '${host}'..."

    : "${ARCH_STATS_GOOGLE_OAUTH_CLIENT_ID:?Google OAuth Client ID is required for clean install}"
    : "${ARCH_STATS_JWT_SECRET:?JWT Secret is required for clean install}"
    : "${CLOUDFLARED_TUNNEL_ID:?Cloudflared Tunnel ID is required for clean install}"

    cred_file="${HOME}/.cloudflared/${CLOUDFLARED_TUNNEL_ID}.json"
    if [[ ! -f "${cred_file}" ]]; then
        log_error "Local Cloudflared credential file not found at: ${cred_file}"
        exit 1
    fi

    render_templates "${temp_dir}" "${APP_NAME}"

    upload_install_assets "${host}" "${temp_dir}" "${cred_file}" "${cert_file}"
    ssh -t "${SSH_OPTS[@]}" "${host}" "sudo bash -c '${REMOTE_SECURE_DIR}/remote_installer.bash \"${APP_NAME}\"'"
}

update() {
    local host="${1}"
    local temp_dir="${2}"

    log_info "Starting update on '${host}'..."
    upload_update_assets "${host}" "${temp_dir}"
    ssh -t "${SSH_OPTS[@]}" "${host}" "
        sudo bash -c '
            source ${REMOTE_SECURE_DIR}/env
            bash ${REMOTE_SECURE_DIR}/install_app.bash ${APP_NAME}
        '
        ec=\$?
        sudo rm -rf ${REMOTE_SECURE_DIR}
        exit \$ec
    "
}

check_connection() {
    local host="${1}"
    log_info "Checking connection to '${host}'..."
    if ! ssh "${SSH_OPTS[@]}" -o ConnectTimeout=5 "${host}" exit 2>/dev/null; then
        log_error "ERROR: Cannot connect to remote host: ${host}"
        exit 1
    fi
}

main() {
    local port="" identity="" host="" action=""

    while [[ $# -gt 0 ]]; do
        case "$1" in
        -p)
            port="$2"
            shift 2
            ;;
        -i)
            identity="$2"
            shift 2
            ;;
        -h | --help)
            usage
            exit 0
            ;;
        *)
            if [[ -z "${host}" ]]; then
                host="$1"
            elif [[ -z "${action}" ]]; then
                action="$1"
            else
                log_error "Unexpected argument: $1"
                usage
                exit 1
            fi
            shift
            ;;
        esac
    done

    if [[ -z "${host}" ]]; then
        usage
        exit 1
    fi

    if [[ "${host}" =~ ^(.*):([0-9]+)$ ]]; then
        host="${BASH_REMATCH[1]}"
        port="${BASH_REMATCH[2]}"
    fi

    if [[ -n "${port}" ]]; then
        SSH_OPTS+=(-p "${port}")
        SCP_OPTS+=(-P "${port}")
    fi

    if [[ -n "${identity}" ]]; then
        SSH_OPTS+=(-i "${identity}")
        SCP_OPTS+=(-i "${identity}")
    fi

    if [[ -f "${SCRIPT_DIR}/../.env" ]]; then
        # shellcheck source=/dev/null
        source "${SCRIPT_DIR}/../.env"
    fi

    check_connection "${host}"

    if [[ -z "${action}" || "${action}" == "install" ]]; then
        action=$(check_remote_action "${host}")
        log_info "Resolved deployment action: ${action}"
        local temp_dir env_temp_file
        temp_dir="$(mktemp -d)"
        env_temp_file="${temp_dir}/env"
        : "${GITHUB_TOKEN:?Environment variable GITHUB_TOKEN must be set}"
        echo "export GITHUB_TOKEN='${GITHUB_TOKEN}'" >"${env_temp_file}"
        echo "export ARCH_STATS_GOOGLE_OAUTH_CLIENT_ID='${ARCH_STATS_GOOGLE_OAUTH_CLIENT_ID:-}'" >>"${env_temp_file}"
        chmod 400 "${env_temp_file}"
        log_info "Creating secure remote temporary directory..."
        ssh "${SSH_OPTS[@]}" "${host}" "mkdir -p -m 700 ${REMOTE_SECURE_DIR}"
        if [[ "${action}" == "install" ]]; then
            install "${host}" "${temp_dir}"
        elif [[ "${action}" == "update" ]]; then
            update "${host}" "${temp_dir}"
        fi
        rm -rf "${temp_dir}"
    elif [[ "${action}" == "uninstall" ]]; then
        uninstall "${host}"
    else
        log_error "ERROR: Invalid action: ${action}"
        usage
        exit 1
    fi
    log_info "Deployment complete."
    exit 0
}

main "$@"
```

- [ ] **Step 2: Run shellcheck and shfmt on `scripts/deploy.bash`**

Run:
```bash
shellcheck --shell=bash -x scripts/deploy.bash
shfmt --language-dialect bash -i 4 -w scripts/deploy.bash
```
Expected: Zero lint and formatting errors.

- [ ] **Step 3: Commit**

```bash
git add scripts/deploy.bash
git commit -m "refactor(scripts): update deploy.bash for single binary Go deployment model"
```

---

### Task 6: Script Linting and Python Purge Verification

**Files:**
- Test: `scripts/install_app.bash`, `scripts/remote_installer.bash`, `scripts/deploy.bash`

**Interfaces:**
- Consumes: All modified scripts
- Produces: Clean linting check passing CI standards

- [ ] **Step 1: Run shellcheck across all deployment scripts**

Run:
```bash
shellcheck scripts/install_app.bash scripts/remote_installer.bash scripts/deploy.bash
```
Expected: Exit code 0, no output.

- [ ] **Step 2: Verify zero Python/venv/uv references remain in deployment scripts**

Run:
```bash
grep -rn "venv\|uvicorn\|uv sync\|pip\|python" scripts/install_app.bash scripts/deploy.bash || true
```
Expected: No matches (zero exit output).

- [ ] **Step 3: Run repository bash check**

Run:
```bash
./scripts/linting.bash --scripts
```
Expected:
```
INFO: Running bash linter
INFO: Running bash formatter
```
with exit status 0.

---

### Task 7: End-to-End Raspberry Pi Emulator Deployment Verification

**Files:**
- Test: Target `root@localhost:2222` inside emulator container

**Interfaces:**
- Consumes: `docker/docker-compose.yaml` (profile `emulator`), `docker/ssh/arch_stats_dev` key
- Produces: Verified active systemd service, verified database migrations, HTTP 200 health check response

- [ ] **Step 1: Set secure permissions on development SSH private key**

Run:
```bash
chmod 600 docker/ssh/arch_stats_dev
```

- [ ] **Step 2: Start the emulator container**

Run:
```bash
docker compose -f docker/docker-compose.yaml --profile emulator up -d --build
```
Expected:
`Container arch-stats-emulator-1 Started`

- [ ] **Step 3: Verify SSH connectivity to the emulator**

Run:
```bash
ssh -i docker/ssh/arch_stats_dev -p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null root@localhost "systemctl is-system-running --wait || true"
```
Expected:
Output contains `running` or `degraded` (with systemd fully booted).

- [ ] **Step 4: Execute deployment against the emulator**

Run:
```bash
./scripts/deploy.bash -p 2222 -i docker/ssh/arch_stats_dev root@localhost install
```
Expected:
Deployment executes to completion without errors and logs `Deployment complete.`

- [ ] **Step 5: Verify systemd service status in emulator**

Run:
```bash
ssh -i docker/ssh/arch_stats_dev -p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null root@localhost "systemctl status arch-stats.service"
```
Expected:
`Active: active (running)`

- [ ] **Step 6: Verify HTTP API health endpoint and migration schema version**

Run:
```bash
ssh -i docker/ssh/arch_stats_dev -p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null root@localhost "curl -s http://localhost:8001/api/v0/health"
```
Expected:
JSON response containing `"status":"ok"` and `"schema_version":...` with HTTP 200.

- [ ] **Step 7: Tear down emulator container**

Run:
```bash
docker compose -f docker/docker-compose.yaml --profile emulator down
```
Expected:
Container and network removed cleanly.

---

### Task 8: Mark Tasks as Done and Finalize

**Files:**
- Modify: `docs/go_refactor/tasks/034-deployment_scripts_and_systemd.md`
- Modify: `docs/plans/task.md`

**Interfaces:**
- Consumes: Verified implementation and test evidence
- Produces: Checked off acceptance criteria and tasks, updated live tracker

- [ ] **Step 1: Update `docs/go_refactor/tasks/034-deployment_scripts_and_systemd.md`**

Mark all Acceptance Criteria and Steps checkboxes as completed (`- [x]`):
- Acceptance Criteria:
  - `- [x] scripts/install_app.bash is simplified...`
  - `- [x] scripts/remote_installer.bash is updated...`
  - `- [x] scripts/deploy.bash is updated...`
  - `- [x] Systemd service file is updated...`
  - `- [x] scripts/start_uvicorn.bash is removed...`
  - `- [x] All scripts pass shellcheck...`
  - `- [x] The deployment script fully runs and passes end-to-end testing in the local emulator...`
- Steps:
  - `- [x] Step 1: Update install_app.bash`
  - `- [x] Step 2: Update remote_installer.bash`
  - `- [x] Step 3: Update deploy.bash`
  - `- [x] Step 4: Delete scripts/start_uvicorn.bash`
  - `- [x] Step 5: Update systemd service template`
  - `- [x] Step 6: Run shellcheck`
  - `- [x] Step 7: Verify no Python/venv references remain`
  - `- [x] Step 8: Test deployment against Raspberry Pi Docker emulator`
  - `- [x] Step 9: Commit`

- [ ] **Step 2: Update `docs/plans/task.md`**

Update `docs/plans/task.md` to record Task 034 as completed:

```markdown
| Task | Status | Description |
| --- | --- | --- |
| Task 1: Git Branch & Legacy Script Removal | DONE | Switch to `refactor/034-deployment-scripts-and-systemd` and remove `scripts/start_uvicorn.bash` |
| Task 2: Systemd Service Template Update | DONE | Update `scripts/templates/arch-stats.service.j2` for single Go binary |
| Task 3: Simplify `install_app.bash` | DONE | Download Go binary and checksum from GitHub Release, verify sha256, place binary, run migrations |
| Task 4: Update `remote_installer.bash` | DONE | Complete `.env` variables, service registration, root execution flow |
| Task 5: Update `deploy.bash` | DONE | Native template rendering, remove Python/uv/uvicorn references, SSH port/identity flags |
| Task 6: Script Linting Verification | DONE | Pass `shellcheck`, `shfmt`, and verify zero Python references |
| Task 7: Raspberry Pi Emulator E2E Verification | DONE | Run emulator container, verify SSH, execute deployment, verify systemd active and HTTP 200 |
| Task 8: Task Tracking Completion | DONE | Mark all acceptance criteria and steps in Task 034 and task.md |
```

- [ ] **Step 3: Final commit**

```bash
git add docs/go_refactor/tasks/034-deployment_scripts_and_systemd.md docs/plans/task.md
git commit -m "docs: mark task 034 deployment scripts and systemd as complete"
```
