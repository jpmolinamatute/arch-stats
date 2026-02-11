#!/usr/bin/env bash

set -euo pipefail
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && cd .. && pwd)"
# shellcheck source=./lib/manage_docker
. "${ROOT_DIR}/scripts/lib/manage_docker"

start_uvicorn() {
    # shellcheck source=../.env
    . "${ROOT_DIR}/.env"
    echo "Starting Uvicorn server"
    cd "${ROOT_DIR}/backend"
    export PYTHONPATH="${ROOT_DIR}/backend/src"
    exec uv run uvicorn --loop uvloop --lifespan on --reload --ws websockets --http h11 --use-colors --log-level debug --timeout-graceful-shutdown 10 --factory --limit-concurrency 10 --port "${ARCH_STATS_SERVER_PORT}" app:run
}

main() {
    start_docker
    start_uvicorn
    exit 0
}

main
