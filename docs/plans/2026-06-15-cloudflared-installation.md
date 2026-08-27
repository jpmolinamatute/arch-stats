# Cloudflared Installation Implementation Plan

> **For Antigravity:** REQUIRED WORKFLOW: Use `.agent/workflows/execute-plan.md` to execute this plan in single-flow mode.

**Goal:** Extend the local and remote installer scripts to ask for the Cloudflared Tunnel ID and credentials directory locally in `local_installer.bash`, validate the credentials file locally, generate the `config.yml` configuration file locally (keeping the remote home directory placeholder), upload all configuration/credential/service files to the remote machine, and configure and start the cloudflared service remotely.

**Architecture:** 
1. Modify `scripts/local_installer.bash` to:
   - Prompt for `tunnel_id` and the local directory containing the credentials file.
   - Validate that the credentials file exists locally and has the correct `TunnelID` key and value.
   - Generate `config.yml` locally from the template `config_template.yaml` substituting `[CLOUDFLARED_TUNNEL_ID]` and `[PROD_UVICORN_PORT]` (leaving `[APP_USER_HOME_DIR]` as a placeholder).
   - Upload the local validated `[TUNNEL_ID].json` credentials file, the generated `config.yml`, and the systemd service files to `/tmp/` on the remote machine.
2. Modify `scripts/remote_installer.bash` to check if a credentials JSON file exists in `/tmp/`:
   - If it exists, extract the Tunnel ID from the filename.
   - Determine the remote home directory of the `arch-stats` system user.
   - Copy the credentials file to the app user's home directory under `.cloudflared/` with correct permissions (700 for the folder, 600 for the file, owned by `arch-stats:arch-stats`).
   - Copy `/tmp/config.yml` to `/etc/cloudflared/config.yml` and replace the `[APP_USER_HOME_DIR]` placeholder with the resolved home directory path.
   - Copy service and timer unit files from `/tmp` to `/etc/systemd/system/`.
   - Enable and restart the systemd services (`cloudflared.service`, `cloudflared-update.timer`).
   - Clean up uploaded files in `/tmp`.

**Tech Stack:** Bash, systemd, cloudflared, jq, sed

---

### Task 1: Update local_installer.bash

**Files:**
- Modify: `scripts/local_installer.bash`

**Step 1: Write code changes**

We will update `scripts/local_installer.bash` to prompt for Cloudflared credentials, perform validation locally using `jq`, generate the configuration file locally in a temporary folder, and upload all files to the remote host.

```bash
# In scripts/local_installer.bash main function:
    local tunnel_id=""
    local cred_dir=""

    if [[ ${action} == "install" ]]; then
        echo "Starting remote installation on '${cred}'."

        # Prompt for Cloudflared Tunnel ID
        while [[ -z "${tunnel_id}" ]]; do
            read -rp "Enter Cloudflared Tunnel ID: " tunnel_id
        done

        # Prompt for Cloudflared Credentials directory
        while [[ -z "${cred_dir}" ]]; do
            read -rp "Enter local directory containing Cloudflared credentials [${tunnel_id}.json]: " cred_dir
        done

        local cred_file="${cred_dir}/${tunnel_id}.json"
        if [[ ! -f "${cred_file}" ]]; then
            echo "ERROR: Local credential file not found at: ${cred_file}" >&2
            exit 1
        fi

        if ! command -v jq >/dev/null 2>&1; then
            echo "ERROR: jq is required locally but not installed." >&2
            exit 1
        fi

        local file_tunnel_id
        file_tunnel_id="$(jq -r '.TunnelID // empty' "${cred_file}")"
        if [[ "${file_tunnel_id}" != "${tunnel_id}" ]]; then
            echo "ERROR: Local validation failed: TunnelID in ${cred_file} is '${file_tunnel_id}', expected '${tunnel_id}'" >&2
            exit 1
        fi

        local temp_dir
        temp_dir="$(mktemp -d)"

        echo "Generating cloudflared configuration file..."
        sed -e "s|\[CLOUDFLARED_TUNNEL_ID\]|${tunnel_id}|g" \
            -e "s|\[PROD_UVICORN_PORT\]|8000|g" \
            "${SCRIPT_DIR}/../OS/cloudflared/config_template.yaml" > "${temp_dir}/config.yml"

        echo "Copying and uploading scripts/configs..."
        
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
```

**Step 2: Run verification**

Run shellcheck and shfmt on `scripts/local_installer.bash`:
`shellcheck --shell=bash -x --exclude=SC1091 scripts/local_installer.bash`
`shfmt --language-dialect bash -d -i 4 scripts/local_installer.bash`

---

### Task 2: Update remote_installer.bash to configure Cloudflared

**Files:**
- Modify: `scripts/remote_installer.bash`

**Step 1: Write code changes**

We will add a new function `install_cloudflared` to `scripts/remote_installer.bash` that checks for a credentials JSON file in `/tmp/`, copies it to the app user's home directory, substitutes the `[APP_USER_HOME_DIR]` placeholder in the config, installs systemd files, and enables/restarts the services.

We will add the following code to `scripts/remote_installer.bash`:

