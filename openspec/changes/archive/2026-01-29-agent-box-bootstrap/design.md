## Context

This is a greenfield Go project. The target users are developers running autonomous coding agents in scripted loops (Ralph) who need isolation without sacrificing agent capabilities.

Anthropic provides a reference devcontainer setup for Claude Code that uses iptables-based network isolation. This design builds on that approach but packages it as a standalone CLI tool with a simpler UX.

Constraints:
- Must work on macOS (Docker Desktop) and Linux
- Must support interactive (TTY) and scripted (piped) usage
- Agent must not be able to bypass firewall rules
- Files created by agent must be usable by host user

## Goals / Non-Goals

**Goals:**
- Drop-in replacement for `claude --dangerously-skip-permissions`
- Single config file (`Agentfile`) captures all sandbox settings
- Secure by default with escape hatches for power users
- Support for multiple agent types via profiles

**Non-Goals:**
- GUI or web interface
- Container orchestration (multiple agents at once)
- Remote execution (always local Docker)
- Windows support (initially)

## Decisions

### 1. Go as implementation language

**Decision**: Use Go with Cobra for CLI and Docker SDK for container management.

**Alternatives considered**:
- Python: Slower startup, requires runtime, dependency management overhead
- Rust: Higher learning curve, slower iteration
- Shell script: Hard to maintain, poor error handling

**Rationale**: Go produces single static binaries, has excellent Docker SDK support, and Cobra is the standard for CLI tools. Fast startup matters for scripted usage.

### 2. Firewall inside container with privilege separation

**Decision**: Run iptables setup as root in container entrypoint, then drop all privileges and exec agent as unprivileged user.

**Alternatives considered**:
- Docker network policies: More complex, less portable
- Host-level firewall: Requires host root access, affects other processes
- No firewall: Defeats the purpose

**Rationale**: Container-internal firewall is self-contained and doesn't require host modifications. Privilege separation prevents agent from disabling rules.

**Implementation**:
```
entrypoint.sh (root)
  1. Apply iptables rules
  2. Remove sudo access (rm /etc/sudoers.d/*)
  3. Make firewall script unreadable (chmod 000)
  4. exec setpriv --reuid=agent --inh-caps=-all -- <agent command>
```

### 3. Agentfile as YAML configuration

**Decision**: Use `Agentfile` (auto-discovered) or `-c/--config` flag. YAML format with schema validation.

**Alternatives considered**:
- TOML: Less familiar to most users
- JSON: No comments, verbose
- CLI flags only: Hard to reproduce, no version control

**Rationale**: YAML is familiar from Docker Compose, Kubernetes, etc. Agentfile name is distinctive and grep-able.

### 4. TTY auto-detection with override flags

**Decision**: Detect if stdin is a terminal. Allocate PTY for interactive use, skip for piped/scripted use. Provide `--tty` and `--no-tty` flags for explicit control.

**Alternatives considered**:
- Always allocate TTY: Breaks piping (stdout/stderr merged, \r added)
- Never allocate TTY: Breaks interactive UI

**Rationale**: Auto-detection handles 90% of cases correctly. Flags provide escape hatches.

### 5. Host UID/GID matching

**Decision**: On Linux, detect host UID/GID and run container with matching IDs. On macOS, use default (Docker Desktop's virtiofs handles this).

**Alternatives considered**:
- Fixed UID 1000: Causes permission issues on Linux hosts with different UIDs
- Podman --userns=keep-id: Would require Podman, reduces portability
- Fix permissions on exit: Slow, touches all files, messes with git

**Rationale**: Matching UIDs is the cleanest solution. macOS doesn't need it due to Docker Desktop's file sharing layer.

### 6. Local image builds with version caching

**Decision**: Build images locally on first run. Cache with version tag (e.g., `agentbox/claude-code:1.0.0`). Provide `agentbox update` to force rebuild.

**Alternatives considered**:
- Pull from registry: Requires publishing, versioning, trust
- Always rebuild: Slow
- No caching: Slow

**Rationale**: Local builds avoid registry dependency. Version tags enable cache invalidation on agentbox updates.

## Risks / Trade-offs

**[Risk] Agent could exploit container escape vulnerabilities**
Mitigation: Use `--security-opt=no-new-privileges`, drop all capabilities after firewall setup, enable seccomp. This is defense in depth, not foolproof.

**[Risk] DNS-based firewall can be bypassed via IP changes**
Mitigation: Resolve domains at container start and add to ipset. Accept that long-running containers may have stale rules. Document this limitation.

**[Risk] File permission issues on Linux**
Mitigation: UID matching by default on Linux. Document the behavior and provide `match_host_user` config option.

**[Risk] Breaking changes in Docker API**
Mitigation: Pin Docker SDK version, test against multiple Docker versions in CI.

**[Trade-off] Firewall in container requires NET_ADMIN capability during init**
Accepted: We drop it immediately after setup. The brief window is acceptable given the isolation benefits.

**[Trade-off] First run is slow (image build)**
Accepted: Subsequent runs use cached image. Provide progress output during build.

## Open Questions

1. Should we support Podman as an alternative runtime? (adds complexity but better security model)
2. How to handle API key injection securely? (env var passthrough vs secrets mount)
3. Should profiles be built into the binary or loaded from `~/.config/agentbox/profiles.yaml`?
