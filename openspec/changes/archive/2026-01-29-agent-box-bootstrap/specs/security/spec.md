## ADDED Requirements

### Requirement: Privilege separation
The security module SHALL ensure firewall runs as root and agent runs as unprivileged user.

#### Scenario: Firewall runs as root
- **WHEN** container entrypoint starts
- **THEN** firewall setup script runs with root privileges

#### Scenario: Agent runs as unprivileged user
- **WHEN** firewall setup completes
- **THEN** agent process is executed as non-root user (uid 1000 or matched host UID)

#### Scenario: No privilege escalation
- **WHEN** agent attempts to use sudo
- **THEN** command fails (sudo access is removed)

### Requirement: Capability dropping
The security module SHALL drop all Linux capabilities after firewall setup.

#### Scenario: Drop all capabilities
- **WHEN** entrypoint transitions to agent user
- **THEN** all inheritable capabilities are dropped (--inh-caps=-all)

#### Scenario: NET_ADMIN only during init
- **WHEN** container is created
- **THEN** NET_ADMIN and NET_RAW capabilities are available only to init process

### Requirement: Sudo removal
The security module SHALL remove sudo access before running agent.

#### Scenario: Remove sudoers files
- **WHEN** firewall setup completes
- **THEN** all files in /etc/sudoers.d/ are removed

#### Scenario: Lock sudoers
- **WHEN** firewall setup completes
- **THEN** /etc/sudoers is made unreadable by non-root users

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

#### Scenario: no-new-privileges enabled
- **WHEN** container is created
- **THEN** security-opt no-new-privileges:true is set

#### Scenario: Setuid binaries ineffective
- **WHEN** agent executes a setuid binary
- **THEN** binary runs with agent's privileges, not elevated privileges

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
