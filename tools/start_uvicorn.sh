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

start_uvicorn() {
    echo "Starting Uvicorn server"
    source "${ROOT_DIR}/backend/.venv/bin/activate"
    cd "${ROOT_DIR}/backend/server"
    export PYTHONPATH="${ROOT_DIR}/backend"
    exec uvicorn --loop uvloop --lifespan on --reload --ws websockets --http h11 --use-colors --log-level debug --timeout-graceful-shutdown 10 --factory app:run

}

main() {
    check_docker
    start_uvicorn
}

main
