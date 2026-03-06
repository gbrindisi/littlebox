## ADDED Requirements

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

### Requirement: Capability dropping
The security module SHALL drop all Linux capabilities after firewall setup.

#### Scenario: Drop all capabilities
- **WHEN** entrypoint transitions to agent user
- **THEN** all inheritable capabilities are dropped (--inh-caps=-all)

#### Scenario: NET_ADMIN only during init
- **WHEN** container is created
- **THEN** NET_ADMIN and NET_RAW capabilities are available only to init process

### Requirement: Sudo removal
The security module SHALL NOT include sudo in the container image.

#### Scenario: Sudo not installed
- **WHEN** container image is built
- **THEN** sudo package is not installed

#### Scenario: No sudoers files
- **WHEN** container image is built
- **THEN** no sudoers configuration files exist in /etc/sudoers.d/

### Requirement: Firewall script protection
The security module SHALL prevent agent from reading or modifying firewall scripts.

#### Scenario: Firewall script unreadable
- **WHEN** firewall setup completes
- **THEN** init-firewall.sh permissions are set to 000

#### Scenario: Cannot modify firewall rules
- **WHEN** agent attempts to run iptables commands
- **THEN** commands fail due to lack of NET_ADMIN capability

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

### Requirement: Seccomp profile
The security module SHALL apply a seccomp profile to limit syscalls.

#### Scenario: Default seccomp profile
- **WHEN** container is created
- **THEN** Docker's default seccomp profile is applied

### Requirement: Read-only root filesystem option
The security module SHALL support running with read-only root filesystem.

#### Scenario: Read-only root enabled
- **WHEN** Agentfile specifies `security.read_only_root: true`
- **THEN** container root filesystem is mounted read-only (workspace remains writable)

#### Scenario: Read-only root disabled
- **WHEN** Agentfile omits `security.read_only_root` or sets it to false
- **THEN** container root filesystem is writable

### Requirement: Container user isolation
The security module SHALL ensure agent cannot access host user information.

#### Scenario: No host user mapping
- **WHEN** container runs
- **THEN** /etc/passwd and /etc/shadow contain only container users

#### Scenario: Isolated home directory
- **WHEN** agent runs
- **THEN** agent home directory is /home/agent inside container, not host home

### Requirement: Allowed IP export
The security module SHALL export the allowed IP list to a file for LD_PRELOAD library access.

#### Scenario: Export after ipset population
- **WHEN** init-firewall.sh finishes populating the allowed_ips ipset
- **THEN** it exports the ipset contents to /run/sandbox/allowed_ips

#### Scenario: File permissions
- **WHEN** /run/sandbox/allowed_ips is created
- **THEN** it is readable by all users (mode 444)

#### Scenario: Directory creation
- **WHEN** init-firewall.sh exports the allowed IPs
- **THEN** it creates /run/sandbox directory if it does not exist
