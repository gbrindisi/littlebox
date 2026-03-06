## MODIFIED Requirements

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

### Requirement: Environment injection
The container manager SHALL inject environment variables into the container from explicit configuration.

#### Scenario: Passthrough variables
- **WHEN** Agentfile specifies `agent.env_passthrough` variables
- **THEN** those variables are copied from host to container environment

#### Scenario: Static variables
- **WHEN** Agentfile specifies `agent.env` variables
- **THEN** those variables are set in container environment