```bash
# Install and configure cloudflared tunnel
install_cloudflared() {
    # Check if there is a credentials JSON file in /tmp/
    local cred_file
    cred_file="$(find /tmp -maxdepth 1 -name "*.json" | head -n 1)"
    if [[ -z "${cred_file}" || ! -f "${cred_file}" ]]; then
        log_info "No Cloudflared credentials file found in /tmp. Skipping Cloudflared setup."
        return 0
    fi

    local tunnel_id
    tunnel_id="$(basename "${cred_file}" .json)"
    log_info "Setting up Cloudflared tunnel: ${tunnel_id}"

    # Install cloudflared binary if not present
    if ! command -v cloudflared >/dev/null 2>&1; then
        log_info "Installing cloudflared via pacman..."
        if ! pacman -S --needed --noconfirm cloudflared; then
            log_error "Failed to install cloudflared package."
            exit 1
        fi
    fi

    # Determine app home dir
    local user_dir
    user_dir="$(getent passwd "$APP" | cut -d: -f6)"
    if [[ ! -d "${user_dir}" ]]; then
        log_error "System user '$APP' does not exist. Cannot determine home directory."
        exit 2
    fi

    # Copy credential file to app home dir
    local cloudflared_dir="${user_dir}/.cloudflared"
    mkdir -p "${cloudflared_dir}"
    chmod 700 "${cloudflared_dir}"
    
    local dest_cred_file="${cloudflared_dir}/${tunnel_id}.json"
    cp "${cred_file}" "${dest_cred_file}"
    chmod 600 "${dest_cred_file}"
    chown -R "${APP}:${APP}" "${cloudflared_dir}"

    # Copy config file to /etc/cloudflared/config.yml
    log_info "Installing cloudflared configuration file..."
    mkdir -p "/etc/cloudflared"
    chmod 755 "/etc/cloudflared"
    
    if [[ ! -f "/tmp/config.yml" ]]; then
        log_error "Cloudflared config file not found at /tmp/config.yml"
        exit 24
    fi
    
    # Copy config and substitute [APP_USER_HOME_DIR]
    sed "s|\[APP_USER_HOME_DIR\]|${user_dir}|g" "/tmp/config.yml" > "/etc/cloudflared/config.yml"
    chmod 644 "/etc/cloudflared/config.yml"

    # Copy systemd files
    log_info "Copying cloudflared systemd files..."
    cp "/tmp/cloudflared.service" "/etc/systemd/system/"
    cp "/tmp/cloudflared-update.service" "/etc/systemd/system/"
    cp "/tmp/cloudflared-update.timer" "/etc/systemd/system/"
    chmod 644 /etc/systemd/system/cloudflared.service /etc/systemd/system/cloudflared-update.service /etc/systemd/system/cloudflared-update.timer

    # Enable and start services
    log_info "Activating cloudflared systemd services..."
    if ! systemctl daemon-reload; then
        log_error "Failed to reload systemd daemon"
        exit 24
    fi
    
    if ! systemctl enable cloudflared.service; then
        log_error "Failed to enable cloudflared.service"
        exit 24
    fi
    if ! systemctl restart cloudflared.service; then
        log_error "Failed to start/restart cloudflared.service"
        exit 24
    fi
    
    if ! systemctl enable cloudflared-update.timer; then
        log_error "Failed to enable cloudflared-update.timer"
        exit 24
    fi
    if ! systemctl restart cloudflared-update.timer; then
        log_error "Failed to start/restart cloudflared-update.timer"
        exit 24
    fi

    # Validate active state
    if ! systemctl is-active --quiet cloudflared.service; then
        log_error "cloudflared.service is not active after activation"
        exit 24
    fi
    if ! systemctl is-active --quiet cloudflared-update.timer; then
        log_error "cloudflared-update.timer is not active after activation"
        exit 24
    fi
    
    log_info "cloudflared services activated successfully"

    # Clean up temp files from /tmp
    log_info "Cleaning up temporary cloudflared files..."
    rm -f "/tmp/cloudflared.service" "/tmp/cloudflared-update.service" "/tmp/cloudflared-update.timer" "/tmp/config.yml" "${cred_file}"
}
```

We will update `main` in `scripts/remote_installer.bash`:
```bash
    stop_system_service
    install_app_as_user
    start_system_service
    assert_system_service_running
    install_cloudflared
    log_info "Remote installation completed."
```

And update help text and exit code descriptions:
```bash
#     23  Cloudflared credential validation failure
#     24  Cloudflared service activation failure
```

**Step 2: Run verification**

Run shellcheck and shfmt on `scripts/remote_installer.bash`:
`shellcheck --shell=bash -x --exclude=SC1091 scripts/remote_installer.bash`
`shfmt --language-dialect bash -d -i 4 scripts/remote_installer.bash`

---

### Task 3: Comprehensive Linting and Verification

**Files:**
- Test: `scripts/linting.bash`

**Step 1: Run project bash checks**

Run:
`./scripts/linting.bash --scripts`
Expected: Passes with no shellcheck or shfmt warnings/errors.
