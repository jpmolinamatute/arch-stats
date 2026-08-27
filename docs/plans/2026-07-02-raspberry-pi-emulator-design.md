# Raspberry Pi 5 Emulator Service Design

**Date**: 2026-07-02  
**Feature**: Emulation of a remote Raspberry Pi 5 target machine for deploy/install script testing.

## Overview

To test remote deployment and installation scripts (`deploy.bash`, `remote_installer.bash`, `install_app.bash`) locally, we need a target environment that:
1. Runs systemd (needed for service unit management and postgresql service status checks).
2. Runs an SSH server.
3. Automatically authorizes a dedicated development SSH key (`docker/rpi/ssh/arch_stats_dev.pub`) for passwordless connectivity.
4. Pre-installs PostgreSQL and required utility binaries (`sudo`, `curl`, `unzip`, `jq`, `tar`, `xz-utils`).

We will implement this using a custom Docker service called `emulator` running `jrei/systemd-debian:12`, configured under the `emulator` Compose profile so it does not start by default.

## Design

### 1. Dockerfile (`docker/Dockerfile.rpi`)
The container image will extend `jrei/systemd-debian:12`.

- Installs `openssh-server`, `sudo`, `curl`, `unzip`, `jq`, `postgresql`, `postgresql-contrib`, `ca-certificates`, `tar`, `xz-utils`.
- Enables root login via public key by updating `/etc/ssh/sshd_config`.
- Adds a helper script `/usr/local/bin/rpi-init.sh` and registers it as a systemd unit (`rpi-init.service`) to run on boot.

### 2. Initialization Service (`docker/rpi/rpi-init.service`)
A oneshot systemd unit running before SSH daemon:

```ini
[Unit]
Description=Initialize Raspberry Pi Emulator SSH Keys
Before=ssh.service

[Service]
Type=oneshot
ExecStart=/usr/local/bin/rpi-init.sh
RemainAfterExit=yes

[Install]
WantedBy=multi-user.target
```

### 3. Initialization Script (`docker/rpi/rpi-init.sh`)
This script:
- Creates `/root/.ssh` if it doesn't exist.
- Appends the dedicated public key from the mounted directory `/host-ssh/arch_stats_dev.pub` to `/root/.ssh/authorized_keys`.
- Sets secure permissions (`700` for `.ssh`, `600` for `authorized_keys`).

### 4. Docker Compose Integration (`docker/docker-compose.yaml`)
Add a new service `emulator` to `docker-compose.yaml`:

- Build context pointing to the repository root.
- Mounts `/sys/fs/cgroup` to support systemd.
- Mounts the project's `./docker/rpi/ssh` directory to `/host-ssh` as read-only.
- Runs in `privileged: true` mode.
- Maps port `2222` to the container's port `22`.
- Placed on the `arch-stats` network.
- Configured with `profiles: [emulator]` so it is isolated from the default services.

### 5. Git Configuration (`.gitignore`)
The private key `docker/rpi/ssh/arch_stats_dev` will be added to `.gitignore` to prevent committing development keys to the repository.
