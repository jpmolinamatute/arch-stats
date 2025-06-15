#!/usr/bin/env bash

set -e
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && cd .. && pwd)"

check_docker() {
    echo "Checking Docker Compose"
    docker_file="${ROOT_DIR}/docker/docker-compose.yaml"
    running=$(docker compose -f "${docker_file}" ps -q)
    if [[ -z ${running} ]]; then
        docker compose -f "${docker_file}" up --build -d
        sleep 2
    fi

}

start_archy() {
    echo "Starting Archy server"
    source "${ROOT_DIR}/backend/.venv/bin/activate"
    cd "${ROOT_DIR}/backend/target_reader"
    export PYTHONPATH="${ROOT_DIR}/backend"
    exec "${ROOT_DIR}/backend/target_reader/archy.py"
}

main() {
    check_docker
    start_archy
}

main
