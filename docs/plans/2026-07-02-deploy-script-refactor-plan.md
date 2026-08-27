# Deploy Script Refactor Implementation Plan

> **For Antigravity:** REQUIRED WORKFLOW: Use `.agent/workflows/execute-plan.md` to execute this plan in single-flow mode.

**Goal:** Refactor `scripts/deploy.bash` to break down its large `main` function into `uninstall`, `install`, and `update` helper functions.

**Architecture:** Extraction of logic blocks from `main` into separate, self-contained functions that receive `host` as an argument and perform their own environment checks.

**Tech Stack:** Bash shell scripting.

---

### Task 1: Refactor deploy.bash helper functions and main

**Files:**
- Modify: `scripts/deploy.bash` (line 35 onwards)

**Step 1: Extract functions and update main**
Update the file `scripts/deploy.bash` to add `uninstall()`, `install()`, and `update()`, and update `main()` to invoke them.

```bash
uninstall() {
    local host="${1}"
    echo "Starting remote uninstallation on '${host}'..."
    scp -o BatchMode=yes "${SCRIPT_DIR}/${UNINSTALLER_SCRIPT}" "${host}":/tmp/
    ssh -t "${host}" "sudo chmod +x /tmp/${UNINSTALLER_SCRIPT} && sudo /tmp/${UNINSTALLER_SCRIPT}"
}

install() {
    local host="${1}"

    # Require GITHUB_TOKEN for install
    : "${GITHUB_TOKEN:?Environment variable GITHUB_TOKEN must be set}"

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
        --app-name "${APP_NAME}" \
        --app-name-label "${APP_NAME}" \
        --prod-uvicorn-port 8000 \
        --app-user-home-dir "/opt/${APP_NAME}" \
        --output-dir "${temp_dir}" \
        --cloudflared-tunnel-id "${CLOUDFLARED_TUNNEL_ID}"

    # Generate remote .env
    cat <<EOF >"${temp_dir}/.env"
POSTGRES_USER=${APP_NAME}
POSTGRES_DB=${APP_NAME}
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
        "${temp_dir}/${APP_NAME}.service" \
        "${temp_dir}/.env" \
        "${cred_file}" \
        "${host}":/tmp/

    rm -rf "${temp_dir}"

    ssh -t "${host}" "sudo chmod +x /tmp/remote_installer.bash && sudo GITHUB_TOKEN='${GITHUB_TOKEN}' /tmp/remote_installer.bash"
}

update() {
    local host="${1}"

    # Require GITHUB_TOKEN for update
    : "${GITHUB_TOKEN:?Environment variable GITHUB_TOKEN must be set}"

    echo "Starting update on '${host}'..."
    scp -o BatchMode=yes "${SCRIPT_DIR}/install_app.bash" "${host}":/tmp/
    ssh -t "${host}" "
        sudo systemctl stop ${APP_NAME}.service && \
        sudo -u ${APP_NAME} GITHUB_TOKEN='${GITHUB_TOKEN}' bash /tmp/install_app.bash /opt/${APP_NAME} && \
        sudo systemctl start ${APP_NAME}.service
    "
}

main() {
    if [[ $# -lt 1 ]]; then
        usage
        exit 1
    fi

    local host="${1}"
    local action="${2:-}"

    # Load environment variables
    if [[ -f "${SCRIPT_DIR}/../.env" ]]; then
        # shellcheck source=/dev/null
        source "${SCRIPT_DIR}/../.env"
    fi

    # Check SSH connectivity
    if ! ssh -o BatchMode=yes -o ConnectTimeout=5 "${host}" exit 2>/dev/null; then
        echo "ERROR: Cannot connect to remote host: ${host}" >&2
        exit 1
    fi

    # Resolve action
    if [[ -z "${action}" || "${action}" == "install" ]]; then
        action=$(check_remote_action "${host}")
        echo "Resolved deployment action: ${action}"
    elif [[ "${action}" != "uninstall" ]]; then
        echo "ERROR: Invalid action: ${action}" >&2
        usage
        exit 1
    fi

    if [[ "${action}" == "uninstall" ]]; then
        uninstall "${host}"
    elif [[ "${action}" == "install" ]]; then
        install "${host}"
    elif [[ "${action}" == "update" ]]; then
        update "${host}"
    fi

    exit 0
}
```

---

### Task 2: Verification

**Files:**
- Test: Run script syntax check and formatter: `shellcheck` & `shfmt` via `./scripts/linting.bash --scripts`

**Step 1: Run linter/formatter**
Run: `./scripts/linting.bash --scripts`
Expected: Output showing bash checks passing with zero errors.
