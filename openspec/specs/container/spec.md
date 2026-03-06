## ADDED Requirements

### Requirement: Container creation with mounts
The container manager SHALL create containers with configured workspace and additional mounts.

#### Scenario: Mount workspace
- **WHEN** container is created with workspace path `/project`
- **THEN** host path `/project` is bind-mounted to `/workspace` in container

#### Scenario: Mount with read-only flag
- **WHEN** mount specifies `readonly: true`
- **THEN** container mount has read-only flag set

#### Scenario: Redacted files excluded
- **WHEN** redact patterns match files in workspace
- **THEN** those files are not visible inside the container

### Requirement: Image building
The container manager SHALL build and use derived images when build script is configured.

#### Scenario: First run builds base image
- **WHEN** user runs agentbox for the first time without build_script
- **THEN** container manager builds `agentbox/base` from embedded Dockerfile

#### Scenario: First run with build script
- **WHEN** user runs agentbox for the first time with build_script configured
- **THEN** container manager builds `agentbox/base` first, then builds derived image

#### Scenario: Derived image cached
- **WHEN** derived image `agentbox/build:<hash>` exists locally
- **THEN** container manager uses cached derived image without rebuilding

#### Scenario: Base image cached
- **WHEN** `agentbox/base` image exists and no build_script configured
- **THEN** container manager uses base image without rebuilding

#### Scenario: Build progress output
- **WHEN** image build is in progress
- **THEN** container manager outputs build progress to stderr

#### Scenario: Force rebuild
- **WHEN** user specifies `--rebuild` flag
- **THEN** container manager rebuilds the derived image even if cached

### Requirement: TTY allocation
The container manager SHALL allocate a PTY for interactive sessions.

#### Scenario: Interactive mode
- **WHEN** stdin is a terminal
- **THEN** container is created with TTY and stdin attached

#### Scenario: Scripted mode
- **WHEN** stdin is not a terminal (piped)
- **THEN** container is created without TTY allocation

#### Scenario: Force TTY flag
- **WHEN** user specifies `--tty` flag
- **THEN** container is created with TTY regardless of stdin

#### Scenario: Force no-TTY flag
- **WHEN** user specifies `--no-tty` flag
- **THEN** container is created without TTY regardless of stdin

### Requirement: Signal forwarding
The container manager SHALL forward signals to the container process.

#### Scenario: SIGINT forwarding
- **WHEN** user presses Ctrl+C (SIGINT)
- **THEN** SIGINT is forwarded to the agent process in container

#### Scenario: SIGTERM forwarding
- **WHEN** process receives SIGTERM
- **THEN** SIGTERM is forwarded to the agent process in container

#### Scenario: Graceful shutdown
- **WHEN** signal is forwarded to container
- **THEN** container manager waits for agent to exit before terminating

### Requirement: Window resize handling
The container manager SHALL forward terminal resize events to the container.

#### Scenario: Terminal resize
- **WHEN** terminal window is resized (SIGWINCH)
- **THEN** container PTY size is updated to match new dimensions

### Requirement: Container cleanup
The container manager SHALL clean up containers based on configuration.

#### Scenario: Ephemeral cleanup
- **WHEN** container.keep is false and agent exits
- **THEN** container is removed immediately

#### Scenario: Keep on exit
- **WHEN** container.keep is true and agent exits
- **THEN** container is stopped but not removed

### Requirement: Exit code capture
The container manager SHALL capture and return the agent's exit code.

#### Scenario: Capture exit code
- **WHEN** agent process exits
- **THEN** container manager returns the exact exit code to caller

### Requirement: Host UID/GID matching
The container manager SHALL run the agent with matching host user IDs on Linux.

#### Scenario: Linux UID matching
- **WHEN** running on Linux host
- **THEN** container runs agent process with host user's UID/GID

#### Scenario: macOS default user
- **WHEN** running on macOS host
- **THEN** container runs with default UID 1000 (Docker Desktop handles file permissions)

#### Scenario: Explicit user override
- **WHEN** Agentfile specifies `workspace.user: 1001:1001`
- **THEN** container runs with specified UID:GID regardless of host

### Requirement: Environment injection
The container manager SHALL inject environment variables into the container from explicit configuration.

#### Scenario: Passthrough variables
- **WHEN** Agentfile specifies `agent.env_passthrough` variables
- **THEN** those variables are copied from host to container environment

#### Scenario: Static variables
- **WHEN** Agentfile specifies `agent.env` variables
- **THEN** those variables are set in container environment
