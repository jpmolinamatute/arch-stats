#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")/.." && pwd)"
ENV_FILE="${ROOT_DIR}/.env"

openapi_via_script() {
    local openapi_source="$1"
    (
        cd "${ROOT_DIR}/backend"
        export PYTHONPATH="${ROOT_DIR}/backend:${ROOT_DIR}/backend/src"
        echo "Info: Generating OpenAPI spec from script"
        uv run ./tools/generate_openapi.py "${openapi_source}"
    )
}

frontend() {
    local openapi_source="$1"
    (
        cd "${ROOT_DIR}/frontend"
        echo "Info: Generating frontend types from ${openapi_source}"
        npx openapi-typescript "${openapi_source}" --export-type --output src/types/types.generated.ts
    )
}

main() {
    local openapi_source
    if [[ -f "${ENV_FILE}" ]]; then
        # shellcheck source=../.env
        source "${ENV_FILE}"
    fi

    openapi_source="http://localhost:${ARCH_STATS_SERVER_PORT:-8000}/api/openapi.json"
    if ! curl --silent --fail --head "${openapi_source}" >/dev/null 2>&1; then
        openapi_source="${ROOT_DIR}/openapi.json"
        openapi_via_script "${openapi_source}"
    fi
    frontend "${openapi_source}"
    exit 0
}

main
