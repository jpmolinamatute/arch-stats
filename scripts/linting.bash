#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")/.." && pwd)"
# shellcheck source=./lib/manage_docker
. "${ROOT_DIR}/scripts/lib/manage_docker"
# shellcheck source=./lib/logging
. "${ROOT_DIR}/scripts/lib/logging"

export PATH="${GOPATH:-$HOME/go}/bin:${PATH}"

usage() {
    cat <<'EOF'
Usage: scripts/linting.bash [--frontend] [--scripts] [--go]

When one or more flags are provided, only the selected checks run and staged file detection is skipped.

Options:
    --frontend  Run JS/TS lint/format/tests for frontend
    --scripts   Run shellcheck and shfmt over scripts/*.bash
    --go        Run Go linter and formatter for backend
    -h, --help       Show this help and exit
EOF
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
    "${ROOT_DIR}/scripts/generate_fe_types.bash"
    cd "${ROOT_DIR}/frontend"
    log_info "Running JS/TS linter and formatter"
    npm run lint
    log_info "Running JS/TS tests"
    npm run test
    build_frontend
    cd -

}

run_bash_checks() {
    log_info "Running bash linter"
    shellcheck --shell=bash -x --exclude=SC1091 "${ROOT_DIR}/scripts"/*\.bash
    log_info "Running bash formatter"
    shfmt --language-dialect bash --write -i 4 "${ROOT_DIR}/scripts"/*\.bash
}

run_go_checks() {
    cd "${ROOT_DIR}/backend"
    log_info "Running Go formatter (gofumpt)..."
    gofumpt -l -w .
    log_info "Running Go linter (golangci-lint)..."
    golangci-lint run --fix ./...
    log_info "Running Go tests..."
    go test ./... -count=1
    cd -
}

main() {
    local needs_frontend=false
    local needs_bash=false
    local needs_go=false
    if [[ $# -gt 0 ]]; then
        while [[ $# -gt 0 ]]; do
            case "$1" in
            --frontend)
                needs_frontend=true
                ;;
            --scripts)
                needs_bash=true
                ;;
            --go)
                needs_go=true
                ;;
            -h | --help)
                usage
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
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
            elif [[ $file =~ ^scripts/.*\.bash$ ]]; then
                needs_bash=true
            elif [[ $file =~ ^backend/.*\.go$ ]] || [[ $file =~ ^backend/go\.(mod|sum)$ ]]; then
                needs_go=true
            fi
        done
    fi
    if $needs_frontend; then
        run_frontend_checks
    fi

    if $needs_bash; then
        run_bash_checks
    fi

    if $needs_go; then
        run_go_checks
    fi
    exit 0
}

main "$@"
