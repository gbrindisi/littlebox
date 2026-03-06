## MODIFIED Requirements

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
