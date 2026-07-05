#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")"
UNINSTALLER_SCRIPT="remote_uninstaller.bash"
APP_NAME="arch-stats"

. "${SCRIPT_DIR}/lib/logging"

usage() {
    cat <<EOF
Usage: $(basename "${0}") <remote-host> [<action>]
Description: Deploy or uninstall ${APP_NAME} remotely.

Arguments:
  <remote-host>   SSH host target (required)
  [<action>]      Option: "install" or "uninstall". Defaults to auto-resolving install/update.
EOF
}

# Pre-flight checklist to verify remote state
check_remote_action() {
    # @TODO: We need to check if the user and the DB are created as well.
    # @TODO: We need to check if cloudflared is configured correctly.
    # @TODO: We need to check if the .env file is present and has the required variables.

    local host="${1}"
    if ssh -q -o BatchMode=yes -o ConnectTimeout=5 "${host}" "
      id -u ${APP_NAME} >/dev/null 2>&1 && \
      command -v cloudflared >/dev/null 2>&1 && \
      command -v uv >/dev/null 2>&1 && \
      systemctl list-units --type=service | grep -q 'postgresql' && \
      [ -d "/opt/${APP_NAME}/backend" ]
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
    scp -o BatchMode=yes "${SCRIPT_DIR}/${UNINSTALLER_SCRIPT}" "${host}":/tmp/
    ssh -t "${host}" "sudo bash /tmp/${UNINSTALLER_SCRIPT}"
}

install() {
    local host="${1}"
    local temp_dir
    local cred_file
    local cert_file="${HOME}/.cloudflared/cert.pem"
    log_info "Starting clean installation on '${host}'..."

    : "${GITHUB_TOKEN:?Environment variable GITHUB_TOKEN must be set}"
    : "${ARCH_STATS_GOOGLE_OAUTH_CLIENT_ID:?Google OAuth Client ID is required for clean install}"
    : "${ARCH_STATS_JWT_SECRET:?JWT Secret is required for clean install}"
    : "${CLOUDFLARED_TUNNEL_ID:?Cloudflared Tunnel ID is required for clean install}"

    cred_file="${HOME}/.cloudflared/${CLOUDFLARED_TUNNEL_ID}.json"
    if [[ ! -f "${cred_file}" ]]; then
        log_error "Local Cloudflared credential file not found at: ${cred_file}"
        exit 1
    fi

    # Create temporary working dir

    temp_dir="$(mktemp -d)"

    # Render configurations
    uv run "${SCRIPT_DIR}/transform_templates.py" \
        --app-name "${APP_NAME}" \
        --app-name-label "${APP_NAME}" \
        --prod-uvicorn-port 8000 \
        --app-user-home-dir "/opt/${APP_NAME}" \
        --output-dir "${temp_dir}" \
        --cloudflared-tunnel-id "${CLOUDFLARED_TUNNEL_ID}"

    log_info "Uploading installation files to '${host}:/tmp'"
    # Upload everything
    scp -o BatchMode=yes \
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
        "${host}":/tmp/

    rm -rf "${temp_dir}"
    ssh -t "${host}" "sudo GITHUB_TOKEN='${GITHUB_TOKEN}' bash /tmp/remote_installer.bash '${APP_NAME}'"
}

update() {
    local host="${1}"

    # Require GITHUB_TOKEN for update
    : "${GITHUB_TOKEN:?Environment variable GITHUB_TOKEN must be set}"

    log_info "Starting update on '${host}'..."
    log_info "Uploading installation files to '${host}:/tmp'"
    scp -o BatchMode=yes "${SCRIPT_DIR}/install_app.bash" "${host}":/tmp/
    ssh -t "${host}" "
        sudo systemctl stop ${APP_NAME}.service && \
        sudo -u ${APP_NAME} GITHUB_TOKEN='${GITHUB_TOKEN}' bash /tmp/install_app.bash /opt/${APP_NAME} && \
        sudo systemctl start ${APP_NAME}.service
    "
}

check_connection() {
    local host="${1}"
    # Check SSH connectivity
    log_info "Checking connection to '${host}'..."
    if ! ssh -o BatchMode=yes -o ConnectTimeout=5 "${host}" exit 2>/dev/null; then
        log_error "ERROR: Cannot connect to remote host: ${host}"
        exit 1
    fi
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

    check_connection "${host}"

    # Resolve action
    if [[ -z "${action}" || "${action}" == "install" ]]; then
        action=$(check_remote_action "${host}")
        log_info "Resolved deployment action: ${action}"
        if [[ "${action}" == "install" ]]; then
            install "${host}"
        elif [[ "${action}" == "update" ]]; then
            update "${host}"
        fi
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
