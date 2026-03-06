## Why

Running autonomous coding agents (Claude Code, Aider, etc.) with `--dangerously-skip-permissions` is necessary for scripted Ralph loops but exposes the host system to accidental damage. We need a sandboxed execution environment that lets agents run with full permissions inside a container while limiting their ability to harm the host filesystem or network.

## What Changes

- New Go CLI tool `agentbox` that wraps Docker to run agents in isolated containers
- `Agentfile` configuration format for declarative sandbox setup
- Built-in network firewall using iptables with configurable allow/deny lists
- Privilege separation: firewall runs as root, agent runs as unprivileged user
- Support for multiple agent profiles (claude-code, aider, custom)
- Automatic TTY detection for interactive vs scripted usage
- Host UID/GID matching to prevent file permission issues

## Capabilities

### New Capabilities

- `cli`: Go CLI commands (run, init, validate, shell) using Cobra framework
- `config`: Agentfile parsing and validation (YAML schema, profiles, presets)
- `container`: Docker container lifecycle management (create, start, attach, cleanup)
- `network`: Firewall rules and network isolation (iptables, ipset, allow/deny lists)
- `security`: Privilege separation and hardening (capability dropping, no-new-privileges)

### Modified Capabilities

None - this is a new project.

## Impact

- **New files**: Go module structure, Dockerfile, entrypoint scripts, firewall scripts
- **Dependencies**: Go standard library, Cobra (CLI), Docker SDK, YAML parser
- **Runtime requirements**: Docker or Podman on host system
- **User workflow**: Replace `claude --dangerously-skip-permissions` with `agentbox run`
