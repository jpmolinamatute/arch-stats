# Raspberry Pi 5 Emulator Service Implementation Plan

> **For Antigravity:** REQUIRED WORKFLOW: Use `.agent/workflows/execute-plan.md` to execute this plan in single-flow mode.

**Goal:** Add a new Docker service `emulator` running Debian 12 with systemd and SSH authorized with a dedicated project-local key to emulate a remote Raspberry Pi 5.

**Architecture:** Extend `jrei/systemd-debian:12`, install openssh/postgresql/utilities, map SSH port to `2222`, mount the project-local `./docker/rpi/ssh` directory, and copy the public key to `/root/.ssh/authorized_keys` via a startup systemd script. Isolate via `emulator` compose profile.

**Tech Stack:** Docker, Docker Compose, systemd, bash.

---

### Task 1: Create Dockerfile and Initialization Files

**Files:**
- Create: `docker/Dockerfile.rpi`
- Create: `docker/rpi/rpi-init.sh`
- Create: `docker/rpi/rpi-init.service`

**Step 1: Write `docker/rpi/rpi-init.sh`**
Create `/usr/local/bin/rpi-init.sh` inside the container via host file:
```bash
#!/bin/bash
mkdir -p /root/.ssh
chmod 700 /root/.ssh
touch /root/.ssh/authorized_keys
chmod 600 /root/.ssh/authorized_keys

if [ -f /host-ssh/arch_stats_dev.pub ]; then
    echo "Authorizing development SSH key..."
    cat /host-ssh/arch_stats_dev.pub >> /root/.ssh/authorized_keys
    # Deduplicate
    sort -u /root/.ssh/authorized_keys -o /root/.ssh/authorized_keys
else
    echo "WARNING: arch_stats_dev.pub not found in /host-ssh. Root SSH login may fail."
fi
```

**Step 2: Write `docker/rpi/rpi-init.service`**
Create systemd service unit:
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

**Step 3: Write `docker/Dockerfile.rpi`**
Create Dockerfile:
```dockerfile
FROM jrei/systemd-debian:12

# Install dependencies
RUN apt-get update && apt-get install -y \
    openssh-server \
    sudo \
    curl \
    unzip \
    jq \
    postgresql \
    postgresql-contrib \
    ca-certificates \
    tar \
    xz-utils \
    && rm -rf /var/lib/apt/lists/*

# Configure SSH daemon
RUN mkdir -p /var/run/sshd && \
    echo "PermitRootLogin yes" >> /etc/ssh/sshd_config && \
    echo "PubkeyAuthentication yes" >> /etc/ssh/sshd_config

# Copy initialization scripts/services
COPY docker/rpi/rpi-init.sh /usr/local/bin/rpi-init.sh
COPY docker/rpi/rpi-init.service /etc/systemd/system/rpi-init.service

RUN chmod +x /usr/local/bin/rpi-init.sh && \
    systemctl enable rpi-init.service
```

---

### Task 2: Update docker-compose.yaml and .gitignore

**Files:**
- Modify: `docker/docker-compose.yaml`
- Modify: `.gitignore`

**Step 1: Add the emulator service to docker-compose.yaml**
Add the `emulator` service with the `emulator` profile and key directory volume mounting:
```yaml
  emulator:
    build:
      context: ..
      dockerfile: docker/Dockerfile.rpi
    privileged: true
    profiles:
      - emulator
    ports:
      - "2222:22"
    volumes:
      - /sys/fs/cgroup:/sys/fs/cgroup:ro
      - ./rpi/ssh:/host-ssh:ro
    networks:
      - arch-stats
```

**Step 2: Update .gitignore**
Append the development private key to `.gitignore`:
```gitignore
# Raspberry Pi emulator development keys
docker/rpi/ssh/arch_stats_dev
```

---

### Task 3: Verification

**Files:**
- Verify connectivity and setup

**Step 1: Instruct user to test**
Instruct the user to:
1. Create the project key directory and generate key pair:
   ```bash
   mkdir -p docker/rpi/ssh
   ssh-keygen -t rsa -b 2048 -f docker/rpi/ssh/arch_stats_dev -N ""
   ```
2. Build and start services (with emulator profile):
   ```bash
   docker compose -f docker/docker-compose.yaml --profile emulator up --build -d
   ```
3. Verify SSH connection:
   ```bash
   ssh -i docker/rpi/ssh/arch_stats_dev -p 2222 root@localhost
   ```
4. Verify systemctl and postgresql are active inside the container:
   ```bash
   systemctl status postgresql
   ```
