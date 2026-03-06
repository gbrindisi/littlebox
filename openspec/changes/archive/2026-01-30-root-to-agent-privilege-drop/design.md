## Context

The agentbox container uses a two-phase security model:
1. **Init phase**: Firewall rules (iptables) are configured - requires root/CAP_NET_ADMIN
2. **Agent phase**: Agent command runs with minimal privileges - no capabilities, no sudo

Currently, the container starts as the `agent` user (UID 1000) and uses sudo to escalate to root for firewall setup. This requires `no_new_privileges: false`, which weakens security by allowing setuid binaries to escalate privileges.

The Linux security option `no_new_privileges` prevents privilege escalation (going up) but does NOT prevent privilege dropping (going down). This insight enables a cleaner design.

## Goals / Non-Goals

**Goals:**
- Enable `no_new_privileges: true` by default (stronger security)
- Eliminate sudo from the container image (reduced attack surface)
- Maintain same security guarantees: firewall runs as root, agent runs unprivileged
- Simplify user experience: default config works without workarounds

**Non-Goals:**
- Changing the firewall implementation
- Modifying network filtering behavior
- Changing the agent user UID/GID handling
- Altering capability dropping behavior (already uses setpriv)

## Decisions

### Decision 1: Container starts as root

**Choice**: Change Dockerfile `USER root` instead of `USER agent`

**Rationale**: Starting as root allows the entrypoint to run firewall setup directly without sudo. The privilege drop to agent happens after firewall init.

**Alternatives considered**:
- Keep sudo approach with better documentation - Rejected: still requires weaker security default
- Use capabilities (CAP_NET_ADMIN) without root - Rejected: still needs privilege escalation mechanism

### Decision 2: Use setpriv for user transition

**Choice**: Use `setpriv --reuid=<uid> --regid=<gid> --init-groups` to drop from root to agent

**Rationale**: setpriv is already installed (util-linux) and used for capability dropping. It can handle both user transition and capability dropping in a single exec call.

**Alternatives considered**:
- Use `su agent -c "command"` - Rejected: spawns intermediate shell, less clean
- Use `gosu` - Rejected: requires additional package installation
- Use `exec su-exec` - Rejected: not commonly available

### Decision 3: Dynamic UID/GID resolution

**Choice**: Read agent user's UID/GID at runtime using `id -u agent` and `id -g agent`

**Rationale**: The agent user's UID/GID may be modified at container creation time via `--user` flag or build args. Runtime resolution ensures correct values.

**Alternatives considered**:
- Hardcode UID 1000 - Rejected: breaks match_host_user feature
- Pass UID/GID via environment - Rejected: unnecessary complexity

### Decision 4: Remove sudo entirely

**Choice**: Remove sudo package from Dockerfile, delete sudoers file setup

**Rationale**: Sudo is no longer needed. Removing it reduces attack surface and image size.

## Risks / Trade-offs

**Risk**: Container runs initial commands as root
- **Mitigation**: Root access is limited to entrypoint execution before agent command. The agent command itself runs as unprivileged user with all capabilities dropped.

**Risk**: setpriv --reuid/--regid may behave differently across systems
- **Mitigation**: setpriv is from util-linux, widely available and stable. Test on target platforms.

**Trade-off**: Dockerfile USER is now root
- This means `docker exec` without `-u` flag runs as root
- **Mitigation**: This is acceptable - the security boundary is the agent process, not ad-hoc exec commands
