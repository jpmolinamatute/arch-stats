#!/usr/bin/env bash

set -e

LOCAL_USER="arch-stats"
LOCAL_HOME="/opt/${LOCAL_USER}"
LOCAL_WEBUI="${LOCAL_HOME}/webui"
LOCAL_BACKEND="${LOCAL_HOME}/backend"
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
    local my_uuid="${1}"
    print_out "Creating user ${LOCAL_USER}..."
    sudo useradd -rm -d "${LOCAL_HOME}" "${LOCAL_USER}"
    sudo mkdir -p "${LOCAL_WEBUI}" "${LOCAL_BACKEND}"
    echo "${my_uuid}" >"${LOCAL_BACKEND}/arch-stats-id"
    sudo chown -R "${LOCAL_USER}:${LOCAL_USER}" "${LOCAL_HOME}"
}

init_postgresql() {
    local my_uuid="${1}"
    print_out "Initializing PostgreSQL..."
    sudo -u postgres psql -c "CREATE ROLE root WITH LOGIN SUPERUSER PASSWORD '***';"
    sudo -u postgres psql -c "CREATE USER \"${LOCAL_USER}\" WITH PASSWORD '***';"
    sudo -u postgres psql -c "CREATE DATABASE \"${LOCAL_USER}\" OWNER \"${LOCAL_USER}\";"
    if [[ -d ${ROOT_DIR}/data ]]; then
        for file in "${ROOT_DIR}"/data/*.sql; do
            print_out "Running ${file}..."
            sudo -u "${LOCAL_USER}" psql -d "${LOCAL_USER}" -f "${file}"
        done
        # we need to insert at least one row to avoid errors when adding shooting data
        sudo -u "${LOCAL_USER}" psql -d "${LOCAL_USER}" -c "INSERT INTO target_track (id, name) VALUES ('${my_uuid}', 'prototype');"
    else
        print_err "Data directory not found"
    fi
}

main() {
    local my_uuid
    my_uuid=$(uuidgen -r)
    create_user "${my_uuid}"
    install_postgresql
    init_postgresql "${my_uuid}"
}

main

exit 0
