# Littlebox

A sandbox for running autonomous coding agents with network and filesystem isolation.

Coding agents are most useful when they can run autonomously, but this introduces risks: rogue activities, data exfiltration, or unintended external communications. Littlebox provides a containerized sandbox around any agent, with sensible defaults for quick setup and enough flexibility to customize for your needs.

(based on rationale outlined [here](https://cloudberry.engineering/article/on-sandboxing-agents/))

## The Security Model

Running agents autonomously means giving them significant control over your system. An agent with unrestricted network access could exfiltrate code to external servers. An agent with full filesystem access could read sensitive credentials or modify critical files.

Littlebox implements defense in depth to mitigate these risks:

| Layer | Protection | What It Prevents |
|-------|------------|------------------|
| **Network firewall** | iptables default-deny policy | Data exfiltration, unauthorized API calls, C2 communication |
| **Privilege separation** | Root configures firewall, then drops to unprivileged user | Agent cannot modify firewall rules or access privileged operations |
| **Capability dropping** | All Linux capabilities removed via `setpriv` | Prevents kernel-level exploits and privilege escalation |
| **No new privileges** | Blocks setuid/setgid binaries | Prevents exploitation of SUID binaries for privilege escalation |
| **Seccomp filtering** | Restricts available syscalls | Limits attack surface for container escapes |
| **No sudo** | sudo not installed in container | Eliminates common privilege escalation vector |

This is a best-effort, defense-in-depth approach. Container escape vulnerabilities could bypass all protections. The goal is to significantly raise the bar for malicious behavior while maintaining agent usability.

## Quick Start

**Requirements:**
- Docker (Docker Desktop on macOS, native Docker on Linux)
- Go 1.21+ (for building from source)

**Install:**
```bash
git clone https://github.com/gbrindisi/littlebox.git
cd littlebox
make build
sudo make install
```

**Run:**
```bash
littlebox init --profile claude-code
export ANTHROPIC_API_KEY=sk-...
littlebox run -- "2+2=?"
```

Note: On macOS, use this to inherit an existing Claude Code session:
```bash
ANTHROPIC_API_KEY=$(security find-generic-password -w -s "Claude Code") littlebox run -- "2+2=?"
```

## Configuration

The sandbox is defined in an `Agentfile`, a YAML configuration file auto-discovered in the current directory.

```yaml
# littlebox configuration
# See: https://github.com/gbrindisi/littlebox

# Agent configuration
agent:
  # Build script runs as root during image build to install tools
  build_script: |
    curl -fsSL https://claude.ai/install.sh | bash
    # Copy claude to system-wide location (install script puts it in ~/.local/bin)
    cp /root/.local/bin/claude /usr/local/bin/claude

  # Command and arguments to run
  command: ["claude"]
  args: ["--dangerously-skip-permissions", "--print"]

  # Environment variables to pass from host to container
  # Supports glob patterns like CLAUDE_CODE_*
  # Exact names are optional (skipped if unset). To require one:
  #   - name: MY_VAR
  #     required: true
  env_passthrough:
   - ANTHROPIC_API_KEY
  #  - CLAUDE_CODE_*

# Workspace configuration
workspace:
  # Directory to mount as /workspace in the container
  # Default: current directory
  path: /path/to/code

# Additional mounts (optional)
mounts:
  - source: $HOME/.claude
    target: /home/agent/.claude
    readonly: false

# Network configuration
network:
  # Composable presets for services your agent needs
  presets:
    - anthropic
    - github
    - npm
    - pypi

  # Additional custom domains or ips (optional):
  # allow:
  #   - custom-api.example.com

# Container security settings (optional)
# container:
#   no_new_privileges: true    # Prevent privilege escalation (default: true)
#   readonly_root: false       # Make root filesystem read-only
#   seccomp_profile: default   # Seccomp profile: default, unconfined, or path

```

### Build Script

Install tools and dependencies at image build time. Runs as root during `docker build`, result is cached based on script content hash.

```yaml
agent:
  build_script: |
    apt-get update && apt-get install -y nodejs npm
    npm install -g my-tool
```

By default the derived image is also keyed on the workspace path, so each new workspace triggers its own build. Set `agent.image_scope: shared` to reuse a single image for every workspace with the same build script:

```yaml
agent:
  image_scope: shared   # workspace (default) | shared
```

Use `--rebuild` to force a rebuild (Docker layer cache still applies). Runtime execution always runs as unprivileged `agent` user.

### Environment Variables

Pass environment variables from host to container by exact name or glob pattern.
Exact names are optional by default: if unset on the host they are silently
skipped. Mark a variable as required with the mapping form; validation then
fails if it is unset or empty. Glob patterns cannot be required.

```yaml
agent:
  env_passthrough:
    - ANTHROPIC_API_KEY    # Exact match, optional
    - OPENAI_API_KEY       # Optional
    - name: GITHUB_TOKEN   # Exact match, required
      required: true
    - CLAUDE_CODE_*        # Glob pattern
    - AWS_*                # All AWS variables
```

Set fixed values with `env` (top level) or `agent.env`. Fixed values override
passthrough values of the same name, and `agent.env` overrides top-level `env`.
Names must match `[A-Za-z_][A-Za-z0-9_]*`.

```yaml
env:
  - name: LOG_LEVEL
    value: info

agent:
  env:
    - name: DISABLE_TELEMETRY
      value: "1"
```

### Mounts

Mount host directories into the container:

```yaml
mounts:
  - source: ~/.ssh
    target: /home/agent/.ssh
    readonly: true
  - source: ~/.gitconfig
    target: /home/agent/.gitconfig
    readonly: true
```

### Firewalling

The firewall uses iptables with a default-deny OUTPUT policy. At container startup, allowed domains are resolved to IP addresses using `dig` and stored in an ipset for efficient matching. DNS queries are restricted to the container's DNS server only. IPv6 is blocked by default to prevent bypass.

A companion `libsandbox.so` library is injected via `LD_PRELOAD` to intercept `connect()` calls. When a connection to a blocked IP is attempted, instead of a silent timeout, the agent receives an informative error message:

```
littlebox: Connection to example.com (203.0.113.50) blocked by sandbox firewall. This is not bypassable.
```

The library also intercepts `getaddrinfo()` and `gethostbyname()`, caching IPv4 lookups for 2 seconds, so the error can name the domain. Without a recent lookup, only the IP is shown:

```
littlebox: Connection to 203.0.113.50 blocked by sandbox firewall. This is not bypassable.
```

This helps agents understand why a connection failed and adjust their behavior accordingly, rather than retrying indefinitely or misdiagnosing the issue.

Control network access with composable presets or custom domain lists:

```yaml
# Combine presets for the services your agent needs
network:
  presets:
    - anthropic
    - github
    - npm

# Add custom domains alongside presets
network:
  presets:
    - anthropic
  allow:
    - custom-api.example.com
```

**Available presets:**

| Category | Preset | Domains |
|----------|--------|---------|
| AI Providers | `anthropic` | api.anthropic.com, anthropic.com, claude.ai |
| | `openai` | api.openai.com, openai.com, platform.openai.com, cdn.openai.com |
| | `google-ai` | generativelanguage.googleapis.com, ai.google.dev, aistudio.google.com |
| | `mistral` | api.mistral.ai, mistral.ai |
| Source Control | `github` | github.com, api.github.com, raw.githubusercontent.com, objects.githubusercontent.com, codeload.github.com |
| | `gitlab` | gitlab.com, registry.gitlab.com |
| | `bitbucket` | bitbucket.org, api.bitbucket.org |
| Package Registries | `npm` | registry.npmjs.org, npmjs.org, npmjs.com |
| | `pypi` | pypi.org, files.pythonhosted.org |
| | `cargo` | crates.io, static.crates.io, index.crates.io |
| | `rubygems` | rubygems.org |
| ML/AI Models | `huggingface` | huggingface.co, cdn-lfs.huggingface.co |

### Container Settings

Fine-tune container security:

```yaml
container:
  no_new_privileges: true    # Prevent privilege escalation (default: true)
  readonly_root: false       # Make root filesystem read-only
  seccomp_profile: default   # default, unconfined, or path to custom profile
```

## CLI Commands

### littlebox init

Create a new Agentfile from a profile template.

```bash
littlebox init                         # Uses claude-code profile
littlebox init --profile claude-code   # Explicit profile
```

### littlebox run

Run an agent in the sandbox.

```bash
littlebox run                          # Run with auto-discovered Agentfile
littlebox run -c /path/to/Agentfile    # Specific config file
littlebox run -w /path/to/project      # Override workspace directory
littlebox run --rebuild                # Force image rebuild
littlebox run --tty                    # Force TTY allocation
littlebox run --no-tty                 # Disable TTY allocation
littlebox run --debug                  # Enable debug output
littlebox run --name my-agent          # Set container name
littlebox run --label k=v --label a=b  # Add container labels (repeatable)
littlebox run -- "your prompt here"    # Pass arguments to agent
```

Every container gets the label `littlebox=1` (find them with `docker ps --filter label=littlebox=1`). Label keys must match `[A-Za-z0-9._/-]` (alphanumeric at both ends); the `littlebox` key is reserved.

### littlebox validate

Validate an Agentfile without running.

```bash
littlebox validate
littlebox validate -c /path/to/Agentfile
```

### littlebox shell

Open a debug shell in the container. Uses the same configuration as `run` but starts bash instead of the agent. Useful for testing firewall rules and inspecting the environment.

```bash
littlebox shell                        # Open shell with auto-discovered Agentfile
littlebox shell --rebuild              # Force image rebuild before opening shell
littlebox shell --debug                # Enable debug output
```

## Good to Know

### Limitations

- **DNS at startup**: Domain names are resolved to IPs when the container starts. Long-running containers won't pick up DNS changes.
- **Container escapes**: A container escape vulnerability would bypass all protections. This is defense in depth, not a security guarantee.
- **Mount access**: The agent has full access to all mounted directories within the container.

### Signals and Cleanup

`SIGINT`, `SIGTERM` and `SIGQUIT` are forwarded to the container. On `SIGHUP` (e.g. the terminal or tmux window is closed), littlebox stops the container (SIGTERM, then SIGKILL after 10 seconds) and removes it, so no orphaned containers are left behind.

### File Ownership Issues

Files created by the agent are owned by UID 1000 (the `agent` user). On Linux, this may differ from your host UID.

**Solutions:**
1. Run `chown` after the container exits
2. Match UIDs in your build script:

```yaml
agent:
  build_script: |
    usermod -u 501 agent      # Replace 501 with your UID
    groupmod -g 501 agent
    chown -R agent:agent /home/agent
```

Note: The container must start as root to configure the firewall before dropping privileges, so runtime UID cannot be overridden.
