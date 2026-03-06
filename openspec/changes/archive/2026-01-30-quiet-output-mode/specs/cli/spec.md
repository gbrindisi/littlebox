## ADDED Requirements

### Requirement: Debug flag for verbose output
The CLI SHALL provide a `--debug` flag on the `run` and `shell` commands to enable verbose output mode.

#### Scenario: Run with debug flag
- **WHEN** user executes `agentbox run --debug`
- **THEN** the CLI shows all verbose output including container logs

#### Scenario: Shell with debug flag
- **WHEN** user executes `agentbox shell --debug`
- **THEN** the CLI shows all verbose output including container logs

#### Scenario: Debug flag in help
- **WHEN** user executes `agentbox run --help`
- **THEN** the output includes `--debug` flag with description "show verbose output"
