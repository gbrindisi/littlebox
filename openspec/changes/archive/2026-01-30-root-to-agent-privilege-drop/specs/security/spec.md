## MODIFIED Requirements

### Requirement: Privilege separation
The security module SHALL ensure firewall runs as root and agent runs as unprivileged user.

#### Scenario: Container starts as root
- **WHEN** container entrypoint starts
- **THEN** process is running as root (UID 0)

#### Scenario: Firewall runs as root
- **WHEN** container entrypoint starts
- **THEN** firewall setup script runs directly with root privileges (no sudo)

#### Scenario: Agent runs as unprivileged user
- **WHEN** firewall setup completes
- **THEN** entrypoint uses setpriv to transition to agent user (uid 1000 or matched host UID)
- **AND** agent process is executed as non-root user

#### Scenario: No privilege escalation
- **WHEN** agent attempts to use sudo
- **THEN** command fails (sudo is not installed in the image)

### Requirement: Sudo removal
The security module SHALL NOT include sudo in the container image.

#### Scenario: Sudo not installed
- **WHEN** container image is built
- **THEN** sudo package is not installed

#### Scenario: No sudoers files
- **WHEN** container image is built
- **THEN** no sudoers configuration files exist in /etc/sudoers.d/

### Requirement: No new privileges flag
The security module SHALL prevent privilege escalation via setuid binaries.

#### Scenario: no-new-privileges enabled by default
- **WHEN** container is created without explicit no_new_privileges config
- **THEN** security-opt no-new-privileges:true is set

#### Scenario: Setuid binaries ineffective
- **WHEN** agent executes a setuid binary
- **THEN** binary runs with agent's privileges, not elevated privileges

#### Scenario: Container functions with no-new-privileges
- **WHEN** container is created with no_new_privileges: true (default)
- **THEN** firewall initialization completes successfully
- **AND** agent command executes successfully

## REMOVED Requirements

### Requirement: Sudo removal
**Reason**: Replaced by new "Sudo removal" requirement above. The old requirement focused on removing sudo ACCESS after init; the new requirement removes sudo entirely from the image.
**Migration**: No user action needed - sudo is simply not available.
