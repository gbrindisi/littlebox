## MODIFIED Requirements

### Requirement: Hash-based image caching
The container manager SHALL cache derived images using a hash of the build script AND workspace path.

#### Scenario: Compute image tag from script and workspace hash
- **WHEN** config contains a `build_script`
- **THEN** container manager computes SHA256 hash of script content concatenated with workspace absolute path (null-byte separated) and uses first 12 characters as tag

#### Scenario: Reuse cached derived image
- **WHEN** derived image `agentbox/build:<hash>` already exists locally
- **AND** hash matches current build script AND workspace path
- **THEN** container manager uses cached image without rebuilding

#### Scenario: Rebuild on script change
- **WHEN** `build_script` content changes
- **THEN** new hash is computed and a new derived image is built

#### Scenario: Rebuild on workspace change
- **WHEN** same `build_script` is used from a different workspace path
- **THEN** new hash is computed and a new derived image is built

#### Scenario: Force rebuild flag
- **WHEN** user specifies `--rebuild` flag
- **THEN** container manager rebuilds derived image even if cached version exists
