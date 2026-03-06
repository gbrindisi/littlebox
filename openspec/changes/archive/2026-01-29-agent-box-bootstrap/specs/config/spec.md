## ADDED Requirements

### Requirement: Agentfile YAML schema
The config system SHALL parse Agentfile as YAML with a defined schema.

#### Scenario: Parse valid Agentfile
- **WHEN** config loader reads a valid Agentfile
- **THEN** it returns a parsed Config struct with all fields populated

#### Scenario: Reject invalid YAML
- **WHEN** config loader reads malformed YAML
- **THEN** it returns an error with line number and description

### Requirement: Agent profile configuration
The config system SHALL support agent profiles that define image, command, and environment.

#### Scenario: Use built-in profile
- **WHEN** Agentfile specifies `agent.profile: claude-code`
- **THEN** config loader applies claude-code defaults (image, command, env passthrough)

#### Scenario: Custom agent configuration
- **WHEN** Agentfile specifies `agent.image`, `agent.command`, and `agent.env`
- **THEN** config loader uses these values instead of a profile

#### Scenario: Profile with overrides
- **WHEN** Agentfile specifies both `agent.profile` and `agent.env`
- **THEN** config loader merges profile defaults with explicit overrides

### Requirement: Workspace configuration
The config system SHALL support workspace path, mount point, and write mode configuration.

#### Scenario: Default workspace
- **WHEN** Agentfile omits `workspace` section
- **THEN** config defaults to current directory mounted at /workspace with write access

#### Scenario: Explicit workspace path
- **WHEN** Agentfile specifies `workspace.path: /other/dir`
- **THEN** config uses specified path as workspace root

#### Scenario: Read-only workspace
- **WHEN** Agentfile specifies `workspace.writable: false`
- **THEN** config marks workspace as read-only mount

### Requirement: Additional mounts configuration
The config system SHALL support mounting additional host paths into the container.

#### Scenario: Mount gitconfig read-only
- **WHEN** Agentfile includes mount `~/.gitconfig:/home/agent/.gitconfig:ro`
- **THEN** config includes this mount with read-only flag

#### Scenario: Expand tilde in paths
- **WHEN** mount source contains `~`
- **THEN** config expands it to user's home directory

### Requirement: Redact sensitive files
The config system SHALL support hiding files from the agent within the workspace.

#### Scenario: Redact .env files
- **WHEN** Agentfile specifies `redact: [".env", ".env.*"]`
- **THEN** matching files are not visible to the agent in the container

#### Scenario: Redact with glob patterns
- **WHEN** Agentfile specifies `redact: ["**/*.pem"]`
- **THEN** all .pem files in any subdirectory are hidden from agent

### Requirement: Network preset configuration
The config system SHALL support network presets for common firewall configurations.

#### Scenario: Strict preset
- **WHEN** Agentfile specifies `network.preset: strict`
- **THEN** config only allows agent API endpoint (e.g., api.anthropic.com)

#### Scenario: Standard preset
- **WHEN** Agentfile specifies `network.preset: standard`
- **THEN** config allows agent API, github.com, and package registries

#### Scenario: Permissive preset
- **WHEN** Agentfile specifies `network.preset: permissive`
- **THEN** config allows all outbound network traffic

### Requirement: Custom network allow/deny lists
The config system SHALL support custom domain allow and deny lists.

#### Scenario: Custom allow list
- **WHEN** Agentfile specifies `network.allow: [api.anthropic.com, github.com]`
- **THEN** config uses this list instead of preset defaults

#### Scenario: Deny all else
- **WHEN** Agentfile specifies `network.deny_all_else: true`
- **THEN** config blocks all domains not in allow list

### Requirement: Container lifecycle configuration
The config system SHALL support container keep/ephemeral behavior.

#### Scenario: Ephemeral by default
- **WHEN** Agentfile omits `container.keep`
- **THEN** container is removed after agent exits

#### Scenario: Keep container
- **WHEN** Agentfile specifies `container.keep: true`
- **THEN** container is preserved after agent exits

### Requirement: Environment variable passthrough
The config system SHALL support passing host environment variables to container.

#### Scenario: Profile env passthrough
- **WHEN** profile defines `env_passthrough: [ANTHROPIC_API_KEY]`
- **THEN** ANTHROPIC_API_KEY from host is passed to container

#### Scenario: Missing required env var
- **WHEN** passthrough variable is not set on host
- **THEN** config validation fails with clear error message
