## ADDED Requirements

### Requirement: Quiet mode is default output
The CLI SHALL default to quiet mode, displaying minimal formatted output during startup.

#### Scenario: Cached image startup
- **WHEN** user runs `agentbox run` with cached images
- **THEN** the CLI displays:
  ```
  ● Using cached image: agentbox/base:0.1.0
  ● Using cached derived image: agentbox/build:<hash>
  ● Setting up the sandbox... done
  ● Running the agent
  ```

#### Scenario: Building image startup
- **WHEN** user runs `agentbox run` and base or derived image needs building
- **THEN** the CLI displays "Building image: <tag>, this may take a while..." before building and "done" after completion

#### Scenario: No derived image configured
- **WHEN** user runs `agentbox run` without a build script configured
- **THEN** the CLI skips the derived image message entirely

### Requirement: Debug mode shows verbose output
The CLI SHALL provide a `--debug` flag that shows all output including container logs.

#### Scenario: Debug mode enabled
- **WHEN** user runs `agentbox run --debug`
- **THEN** the CLI displays all output without bullet formatting, including:
  - Image cache/build messages
  - Docker build output (if building)
  - Container entrypoint logs (`[entrypoint] ...`)
  - Container firewall logs (`[firewall] ...`)

#### Scenario: Debug mode with shell command
- **WHEN** user runs `agentbox shell --debug`
- **THEN** the CLI displays verbose output same as run command

### Requirement: Errors always display regardless of verbosity
The CLI SHALL always display errors to stderr, independent of the verbosity mode.

#### Scenario: Build error in quiet mode
- **WHEN** image build fails in quiet mode
- **THEN** the CLI displays structured error output including:
  - Dockerfile content with line numbers
  - Visual indicator (`>`) on the failed line
  - Error message from BuildKit
  - Truncated build output (last 20 lines)
  - Hint to use `--debug` for full output if truncated

#### Scenario: Build error in debug mode
- **WHEN** image build fails in debug mode
- **THEN** the CLI displays the full build output (already streamed) plus the error message

#### Scenario: Container startup error in quiet mode
- **WHEN** container fails to start in quiet mode
- **THEN** the CLI displays the error message to stderr

#### Scenario: Sandbox setup error in quiet mode
- **WHEN** entrypoint or firewall initialization fails in quiet mode
- **THEN** the CLI displays the error and exits with non-zero code
