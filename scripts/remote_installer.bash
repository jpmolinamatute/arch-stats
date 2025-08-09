#!/usr/bin/env bash

# Purpose: Create system user, run install.bash as that user, and set up systemd service
# Usage (as root): remote_installer.bash <db_user> <db_name> [db_socket_dir=/var/run/postgresql] [server_port=8000]

set -euo pipefail

SYSTEM_USER="archy"
SCRIPT_DIR="$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")"

get_user_home() {
    getent passwd "${SYSTEM_USER}" | cut -d: -f6
}

create_user() {
    local user_home="/opt/${SYSTEM_USER}"
    if ! getent passwd "${SYSTEM_USER}" >/dev/null; then
        echo "Creating user ${SYSTEM_USER} with home ${user_home}..."
        useradd --create-home --system --user-group --shell /bin/bash --home-dir "${user_home}" "${SYSTEM_USER}"
        echo "User ${SYSTEM_USER} created."
    else
        echo "User ${SYSTEM_USER} already exists."
    fi
}

create_systemd_service_file() {
    local user_home
    user_home="$(get_user_home)"
    local svc="/etc/systemd/system/arch-stats.service"

    echo "Writing ${svc}..."
    cat >"${svc}" <<EOF
[Unit]
Description=Arch-Stats FastAPI Server
After=network.target

[Service]
Type=simple
User=${SYSTEM_USER}
Group=${SYSTEM_USER}
WorkingDirectory=${user_home}/backend/src
EnvironmentFile=${user_home}/backend/.env
# uvicorn discovered via PATH from EnvironmentFile
ExecStart=${user_home}/backend/.venv/bin/uvicorn --loop uvloop --lifespan on --ws websockets --http h11 --use-colors --log-level info --timeout-graceful-shutdown 10 --factory --limit-concurrency 10 server.app:run
Restart=on-failure
RestartSec=5
KillSignal=SIGINT
TimeoutStopSec=20

[Install]
WantedBy=multi-user.target
EOF
    echo "Service file created."
}

activate_systemd_service() {
    echo "Reloading systemd and enabling service..."
    systemctl daemon-reload
    systemctl enable --now arch-stats.service
    systemctl --no-pager --full status arch-stats.service || true
    echo "Tail logs with: journalctl -u arch-stats.service -f"
}

install_as_system_user() {
    local install_file="${SCRIPT_DIR}/install.bash"
    if [[ ! -f "${install_file}" ]]; then
        echo "ERROR: ${install_file} not found" >&2
        exit 1
    fi

    chmod +x "${install_file}"
    chown "${SYSTEM_USER}:${SYSTEM_USER}" "${install_file}"

    # Forward DB params to install.bash
    local db_user="${1:?db_user required}"
    local db_name="${2:?db_name required}"
    local db_socket_dir="${3:-/var/run/postgresql}"
    local server_port="${4:-8000}"

    echo "Running install as ${SYSTEM_USER}..."
    sudo --login --user="${SYSTEM_USER}" -- "${install_file}" "${db_user}" "${db_name}" "${db_socket_dir}" "${server_port}"
}

main() {
    if [[ $EUID -ne 0 ]]; then
        echo "Please run as root." >&2
        exit 1
    fi

    if [[ $# -lt 2 ]]; then
        echo "Usage: $0 <db_user> <db_name> [db_socket_dir=/var/run/postgresql] [server_port=8000]" >&2
        exit 2
    fi

    create_user
    install_as_system_user "$@"
    create_systemd_service_file
    activate_systemd_service
    echo "Remote installation completed."
    exit 0
}

main "$@"
