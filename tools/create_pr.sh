#!/usr/bin/env bash

set -euo pipefail

skip_main() {
    local branch="${1}"
    if [[ ${branch} =~ ^main$ ]]; then
        echo "🚫 Skipping PR creation for protected branch: '${branch}'"
        exit 0
    fi
}

check_pr() {
    local branch="${1}"
    if gh pr view "${branch}" &>/dev/null; then
        echo "✅ PR already exists for '${branch}'"
        exit 0
    fi
}

create_pr() {
    local branch="${1}"
    local changed_files
    local labels_to_add=()

    changed_files=$(git diff --name-only origin/main...HEAD)
    if echo "${changed_files}" | grep -q '^frontend/'; then
        labels_to_add+=("--label" "frontend")
    fi
    if echo "${changed_files}" | grep -q '^backend/'; then
        labels_to_add+=("--label" "backend")
    fi
    if echo "${changed_files}" | grep -q '\.md$'; then
        labels_to_add+=("--label" "docs")
    fi

    echo "📦 Creating pull request for branch '${branch}'..."
    if [[ -n "${labels_to_add[*]}" ]]; then
        gh pr create --project "Arch Stats" --assignee "@me" --base main --milestone "MVP with Dummy Data" --head "${branch}" --fill "${labels_to_add[*]}"
    else
        gh pr create --project "Arch Stats" --assignee "@me" --base main --milestone "MVP with Dummy Data" --head "${branch}" --fill
    fi
    
}

main() {
    local branch
    branch=$(git rev-parse --abbrev-ref HEAD)

    skip_main "${branch}"
    check_pr "${branch}"
    create_pr "${branch}"
}





main

exit 0
