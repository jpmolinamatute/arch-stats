#!/usr/bin/env bash

set -e

ARCH_STATS_USER="arch-stats"
ARCH_STATS_DIR="/opt/${ARCH_STATS_USER}"
ARCH_STATS_WEBUI_PATH="${ARCH_STATS_DIR}/webui"
ARCH_STATS_BACKEND_PATH="${ARCH_STATS_DIR}/backend"
ARCH_STATS_ID_FILE="arch-stats-id"
ROOT_DIR="$(dirname "$(realpath "$0")")"

# shellcheck source=./lib/helpers
. "${ROOT_DIR}/lib/helpers"

install_postgresql() {
    print_out "Installing PostgreSQL..."
    sudo apt install -y postgresql
    if [[ -d ${ROOT_DIR}/data ]]; then
        print_out "Copying PostgreSQL config files..."
        sudo cp -f "${ROOT_DIR}/data/pg_hba.conf" "${ROOT_DIR}/data/postgresql.conf" /etc/postgresql/15/main/
        sudo chown postgres:postgres /etc/postgresql/15/main/pg_hba.conf /etc/postgresql/15/main/postgresql.conf
    else
        print_err "Data directory not found"
    fi
    sudo systemctl enable --now postgresql
}

create_user() {
    local my_uuid
    my_uuid=$(uuidgen -r)
    print_out "Creating user ${ARCH_STATS_USER}..."
    sudo useradd -rm -d "${ARCH_STATS_DIR}" "${ARCH_STATS_USER}"
    sudo mkdir -p "${ARCH_STATS_WEBUI_PATH}" "${ARCH_STATS_BACKEND_PATH}"
    echo "${my_uuid}" >"${ARCH_STATS_BACKEND_PATH}/${ARCH_STATS_ID_FILE}"
    sudo chown -R "${ARCH_STATS_USER}:${ARCH_STATS_USER}" "${ARCH_STATS_DIR}"
}

main() {
    create_user
    install_postgresql
    "${ROOT_DIR}/db-init.sh" "${ARCH_STATS_USER}"
}

main

exit 0
