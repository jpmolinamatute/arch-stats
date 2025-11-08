#!/usr/bin/env bash

set -Eeuo pipefail

main(){
    local app_dir="${1:-}"
    local backend_dir="${app_dir}/backend"

    if [[ ! -d "$app_dir" ]]; then
        echo "ERROR: missing required argument: app_dir or app_dir is not a real directory" >&2
        echo "Usage: $0 /path/to/arch-stats" >&2
        exit 2
    fi

    
    if [[ ! -d "$backend_dir" ]]; then
        echo "ERROR: backend directory not found: $backend_dir" >&2
        exit 3
    fi

    cd "$backend_dir"

    echo "INFO: Running 'uv self update'"
    if ! uv self update; then
        echo "ERROR: uv self update failed" >&2
        exit 5
    fi

    echo "INFO: Syncing production dependencies (no dev, frozen)"
    if ! uv sync --no-dev --frozen --python "$(cat .python-version)"; then
        echo "ERROR: uv sync failed" >&2
        exit 6
    fi

    echo "INFO: Dependency installation completed successfully"
    exit 0
}

main "$@"
