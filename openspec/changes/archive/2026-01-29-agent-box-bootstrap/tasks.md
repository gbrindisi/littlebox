## 1. Project Setup

- [ ] 1.1 Initialize Go module with `go mod init github.com/gbrindisi/agentbox`
- [ ] 1.2 Add dependencies: cobra, docker SDK, yaml.v3
- [ ] 1.3 Create directory structure: cmd/, internal/config/, internal/container/, internal/network/, internal/security/, pkg/
- [ ] 1.4 Create main.go with Cobra root command

## 2. Configuration System

- [ ] 2.1 Define Agentfile YAML schema as Go structs (Config, Agent, Workspace, Mount, Network, Security, Container)
- [ ] 2.2 Implement config loader with YAML parsing and validation
- [ ] 2.3 Implement tilde expansion for mount paths
- [ ] 2.4 Implement built-in profiles (claude-code, aider) with defaults
- [ ] 2.5 Implement network presets (strict, standard, permissive)
- [ ] 2.6 Implement environment variable passthrough logic
- [ ] 2.7 Add config validation (required fields, valid paths, valid presets)

## 3. Docker Image

- [ ] 3.1 Create Dockerfile with base image (node:20 or debian)
- [ ] 3.2 Install required packages: iptables, ipset, iproute2, dnsutils, aggregate, jq
- [ ] 3.3 Create agent user (uid 1000) with home directory
- [ ] 3.4 Copy entrypoint.sh and init-firewall.sh scripts
- [ ] 3.5 Set up sudoers for firewall init (to be removed at runtime)
- [ ] 3.6 Embed Dockerfile in Go binary using embed package

## 4. Firewall Scripts

- [ ] 4.1 Create init-firewall.sh based on Anthropic reference
- [ ] 4.2 Implement Docker DNS preservation logic
- [ ] 4.3 Implement domain resolution and ipset creation
- [ ] 4.4 Implement GitHub IP range fetching
- [ ] 4.5 Implement firewall verification (test blocked/allowed)
- [ ] 4.6 Create entrypoint.sh with privilege separation (firewall setup, sudo removal, capability drop, exec agent)

## 5. Container Manager

- [ ] 5.1 Implement image build logic with version tagging
- [ ] 5.2 Implement image cache check (skip build if cached)
- [ ] 5.3 Implement container creation with mounts and environment
- [ ] 5.4 Implement TTY detection and allocation logic
- [ ] 5.5 Implement stdin/stdout/stderr attachment
- [ ] 5.6 Implement signal forwarding (SIGINT, SIGTERM)
- [ ] 5.7 Implement window resize handling (SIGWINCH)
- [ ] 5.8 Implement container wait and exit code capture
- [ ] 5.9 Implement container cleanup (remove on exit)
- [ ] 5.10 Implement host UID/GID detection and matching (Linux vs macOS)

## 6. Security Hardening

- [ ] 6.1 Implement no-new-privileges security option
- [ ] 6.2 Implement capability dropping in entrypoint (setpriv --inh-caps=-all)
- [ ] 6.3 Implement sudo removal in entrypoint
- [ ] 6.4 Implement firewall script protection (chmod 000)
- [ ] 6.5 Add seccomp profile configuration
- [ ] 6.6 Add read-only root filesystem option

## 7. CLI Commands

- [ ] 7.1 Implement `run` command with workspace and config flags
- [ ] 7.2 Implement `--` argument passthrough to agent
- [ ] 7.3 Implement `--tty` and `--no-tty` flags
- [ ] 7.4 Implement `init` command with profile flag and template generation
- [ ] 7.5 Implement `validate` command with error reporting
- [ ] 7.6 Implement `shell` command for debug access
- [ ] 7.7 Implement Agentfile auto-discovery logic

## 8. File Redaction

- [ ] 8.1 Implement glob pattern matching for redact config
- [ ] 8.2 Implement mount exclusion for redacted files
- [ ] 8.3 Test redaction with .env and *.pem patterns

## 9. Testing

- [ ] 9.1 Write unit tests for config parsing and validation
- [ ] 9.2 Write unit tests for profile and preset resolution
- [ ] 9.3 Write integration test for container lifecycle
- [ ] 9.4 Write integration test for firewall blocking
- [ ] 9.5 Write integration test for signal handling
- [ ] 9.6 Test on macOS with Docker Desktop
- [ ] 9.7 Test on Linux with native Docker

## 10. Documentation and Release

- [ ] 10.1 Write README with installation and usage instructions
- [ ] 10.2 Document Agentfile schema with examples
- [ ] 10.3 Create example Agentfiles for common use cases
- [ ] 10.4 Set up goreleaser for cross-platform builds
- [ ] 10.5 Create initial release (v0.1.0)
