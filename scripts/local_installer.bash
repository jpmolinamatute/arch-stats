#!/usr/bin/env bash

# assumes that ssh is configured to connect to the raspberry pi on the local network using only a "host".

set -euo pipefail

SCRIPT_DIR="$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")"

check_remote() {
    local cred="${1}"
    echo "Checking SSH connection"
    if ssh -o BatchMode=yes -o ConnectTimeout=5 "${cred}" exit; then
        echo "Remote host '${cred}' is up and accepting connections"
    else
        echo "ERROR: Remote host '${cred}' is down or not accepting connections" >&2
        exit 1
    fi
}

upload_scripts() {
    local cred="${1}"
    shift
    local files=("$@")
    local missing=0

    for file in "${files[@]}"; do
        if [[ ! -f "$file" ]]; then
            echo "ERROR: File '$file' not found." >&2
            missing=1
        fi
    done
    if [[ ${missing} -ne 0 ]]; then
        echo "Aborting upload due to missing file(s)." >&2
        exit 1
    fi
    echo "Uploading ${#files[@]} file(s) to '${cred}:/tmp/'"
    scp -o BatchMode=yes -o ConnectTimeout=5 "${files[@]}" "${cred}":/tmp/
}

execute_remote_script() {
    local cred="${1}"
    local script_to_execute="${2}"
    echo "Executing remote installer on '${cred}'"
    # shellcheck disable=SC2029
    ssh "${cred}" "chmod +x ${script_to_execute} && ${script_to_execute}"
    # shellcheck disable=SC2029
    ssh "${cred}" "rm -f ${script_to_execute}"
}

main() {
    local cred
    local action
    if [[ $# -ne 2 ]]; then
        echo "Usage: $0 <remote-host> <action>" >&2
        exit 1
    fi
    cred="${1}"
    action="${2}"
    check_remote "${cred}"

    if [[ ${action} == "install" ]]; then
        echo "Starting remote installation on '${cred}'."
        upload_scripts "${cred}" "${SCRIPT_DIR}/remote_installer.bash" "${SCRIPT_DIR}/install.bash"
        execute_remote_script "${cred}" /tmp/remote_installer.bash
    elif [[ ${action} == "uninstall" ]]; then
        echo "Starting remote uninstallation on '${cred}'."
        upload_scripts "${cred}" "${SCRIPT_DIR}/remote_uninstaller.bash"
        execute_remote_script "${cred}" /tmp/remote_uninstaller.bash
    else
        echo "ERROR: invalid action '${action}'" >&2
        exit 2
    fi

    exit 0
}

main "$@"
