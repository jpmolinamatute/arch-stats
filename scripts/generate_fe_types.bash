#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")/.." && pwd)"
ENV_FILE="${ROOT_DIR}/.env"

convert_swagger_to_openapi() {
    local src="$1"
    local dst="$2"
    (
        cd "${ROOT_DIR}/frontend"
        echo "Info: Converting Swagger 2.0 spec (${src}) to OpenAPI 3.0 (${dst})"
        npx --yes swagger2openapi "${src}" -o "${dst}" -p
    )
    if [[ -f "${ROOT_DIR}/scripts/enrich_openapi.py" ]]; then
        echo "Info: Enriching OpenAPI 3.0 spec for frontend compatibility"
        python3 "${ROOT_DIR}/scripts/enrich_openapi.py" "${dst}"
    fi
}

openapi_via_script() {
    local openapi_source="$1"
    (
        cd "${ROOT_DIR}/backend-old"
        export PYTHONPATH="${ROOT_DIR}/backend-old:${ROOT_DIR}/backend-old/src"
        echo "Info: Generating OpenAPI spec from script"
        uv run ./tools/generate_openapi.py "${openapi_source}"
    )
}

frontend() {
    local openapi_source="$1"
    (
        cd "${ROOT_DIR}/frontend"
        if grep -q '"swagger":' "${openapi_source}" 2>/dev/null; then
            local v3_target="${ROOT_DIR}/openapi.json"
            convert_swagger_to_openapi "${openapi_source}" "${v3_target}"
            openapi_source="${v3_target}"
        fi
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
        if command -v swag >/dev/null 2>&1 && [[ -d "${ROOT_DIR}/backend" ]]; then
            (
                cd "${ROOT_DIR}/backend"
                echo "Info: Regenerating Swagger 2.0 specs via swag"
                swag init -g cmd/arch-stats/main.go -o specs/ --parseInternal --useStructName --requiredByDefault >/dev/null 2>&1 || true
            )
        fi
        if [[ -f "${ROOT_DIR}/backend/specs/swagger.json" ]]; then
            openapi_source="${ROOT_DIR}/openapi.json"
            convert_swagger_to_openapi "${ROOT_DIR}/backend/specs/swagger.json" "${openapi_source}"
        elif [[ -f "${ROOT_DIR}/backend-old/tools/generate_openapi.py" ]]; then
            openapi_source="${ROOT_DIR}/openapi.json"
            openapi_via_script "${openapi_source}"
        else
            openapi_source="${ROOT_DIR}/openapi.json"
        fi
    fi
    frontend "${openapi_source}"
    exit 0
}

main
