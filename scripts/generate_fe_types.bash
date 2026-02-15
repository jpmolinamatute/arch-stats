#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")/.." && pwd)"
ENV_FILE="${ROOT_DIR}/.env"

openapi_via_script() {
    local openapi_source
    openapi_source="$1"
    cd "${ROOT_DIR}/backend"
    export PYTHONPATH="${ROOT_DIR}/backend:${ROOT_DIR}/backend/src"
    uv run ./tools/generate_openapi.py "${openapi_source}"
    cd "${ROOT_DIR}"
}

frontend() {
    local openapi_source
    openapi_source="$1"
    cd "${ROOT_DIR}/frontend"
    npx openapi-typescript "${openapi_source}" --export-type --output src/types/types.generated.ts
    cd "${ROOT_DIR}"
}

main() {
    local count openapi_source
    if [[ ! -f "${ENV_FILE}" ]]; then
        echo "Error: .env file not found at ${ENV_FILE}." >&2
        exit 1
    fi
    # shellcheck source=../.env
    source "${ENV_FILE}"
    openapi_source="http://localhost:${ARCH_STATS_SERVER_PORT}/api/openapi.json"
    count=$(pgrep -af uvicorn | wc -l)
    if [[ "$count" -ne 2 ]]; then
        openapi_source="${ROOT_DIR}/openapi.json"
        openapi_via_script "${openapi_source}"
    fi
    frontend "${openapi_source}"
    exit 0
}

main
