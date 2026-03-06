## ADDED Requirements

### Requirement: Parse BuildKit protobuf traces
The container builder SHALL parse BuildKit trace messages from Docker build responses.

#### Scenario: Decode buildkit trace from Aux field
- **WHEN** Docker build response contains JSONMessage with `ID = "moby.buildkit.trace"`
- **THEN** the builder decodes the base64-encoded protobuf into `StatusResponse`

#### Scenario: Collect trace messages during build
- **WHEN** Docker build is in progress
- **THEN** the builder accumulates all `StatusResponse` messages for potential error formatting

#### Scenario: Extract vertex error
- **WHEN** a `StatusResponse` contains a `Vertex` with non-empty `Error` field
- **THEN** the builder captures the error message for display

#### Scenario: Extract vertex logs
- **WHEN** a `StatusResponse` contains `VertexLog` entries
- **THEN** the builder captures the log data (build output) for error display

### Requirement: Format build errors with Dockerfile context
The container builder SHALL display build errors with Dockerfile context and line numbers.

#### Scenario: Display Dockerfile with line numbers
- **WHEN** a build fails
- **THEN** the error output includes the Dockerfile content with line numbers in format `  N | <line>`

#### Scenario: Highlight failed line
- **WHEN** a build fails at a specific Dockerfile instruction
- **THEN** the failed line is prefixed with `>` instead of space (e.g., `> 3 | RUN ...`)

#### Scenario: Display error message
- **WHEN** a build fails
- **THEN** the error output includes the error message from the failed vertex (e.g., "exit code 1")

### Requirement: Display truncated build output
The container builder SHALL display truncated build output on failure with hint for full logs.

#### Scenario: Show last N lines of output
- **WHEN** a build fails with more than 20 lines of output
- **THEN** the error display shows only the last 20 lines of build output

#### Scenario: Show truncation hint
- **WHEN** build output is truncated
- **THEN** the error display shows "... (N more lines, use --debug for full output)"

#### Scenario: Show full output when short
- **WHEN** a build fails with 20 or fewer lines of output
- **THEN** the error display shows all output lines without truncation hint

### Requirement: Integration tests with failing builds
The container builder SHALL have integration tests that verify error display for failing builds.

#### Scenario: Test command not found error
- **WHEN** build script contains a command that doesn't exist (e.g., `nonexistent_command`)
- **THEN** integration test verifies error output includes the command and "not found" message

#### Scenario: Test exit code error
- **WHEN** build script contains a command that exits non-zero (e.g., `exit 1`)
- **THEN** integration test verifies error output includes "exit code 1"

#### Scenario: Test network error
- **WHEN** build script contains a curl to non-existent domain
- **THEN** integration test verifies error output includes the failed URL and network error

#### Scenario: Test multiline script error
- **WHEN** build script has multiple commands and a later one fails
- **THEN** integration test verifies error output shows context from earlier successful commands
