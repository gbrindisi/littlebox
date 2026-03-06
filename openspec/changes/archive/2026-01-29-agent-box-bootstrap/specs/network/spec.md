## ADDED Requirements

### Requirement: Firewall initialization
The network module SHALL configure iptables rules at container startup.

#### Scenario: Firewall setup on start
- **WHEN** container entrypoint runs
- **THEN** iptables rules are applied before agent process starts

#### Scenario: DNS access allowed
- **WHEN** firewall is initialized
- **THEN** outbound DNS (UDP port 53) is allowed for domain resolution

#### Scenario: Localhost allowed
- **WHEN** firewall is initialized
- **THEN** traffic to/from localhost (127.0.0.1) is always allowed

### Requirement: Domain-based allowlist
The network module SHALL allow traffic only to configured domains.

#### Scenario: Resolve and allow domain
- **WHEN** domain is in allow list (e.g., api.anthropic.com)
- **THEN** firewall resolves domain IPs and adds them to ipset allowlist

#### Scenario: GitHub IP ranges
- **WHEN** github.com is in allow list
- **THEN** firewall fetches GitHub's published IP ranges and adds all to allowlist

#### Scenario: Block unlisted domains
- **WHEN** agent attempts connection to domain not in allowlist
- **THEN** connection is rejected with immediate feedback (ICMP admin-prohibited)

### Requirement: Default deny policy
The network module SHALL deny all traffic not explicitly allowed.

#### Scenario: Drop unknown outbound
- **WHEN** firewall is initialized
- **THEN** OUTPUT chain default policy is set to DROP

#### Scenario: Allow established connections
- **WHEN** connection was initiated to an allowed destination
- **THEN** return traffic (ESTABLISHED, RELATED) is allowed

### Requirement: Docker DNS preservation
The network module SHALL preserve Docker's internal DNS resolution.

#### Scenario: Preserve Docker DNS rules
- **WHEN** firewall flushes iptables
- **THEN** Docker's DNS NAT rules for 127.0.0.11 are preserved and restored

### Requirement: Host network access
The network module SHALL allow communication with the host machine.

#### Scenario: Host network allowed
- **WHEN** firewall is initialized
- **THEN** traffic to host network (detected from default route) is allowed

### Requirement: Firewall verification
The network module SHALL verify firewall effectiveness after setup.

#### Scenario: Verify blocked access
- **WHEN** firewall setup completes
- **THEN** attempt to reach example.com fails (verification that blocking works)

#### Scenario: Verify allowed access
- **WHEN** firewall setup completes
- **THEN** attempt to reach allowed API endpoint succeeds

#### Scenario: Fail on verification error
- **WHEN** firewall verification fails
- **THEN** container startup aborts with clear error message

### Requirement: Network preset implementation
The network module SHALL implement standard network presets.

#### Scenario: Strict preset domains
- **WHEN** network.preset is "strict"
- **THEN** only agent API endpoint is allowed (e.g., api.anthropic.com for claude-code)

#### Scenario: Standard preset domains
- **WHEN** network.preset is "standard"
- **THEN** agent API, github.com IP ranges, npm registry, and common package registries are allowed

#### Scenario: Permissive preset
- **WHEN** network.preset is "permissive"
- **THEN** no firewall rules are applied (all outbound traffic allowed)
