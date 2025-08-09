#!/usr/bin/env bash

set -eu
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && cd .. && pwd)"
# shellcheck source=./lib/manage_docker
. "${ROOT_DIR}/scripts/lib/manage_docker"

start_uvicorn() {
    echo "Starting Uvicorn server"
    cd "${ROOT_DIR}/backend/src"
    # shellcheck source=../backend/.venv/bin/activate
    source "${ROOT_DIR}/backend/.venv/bin/activate"
    exec uvicorn --loop uvloop --lifespan on --reload --ws websockets --http h11 --use-colors --log-level debug --timeout-graceful-shutdown 10 --factory --limit-concurrency 10 server.app:run
}

main() {
    start_docker
    start_uvicorn
}

main
