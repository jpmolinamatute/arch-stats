#!/usr/bin/env bash

set -euo pipefail

SYSTEM_USER="archy"
SCRIPT_DIR="$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")"

create_user(){
    local user_home="/opt/${SYSTEM_USER}"
    if ! getent passwd "${SYSTEM_USER}" > /dev/null; then
        echo "Creating user ${SYSTEM_USER}..."
        sudo useradd --create-home --system --user-group --shell /bin/bash --home-dir "${user_home}" "${SYSTEM_USER}"
        echo "User ${SYSTEM_USER} created successfully."
    else
        echo "User ${SYSTEM_USER} already exists."
    fi
}

create_systemd_service_files(){
    echo "TODO"
}

activate_systemd_services(){
    echo "TODO"
}

install(){
    local install_file="${SCRIPT_DIR}/install.bash"
    if [[ -f "${install_file}" ]]; then
        chmod +x "${install_file}"
        sudo chown "${SYSTEM_USER}:${SYSTEM_USER}" "${install_file}"
        sudo --login --user="${SYSTEM_USER}" -- "${install_file}"
    else
        echo "ERROR: ${install_file} not found" >&2
        exit 1
    fi
    
}

main() {
    create_user
    install
    create_systemd_service_files
    activate_systemd_services
    exit 0
}

main "$@"
