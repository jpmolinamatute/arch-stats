#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")"
TAR_FILE="arch-stats.tar.xz"
SYSTEM_USER="archy"



delete_user(){
    echo "Deleting user ${SYSTEM_USER}..."
    sudo userdel -r "${SYSTEM_USER}" || true
}


cleaning_files(){
    echo "Deleting files..."
    rm -f "${SCRIPT_DIR}/${TAR_FILE}"
}

stop_services(){
    echo "TODO"
}


main(){
    echo "Uninstalling arch-stats..."
    stop_services
    delete_user
    cleaning_files
    exit 0
}

main "$@"
