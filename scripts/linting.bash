#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")/.." && pwd)"
# shellcheck source=./lib/manage_docker
. "${ROOT_DIR}/scripts/lib/manage_docker"

usage() {
    cat <<'EOF'
Usage: scripts/linting.bash [--backend] [--frontend] [--scripts]

When one or more flags are provided, only the selected checks run and staged file detection is skipped.

Options:
    --backend   Run Python format/lint/type/tests for backend
    --frontend  Run JS/TS lint/format/tests for frontend
    --scripts   Run shellcheck and shfmt over scripts/*.bash
    -h, --help       Show this help and exit
EOF
}

log_info() { echo "INFO: $*"; }

run_python_tests() {
    local pyproject_path="${ROOT_DIR}/backend/pyproject.toml"
    start_docker
    log_info "running python tests..."
    uv run pytest --config-file "${pyproject_path}"
    stop_docker
}

run_python_checks() {
    local pyproject_path="${ROOT_DIR}/backend/pyproject.toml"
    cd "${ROOT_DIR}/backend"

    log_info "Running Ty..."
    uv run ty check
    log_info "Running ruff check --fix..."
    uv run ruff check --fix --config "${pyproject_path}"
    log_info "Running ruff format..."
    uv run ruff format --config "${pyproject_path}"
    run_python_tests
    cd -
}

run_generate_types() {
    echo "Generating OpenAPI spec"
    cd "${ROOT_DIR}/backend"
    export PYTHONPATH="${ROOT_DIR}/backend/src"
    uv run "${ROOT_DIR}/scripts/generate_openapi.py"
    cd "${ROOT_DIR}/frontend"
    log_info "Generating TypeScript types from OpenAPI spec"
    npx openapi-typescript "${ROOT_DIR}/openapi.json" --export-type --alphabetize --root-types --root-types-no-schema-prefix --output "${ROOT_DIR}/frontend/src/types/types.generated.ts"
}

build_frontend() {
    local tmp_dir
    tmp_dir="$(mktemp -d)"
    log_info "Building frontend"
    # we are building the frontend as a test to ensure there are no build errors
    npx vue-tsc -b
    npx vite build --outDir "${tmp_dir}"
    rm -r "${tmp_dir}"
}

run_frontend_checks() {
    run_generate_types
    cd "${ROOT_DIR}/frontend"
    log_info "Running JS/TS linter and formatter"
    npm run lint
    echo "Running JS/TS tests"
    npm run test
    build_frontend
    cd -

}

run_bash_checks() {
    echo "Running bash linter"
    shellcheck --shell=bash -x --exclude=SC1091 "${ROOT_DIR}/scripts"/*\.bash
    echo "Running bash formatter"
    shfmt --language-dialect bash --write -i 4 "${ROOT_DIR}/scripts"/*\.bash
}

main() {
    local needs_frontend=false
    local needs_backend=false
    local needs_scripts=false
    if [[ $# -gt 0 ]]; then
        while [[ $# -gt 0 ]]; do
            case "$1" in
            --backend)
                needs_backend=true
                ;;
            --frontend)
                needs_frontend=true
                ;;
            --scripts)
                needs_scripts=true
                ;;
            -h | --help)
                usage
                exit 0
                ;;
            *)
                echo "Unknown option: $1" >&2
                usage
                exit 1
                ;;
            esac
            shift
        done
    else
        staged_files=$(git diff --cached --name-only)
        for file in $staged_files; do
            if [[ $file =~ ^frontend/ ]]; then
                needs_frontend=true
            elif [[ $file =~ ^backend/ ]]; then
                needs_backend=true
            elif [[ $file =~ ^scripts/ ]]; then
                needs_scripts=true
            fi
        done
    fi
    if $needs_frontend; then
        run_frontend_checks
    fi

    if $needs_backend; then
        run_python_checks
    fi

    if $needs_scripts; then
        run_bash_checks
    fi
    exit 0
}

main "$@"
