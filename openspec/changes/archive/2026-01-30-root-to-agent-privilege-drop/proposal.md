## Why

The current entrypoint requires `sudo` to run firewall initialization, which conflicts with the `no_new_privileges` security hardening. Users must explicitly set `no_new_privileges: false` for the container to work, weakening the default security posture and creating a confusing out-of-box experience.

## What Changes

- **Container starts as root instead of agent** - Dockerfile USER directive changes from `agent` to `root`
- **Entrypoint runs firewall directly** - No sudo needed since already running as root
- **Entrypoint drops to agent user via setpriv** - Uses `--reuid`/`--regid` flags to transition from root to agent before executing the agent command
- **Remove sudo from image** - No longer needed, reduces attack surface
- **`no_new_privileges: true` works by default** - Privilege flow is always downward (root to agent), never upward

## Capabilities

### New Capabilities

None - this is a reimplementation of existing privilege separation.

### Modified Capabilities

- `security`: The privilege separation mechanism changes from "escalate via sudo" to "drop via setpriv". The security requirements remain the same (firewall runs as root, agent runs unprivileged, no sudo access after init), but the implementation approach changes fundamentally.

## Impact

- `internal/container/docker/Dockerfile` - Change USER directive, remove sudo package and sudoers setup
- `internal/container/docker/entrypoint.sh` - Remove sudo call, add setpriv user/group drop
- `internal/container/docker/init-firewall.sh` - Remove sudoers cleanup (no longer needed)
- `internal/config/defaults.go` - `no_new_privileges` default of `true` now works correctly
- Tests - Tests that set `no_new_privileges: false` can be simplified
- User Agentfiles - Users no longer need `no_new_privileges: false` workaround
