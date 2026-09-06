#!/usr/bin/env bash

set -eu

SCRIPT_DIR="$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")"
UNINSTALLER_SCRIPT="remote_uninstaller.bash"
APP_NAME="arch-stats"
REMOTE_SECURE_DIR="/tmp/deploy_assets"
# shellcheck source=scripts/lib/logging disable=SC1091
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
        -e "s|{{ server_port }}|${server_port}|g" \
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
    # shellcheck disable=SC2029
    ssh -t "${SSH_OPTS[@]}" "${host}" "sudo bash -c '${REMOTE_SECURE_DIR}/remote_installer.bash \"${APP_NAME}\"'"
}

update() {
    local host="${1}"
    local temp_dir="${2}"

    log_info "Starting update on '${host}'..."
    upload_update_assets "${host}" "${temp_dir}"
    # shellcheck disable=SC2029
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
        # shellcheck disable=SC2029
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
