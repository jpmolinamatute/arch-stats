#!/usr/bin/env bash

set -e

LOCAL_USER="arch-stats"
ROOT_DIR="$(dirname "$(realpath "$0")")"

print_out() {
    local msg="${1}"
    echo -e "\e[1;32m${msg}\e[0m"
}

print_err() {
    local msg="${1}"
    echo -e "\e[1;31m${msg}\e[0m" >&2
    exit 1
}

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
    print_out "Creating user ${LOCAL_USER}..."
    sudo useradd -rm -d "/opt/${LOCAL_USER}" "${LOCAL_USER}"
    sudo mkdir -p "/opt/${LOCAL_USER}/{backend,webui}"
    uuidgen -r >"/opt/${LOCAL_USER}/backend/uuid"
    sudo chown -R "${LOCAL_USER}:${LOCAL_USER}" "/opt/${LOCAL_USER}"
}

init_postgresql() {
    print_out "Initializing PostgreSQL..."
    sudo -u postgres psql -c "CREATE ROLE root WITH LOGIN SUPERUSER PASSWORD '***';"
    sudo -u postgres psql -c "CREATE USER \"${LOCAL_USER}\" WITH PASSWORD '***';"
    sudo -u postgres psql -c "CREATE DATABASE \"${LOCAL_USER}\" OWNER \"${LOCAL_USER}\";"
    if [[ -d ${ROOT_DIR}/data ]]; then
        for file in "${ROOT_DIR}"/data/*.sql; do
            print_out "Running ${file}..."
            sudo -u "${LOCAL_USER}" psql -d "${LOCAL_USER}" -f "${file}"
        done
    else
        print_err "Data directory not found"
    fi
}

main() {
    create_user
    install_postgresql
    init_postgresql
}

main

exit 0
