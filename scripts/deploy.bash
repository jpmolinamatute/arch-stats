#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")"
UNINSTALLER_SCRIPT="remote_uninstaller.bash"
APP_NAME="arch-stats"
REMOTE_SECURE_DIR="/tmp/deploy_assets"
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

upload_install_assets() {
    local host="${1}"
    local temp_dir="${2}"
    local cred_file="${3}"
    local cert_file="${4}"
    local env_temp_file="${temp_dir}/env"

    log_info "Uploading installation assets to '${host}:${REMOTE_SECURE_DIR}'..."
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
        "${env_temp_file}" \
        "${host}:${REMOTE_SECURE_DIR}/"
}

upload_update_assets() {
    local host="${1}"
    local temp_dir="${2}"
    local env_temp_file="${temp_dir}/env"

    log_info "Uploading update assets to '${host}:${REMOTE_SECURE_DIR}'..."
    scp -o BatchMode=yes \
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

    # Render configurations
    uv run "${SCRIPT_DIR}/transform_templates.py" \
        --app-name "${APP_NAME}" \
        --app-name-label "${APP_NAME}" \
        --prod-uvicorn-port 8000 \
        --app-user-home-dir "/opt/${APP_NAME}" \
        --output-dir "${temp_dir}" \
        --cloudflared-tunnel-id "${CLOUDFLARED_TUNNEL_ID}"

    # Upload everything securely
    upload_install_assets "${host}" "${temp_dir}" "${cred_file}" "${cert_file}"
    ssh -t "${host}" "sudo bash -c '${REMOTE_SECURE_DIR}/remote_installer.bash \"${APP_NAME}\"'"
}

update() {
    local host="${1}"
    local temp_dir="${2}"

    log_info "Starting update on '${host}'..."
    upload_update_assets "${host}" "${temp_dir}"
    ssh -t "${host}" "
        sudo systemctl stop ${APP_NAME}.service && \
        (
            sudo -u ${APP_NAME} bash -c '
                source ${REMOTE_SECURE_DIR}/env
                bash ${REMOTE_SECURE_DIR}/install_app.bash /opt/${APP_NAME}
            ' && \
            sudo systemctl start ${APP_NAME}.service
        )
        ec=\$?
        sudo rm -rf ${REMOTE_SECURE_DIR}
        exit \$ec
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
    local temp_dir
    local env_temp_file
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
        temp_dir="$(mktemp -d)"
        env_temp_file="${temp_dir}/env"
        : "${GITHUB_TOKEN:?Environment variable GITHUB_TOKEN must be set}"
        echo "export GITHUB_TOKEN='${GITHUB_TOKEN}'" >"${env_temp_file}"
        chmod 400 "${env_temp_file}"
        log_info "Creating secure remote temporary directory..."
        ssh -o BatchMode=yes "${host}" "mkdir -p -m 700 ${REMOTE_SECURE_DIR}"
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
