#!/usr/bin/env bash

set -eu

ROOT_DIR="$(dirname "$(realpath "$0")")"

# shellcheck source=./lib/helpers
. "${ROOT_DIR}/lib/helpers"

if [[ -f ${ROOT_DIR}/.env ]]; then
    # shellcheck source=./.env
    . "${ROOT_DIR}/.env"
fi

: "${ARCH_STATS_USER:?Environment variable ARCH_STATS_USER is not set}"
: "${ARCH_STATS_DIR:?Environment variable ARCH_STATS_DIR is not set}"
: "${ARCH_STATS_ID_FILE:?Environment variable ARCH_STATS_ID_FILE is not set}"
: "${DB_PASSWORD:?Environment variable DB_PASSWORD is not set}"
: "${DB_HOST:?Environment variable DB_HOST is not set}"
: "${DB_PORT:?Environment variable DB_PORT is not set}"
: "${POSTGRES_USER:?Environment variable POSTGRES_USER is not set}"
: "${POSTGRES_PASSWORD:?Environment variable POSTGRES_PASSWORD is not set}"

run() {
    local file="${1}"
    print_out "Running ${file}..."
    if [[ -n ${DEV} ]]; then
        PGPASSWORD=${DB_PASSWORD} psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${ARCH_STATS_USER}" -f "${file}"
    else
        sudo -u "${ARCH_STATS_USER}" psql -f "${file}"
    fi
}

create_root_user() {
    if [[ -z ${DEV} ]]; then
        print_out "Creating ${POSTGRES_USER} user..."
        sudo -u postgres psql -c "CREATE ROLE ${POSTGRES_USER} WITH LOGIN SUPERUSER PASSWORD '${POSTGRES_PASSWORD}';"
        sudo -u postgres psql -c "CREATE DATABASE ${POSTGRES_USER} OWNER ${POSTGRES_USER};"
    fi
}

create_app_user() {
    local create_user_sql="CREATE USER \"${ARCH_STATS_USER}\" WITH PASSWORD '${DB_PASSWORD}';"
    local create_database_sql="CREATE DATABASE \"${ARCH_STATS_USER}\" OWNER \"${ARCH_STATS_USER}\";"

    print_out "Creating ${ARCH_STATS_USER} user..."
    if [[ -n ${DEV} ]]; then
        PGPASSWORD=${POSTGRES_PASSWORD} psql -U "${POSTGRES_USER}" -h "${DB_HOST}" -p "${DB_PORT}" -c "${create_user_sql}"
        PGPASSWORD=${POSTGRES_PASSWORD} psql -U "${POSTGRES_USER}" -h "${DB_HOST}" -p "${DB_PORT}" -c "${create_database_sql}"
    else
        sudo -u postgres psql -c "${create_user_sql}"
        sudo -u postgres psql -c "${create_database_sql}"
    fi
}

init_postgresql() {
    print_out "Initializing PostgreSQL..."

    if [[ -d ${ROOT_DIR}/data ]]; then
        for file in "${ROOT_DIR}"/data/*.sql; do
            run "${file}"
        done
    else
        print_err "Data directory not found"
    fi
}

insert_target_track_data() {
    local my_uuid
    local insert_sql
    my_uuid=$(cat "${ARCH_STATS_DIR}/backend/${ARCH_STATS_ID_FILE}")
    insert_sql="INSERT INTO target_track (id, name) VALUES ('${my_uuid}', 'Raspberry Pi 1');"
    print_out "Inserting target_track data..."

    if [[ -n $DEV ]]; then
        PGPASSWORD=${DB_PASSWORD} psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${ARCH_STATS_USER}" -c "${insert_sql}"
    else
        sudo -u "${ARCH_STATS_USER}" psql -c "${insert_sql}"
    fi
}

main() {
    create_root_user
    create_app_user
    init_postgresql
    insert_target_track_data
}

main "$@"