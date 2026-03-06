## MODIFIED Requirements

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
