#!/usr/bin/env bash

# This script is meant to be run on a local machine to install arch-stats in a remote machine,
# such as a Raspberry Pi, over SSH. This script assumes that ssh is configured to connect to the
# raspberry pi using only a "host".

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
    echo "Executing remote installer on '${cred}' as root"
    # shellcheck disable=SC2029
    ssh "${cred}" "chmod +x ${script_to_execute} && sudo GITHUB_TOKEN='${GITHUB_TOKEN}' ${script_to_execute}"
    # shellcheck disable=SC2029
    ssh "${cred}" "rm -f ${script_to_execute}"
}

gather_clouldflared_config(){
    local temp_dir="${1}"
    local tunnel_id=""
    local cred_dir=""
    local cred_file
    local file_tunnel_id
    
    # Prompt for Cloudflared Tunnel ID
    while [[ -z "${tunnel_id}" ]]; do
        read -rp "Enter Cloudflared Tunnel ID: " tunnel_id
    done

    # Prompt for Cloudflared Credentials directory
    while [[ -z "${cred_dir}" ]]; do
        read -rp "Enter local directory containing Cloudflared credentials [${tunnel_id}.json]: " cred_dir
    done
    cred_file="${cred_dir}/${tunnel_id}.json"
    if [[ ! -f "${cred_file}" ]]; then
        echo "ERROR: Local credential file not found at: ${cred_file}" >&2
        exit 1
    fi
    file_tunnel_id="$(jq -r '.TunnelID // empty' "${cred_file}")"
    if [[ "${file_tunnel_id}" != "${tunnel_id}" ]]; then
        echo "ERROR: Local validation failed: TunnelID in ${cred_file} is '${file_tunnel_id}', expected '${tunnel_id}'" >&2
        exit 1
    fi
    remote_user_dir="delete_me"
    echo "Generating cloudflared configuration file..."
    sed -e "s|\[CLOUDFLARED_TUNNEL_ID\]|${tunnel_id}|g" \
        -e "s|\[PROD_UVICORN_PORT\]|8000|g" \
        -e "s|\[APP_USER_HOME_DIR\]|${remote_user_dir}|g" \
        "${SCRIPT_DIR}/../OS/cloudflared/config_template.yaml" >"${temp_dir}/config.yml"
}


install() {
    local cred="${1}"
    local temp_dir

    echo "Starting remote installation on '${cred}'."
    temp_dir="$(mktemp -d)"

    gather_clouldflared_config "${temp_dir}"

    echo "Uploading scripts and configs to remote..."
    upload_scripts "${cred}" \
        "${SCRIPT_DIR}/remote_installer.bash" \
        "${SCRIPT_DIR}/install_app.bash" \
        "${SCRIPT_DIR}/../OS/cloudflared/cloudflared.service" \
        "${SCRIPT_DIR}/../OS/cloudflared/cloudflared-update.service" \
        "${SCRIPT_DIR}/../OS/cloudflared/cloudflared-update.timer" \
        "${temp_dir}/config.yml" \
        "${cred_file}"

    rm -rf "${temp_dir}"

    execute_remote_script "${cred}" /tmp/remote_installer.bash
}

main() {
    local cred
    local action

    if [[ ! -f "${SCRIPT_DIR}/.env" ]]; then
        echo "ERROR: .env file not found at ${SCRIPT_DIR}/.env" >&2
        exit 1
    fi
    # shellcheck source=../.env
    source "${SCRIPT_DIR}/.env"

    : "${GITHUB_TOKEN:?Environment variable GITHUB_TOKEN is not set}"

    if [[ $# -ne 2 ]]; then
        echo "Usage: $0 <remote-host> <action>" >&2
        exit 1
    fi
    cred="${1}"
    action="${2}"

    check_remote "${cred}"

    if [[ ${action} == "install" ]]; then
        install "${cred}"
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
