#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")/.." && pwd)"
OPENAPI_JSON="${ROOT_DIR}/openapi.json"


backend(){
    cd "${ROOT_DIR}/backend"
    export PYTHONPATH="${ROOT_DIR}/backend/src"
    uv run ./tools/generate_openapi.py "${OPENAPI_JSON}"
    cd "${ROOT_DIR}"
}

frontend(){
    cd "${ROOT_DIR}/frontend"
    npx openapi-typescript "${OPENAPI_JSON}" --export-type --output src/types/types.generated.ts
    cd "${ROOT_DIR}"
}

main(){
    backend
    frontend
 exit 0
}

main
