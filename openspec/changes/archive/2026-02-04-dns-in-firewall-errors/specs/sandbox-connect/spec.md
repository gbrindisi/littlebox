## MODIFIED Requirements

### Requirement: Blocked connection messaging
The sandbox-connect module SHALL print human-readable error messages for blocked connections.

#### Scenario: Connection to blocked IP with hostname
- **WHEN** connect() is called with a destination IP not in the allowed list
- **AND** the IP has a cached hostname from recent DNS lookup
- **THEN** libsandbox.so prints an error message to stderr
- **AND** the message includes both the hostname and the blocked IP address
- **AND** the message explains this is a sandbox firewall restriction
- **AND** the message indicates the restriction is not bypassable
- **AND** connect() returns -1 with errno set to ECONNREFUSED

#### Scenario: Connection to blocked IP without hostname
- **WHEN** connect() is called with a destination IP not in the allowed list
- **AND** no cached hostname is available for the IP
- **THEN** libsandbox.so prints an error message to stderr
- **AND** the message includes the blocked IP address only
- **AND** the message explains this is a sandbox firewall restriction
- **AND** the message indicates the restriction is not bypassable
- **AND** connect() returns -1 with errno set to ECONNREFUSED

#### Scenario: Connection to allowed IP
- **WHEN** connect() is called with a destination IP in the allowed list
- **THEN** libsandbox.so passes through to the real connect() syscall
- **AND** no message is printed to stderr

#### Scenario: Non-TCP connections pass through
- **WHEN** connect() is called for a non-TCP socket (e.g., Unix domain socket)
- **THEN** libsandbox.so passes through to the real connect() without checking

### Requirement: Error message format
The sandbox-connect module SHALL use a consistent, informative error message format.

#### Scenario: Error message with hostname
- **WHEN** a connection is blocked and hostname is available
- **THEN** the message follows the format: "agentbox: Connection to <hostname> (<IP>) blocked by sandbox firewall. This is not bypassable."

#### Scenario: Error message without hostname
- **WHEN** a connection is blocked and hostname is not available
- **THEN** the message follows the format: "agentbox: Connection to <IP> blocked by sandbox firewall. This is not bypassable."
