## ADDED Requirements

### Requirement: Connection interception
The sandbox-connect module SHALL intercept outbound TCP connections via LD_PRELOAD.

#### Scenario: Library loaded for agent processes
- **WHEN** entrypoint drops privileges and executes agent command
- **THEN** LD_PRELOAD is set to load libsandbox.so

#### Scenario: Connect syscall intercepted
- **WHEN** a process calls connect() for a TCP socket
- **THEN** libsandbox.so intercepts the call before it reaches the kernel

### Requirement: Blocked connection messaging
The sandbox-connect module SHALL print human-readable error messages for blocked connections.

#### Scenario: Connection to blocked IP
- **WHEN** connect() is called with a destination IP not in the allowed list
- **THEN** libsandbox.so prints an error message to stderr
- **AND** the message includes the blocked IP address
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

### Requirement: Allowed IP list loading
The sandbox-connect module SHALL load the allowed IP list from a file.

#### Scenario: Load on first connect
- **WHEN** the first connect() call is intercepted
- **THEN** libsandbox.so reads /run/sandbox/allowed_ips
- **AND** parses and caches the allowed IPs in memory

#### Scenario: File not found fallback
- **WHEN** /run/sandbox/allowed_ips does not exist
- **THEN** libsandbox.so allows all connections (fail-open for compatibility)
- **AND** logs a warning to stderr

#### Scenario: CIDR range support
- **WHEN** allowed_ips file contains CIDR notation (e.g., 192.168.1.0/24)
- **THEN** libsandbox.so correctly matches IPs within that range

### Requirement: Error message format
The sandbox-connect module SHALL use a consistent, informative error message format.

#### Scenario: Error message content
- **WHEN** a connection is blocked
- **THEN** the message follows the format: "agentbox: Connection to <IP> blocked by sandbox firewall. This is not bypassable."
