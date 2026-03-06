## ADDED Requirements

### Requirement: Build script configuration
The config system SHALL support a `build_script` field for installing tools at image build time.

#### Scenario: Specify build script
- **WHEN** Agentfile specifies `agent.build_script` with shell commands
- **THEN** config stores the script content for use during image building

#### Scenario: Multi-line build script
- **WHEN** Agentfile specifies `agent.build_script` as a multi-line YAML literal
- **THEN** config preserves the full script including newlines

#### Scenario: No build script
- **WHEN** Agentfile omits `agent.build_script`
- **THEN** container uses the base `agentbox/base` image without modifications

### Requirement: Derived image building
The container manager SHALL build derived images from the build script.

#### Scenario: Build derived image
- **WHEN** config contains a `build_script`
- **THEN** container manager generates a Dockerfile extending `agentbox/base` and runs the script

#### Scenario: Build script runs as root
- **WHEN** derived image is being built
- **THEN** build script executes as root user for installation flexibility

#### Scenario: Runtime user is agent
- **WHEN** derived image build completes
- **THEN** the image switches back to `agent` user for runtime execution

### Requirement: Hash-based image caching
The container manager SHALL cache derived images using a hash of the build script.

#### Scenario: Compute image tag from script hash
- **WHEN** config contains a `build_script`
- **THEN** container manager computes SHA256 hash of script and uses first 12 characters as tag

#### Scenario: Reuse cached derived image
- **WHEN** derived image `agentbox/build:<hash>` already exists locally
- **THEN** container manager uses cached image without rebuilding

#### Scenario: Rebuild on script change
- **WHEN** `build_script` content changes
- **THEN** new hash is computed and a new derived image is built

#### Scenario: Force rebuild flag
- **WHEN** user specifies `--rebuild` flag
- **THEN** container manager rebuilds derived image even if cached version exists
