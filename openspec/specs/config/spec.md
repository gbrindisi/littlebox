## ADDED Requirements

### Requirement: Agentfile YAML schema
The config system SHALL parse Agentfile as YAML with a defined schema.

#### Scenario: Parse valid Agentfile
- **WHEN** config loader reads a valid Agentfile
- **THEN** it returns a parsed Config struct with all fields populated

#### Scenario: Reject invalid YAML
- **WHEN** config loader reads malformed YAML
- **THEN** it returns an error with line number and description

### Requirement: Workspace configuration
The config system SHALL support workspace path, mount point, and write mode configuration.

#### Scenario: Default workspace to Agentfile directory
- **WHEN** Agentfile omits `workspace.path`
- **THEN** config defaults to the directory containing the Agentfile (`.`)

#### Scenario: Relative workspace path
- **WHEN** Agentfile specifies `workspace.path: ../other-project`
- **THEN** config resolves it relative to Agentfile location

#### Scenario: Tilde workspace path
- **WHEN** Agentfile specifies `workspace.path: ~/projects/myapp`
- **THEN** config expands `~` to user home directory

#### Scenario: Absolute workspace path unchanged
- **WHEN** Agentfile specifies `workspace.path: /absolute/path`
- **THEN** config uses the path as-is

#### Scenario: Read-only workspace
- **WHEN** Agentfile specifies `workspace.writable: false`
- **THEN** config marks workspace as read-only mount

### Requirement: Additional mounts configuration
The config system SHALL support mounting additional host paths into the container.

#### Scenario: Mount with relative source path
- **WHEN** Agentfile includes mount with `source: ./local-config`
- **THEN** config resolves source relative to Agentfile location

#### Scenario: Mount with tilde source path
- **WHEN** Agentfile includes mount with `source: ~/.claude`
- **THEN** config expands `~` to user home directory

#### Scenario: Mount with absolute source path
- **WHEN** Agentfile includes mount with `source: /absolute/path`
- **THEN** config uses the source path as-is

#### Scenario: Expand tilde in paths
- **WHEN** mount source contains `~`
- **THEN** config expands it to user's home directory

### Requirement: Path resolution relative to Agentfile
The config system SHALL resolve relative paths in workspace and mounts relative to the Agentfile location.

#### Scenario: Resolve workspace relative to Agentfile
- **WHEN** Agentfile at /projects/myapp/Agentfile specifies `workspace.path: .`
- **THEN** config resolves to /projects/myapp

#### Scenario: Resolve parent directory reference
- **WHEN** Agentfile at /projects/myapp/Agentfile specifies `workspace.path: ../shared`
- **THEN** config resolves to /projects/shared

#### Scenario: Resolve mount source relative to Agentfile
- **WHEN** Agentfile at /projects/myapp/Agentfile has mount with `source: ./config`
- **THEN** config resolves source to /projects/myapp/config

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
The config system SHALL support passing host environment variables to container via explicit configuration.

#### Scenario: Explicit env passthrough
- **WHEN** Agentfile specifies `agent.env_passthrough: [ANTHROPIC_API_KEY]`
- **THEN** ANTHROPIC_API_KEY from host is passed to container

#### Scenario: Glob pattern env passthrough
- **WHEN** Agentfile specifies `agent.env_passthrough: [CLAUDE_CODE_*]`
- **THEN** all environment variables matching the pattern are passed to container

#### Scenario: Missing optional env var
- **WHEN** passthrough variable is not set on host
- **THEN** variable is silently omitted from container environment
