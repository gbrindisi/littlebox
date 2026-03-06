## ADDED Requirements

### Requirement: Run command executes agent in container
The CLI SHALL provide a `run` command that executes the configured agent inside a sandboxed Docker container.

#### Scenario: Basic run with defaults
- **WHEN** user executes `agentbox run` in a directory with an Agentfile
- **THEN** the CLI starts a container with the configured agent and attaches stdin/stdout/stderr

#### Scenario: Run with config override
- **WHEN** user executes `agentbox run -c custom.yaml`
- **THEN** the CLI uses the specified config file instead of Agentfile

#### Scenario: Pass arguments to agent
- **WHEN** user executes `agentbox run -- --print "fix the bug"`
- **THEN** the CLI passes arguments after `--` to the agent command inside the container

#### Scenario: Run with workspace override
- **WHEN** user executes `agentbox run --workspace /other/project`
- **THEN** the CLI mounts the specified directory as the workspace instead of current directory

### Requirement: Init command creates template Agentfile
The CLI SHALL provide an `init` command that creates a template Agentfile in the current directory.

#### Scenario: Create reference template without profile
- **WHEN** user executes `agentbox init` without --profile flag
- **THEN** the CLI creates an Agentfile with all sections commented out as a reference template

#### Scenario: Reference template lists available profiles
- **WHEN** user executes `agentbox init` without --profile flag
- **THEN** the generated file includes a header listing all available profile names

#### Scenario: Reference template lists network presets
- **WHEN** user executes `agentbox init` without --profile flag
- **THEN** the generated file shows all available network presets in comments

#### Scenario: Init with profile outputs embedded template
- **WHEN** user executes `agentbox init --profile claude-code`
- **THEN** the CLI writes the embedded Agentfile.claude-code content verbatim to Agentfile

#### Scenario: Init refuses to overwrite
- **WHEN** user executes `agentbox init` and Agentfile already exists
- **THEN** the CLI exits with error unless `--force` flag is provided

#### Scenario: Init with unknown profile
- **WHEN** user executes `agentbox init --profile unknown`
- **THEN** the CLI exits with error listing available profile names

### Requirement: Validate command checks Agentfile
The CLI SHALL provide a `validate` command that checks Agentfile syntax and configuration.

#### Scenario: Valid config
- **WHEN** user executes `agentbox validate` with a valid Agentfile
- **THEN** the CLI outputs "Agentfile is valid" and exits with code 0

#### Scenario: Invalid config
- **WHEN** user executes `agentbox validate` with an invalid Agentfile
- **THEN** the CLI outputs specific validation errors and exits with non-zero code

### Requirement: Shell command opens debug shell
The CLI SHALL provide a `shell` command that opens an interactive shell inside the container for debugging.

#### Scenario: Open debug shell
- **WHEN** user executes `agentbox shell`
- **THEN** the CLI starts the container and drops into a bash/zsh shell instead of running the agent

### Requirement: Agentfile auto-discovery
The CLI SHALL automatically discover Agentfile in the current directory when no config is specified.

#### Scenario: Agentfile found
- **WHEN** user executes `agentbox run` and `./Agentfile` exists
- **THEN** the CLI uses `./Agentfile` as configuration

#### Scenario: No Agentfile found
- **WHEN** user executes `agentbox run` and no Agentfile exists
- **THEN** the CLI exits with error suggesting `agentbox init`

### Requirement: Exit code passthrough
The CLI SHALL exit with the same exit code as the agent process inside the container.

#### Scenario: Agent exits successfully
- **WHEN** agent process exits with code 0
- **THEN** the CLI exits with code 0

#### Scenario: Agent exits with error
- **WHEN** agent process exits with code 1
- **THEN** the CLI exits with code 1

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
