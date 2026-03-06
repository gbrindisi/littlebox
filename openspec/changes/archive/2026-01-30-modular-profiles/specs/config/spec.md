## MODIFIED Requirements

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

## ADDED Requirements

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
