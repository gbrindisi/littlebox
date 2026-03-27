#!/bin/bash
# littlebox container entrypoint
# This script initializes the firewall as root and then drops privileges
# to the agent user before executing the agent command.
#
# Security flow:
# 1. Verify we're running as root (container starts as root)
# 2. Initialize firewall directly (we are root, no sudo needed)
#    - The firewall script makes itself unreadable after initialization
# 3. Resolve agent user UID/GID and home directory dynamically
# 4. Set up environment variables (HOME, USER, LOGNAME) for agent user
# 5. Drop to agent user via setpriv with all capabilities cleared
# 6. Execute agent command

set -euo pipefail

log() {
    echo "[entrypoint] $*"
}

log "Starting littlebox container"

# 1. Verify we're running as root (expected since Dockerfile has USER root)
if [ "$(id -u)" -ne 0 ]; then
    log "ERROR: Entrypoint must run as root for firewall initialization"
    exit 1
fi

# 2. Initialize firewall directly (we are root)
# Note: We always call the script even if ALLOWED_DOMAINS is empty.
# The script handles that case gracefully and still does privilege cleanup.
if [ -n "${ALLOWED_DOMAINS:-}" ]; then
    log "Initializing firewall with allowed domains..."
else
    log "Initializing firewall (no domains configured)..."
fi
/usr/local/bin/init-firewall.sh

# 3. Resolve agent user UID/GID dynamically
# The agent user's UID/GID may be modified at container creation time
AGENT_UID=$(id -u agent)
AGENT_GID=$(id -g agent)
AGENT_HOME=$(getent passwd agent | cut -d: -f6)
log "Resolved agent user: UID=$AGENT_UID GID=$AGENT_GID HOME=$AGENT_HOME"

# 4. Set up environment for the agent user
# setpriv doesn't set these, so we export them for the exec'd process
export HOME="$AGENT_HOME"
export USER=agent
export LOGNAME=agent

# 4.5. Set up LD_PRELOAD for friendly firewall error messages
# The libsandbox.so library intercepts connect() calls and provides user-friendly
# error messages when connections are blocked by the firewall
export LD_PRELOAD=/usr/local/lib/libsandbox.so

log "Dropping privileges to agent user and executing: $*"

# 5. Drop to agent user and execute the agent command
# setpriv options:
#   --reuid: Change real and effective UID to agent user
#   --regid: Change real and effective GID to agent group
#   --init-groups: Initialize supplementary groups for the agent user
#   --inh-caps=-all: Drop all inheritable capabilities
#   --bounding-set=-all: Clear bounding set (prevents future capability acquisition)
# This ensures the agent process runs as unprivileged user with no capabilities
exec setpriv --reuid="$AGENT_UID" --regid="$AGENT_GID" --init-groups --inh-caps=-all --bounding-set=-all -- "$@"
