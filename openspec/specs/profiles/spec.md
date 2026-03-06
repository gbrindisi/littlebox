## ADDED Requirements

### Requirement: Profile registry with self-registration
The profiles package SHALL discover profiles from embedded `Agentfile.*` files at compile time.

#### Scenario: Profile discovered from embedded file
- **WHEN** an `Agentfile.claude-code` file exists in the profiles package
- **THEN** the profile `claude-code` is available via profiles.Get("claude-code")

#### Scenario: Get non-existent profile
- **WHEN** profiles.Get("unknown") is called
- **THEN** it returns empty string and false

#### Scenario: List all profiles
- **WHEN** profiles.Names() is called
- **THEN** it returns all profile names derived from embedded Agentfile.* filenames

#### Scenario: List profile names sorted
- **WHEN** profiles.Names() is called
- **THEN** it returns profile names in alphabetical order

### Requirement: Profile content is embedded Agentfile
The profiles.Get() function SHALL return the raw content of the embedded Agentfile template.

#### Scenario: Get profile content
- **WHEN** profiles.Get("claude-code") is called
- **THEN** it returns the exact content of the embedded Agentfile.claude-code file

#### Scenario: Profile content is valid YAML
- **WHEN** profiles.Get(name) returns content for any profile
- **THEN** the content is parseable as a valid config.Config YAML

### Requirement: Built-in claude-code profile
The profiles package SHALL include a claude-code profile for Claude Code agent.

#### Scenario: Claude-code profile exists
- **WHEN** profiles.Get("claude-code") is called
- **THEN** it returns an embedded Agentfile configured for Claude Code

#### Scenario: Claude-code network presets
- **WHEN** claude-code profile content is parsed
- **THEN** network.presets includes anthropic, github, npm, pypi

#### Scenario: Claude-code mounts
- **WHEN** claude-code profile content is parsed
- **THEN** mounts includes ~/.claude mounted to /home/agent/.claude

### Requirement: Built-in openhands profile
The profiles package SHALL include an openhands profile stub.

#### Scenario: Openhands profile exists
- **WHEN** profiles.Get("openhands") is called
- **THEN** it returns an embedded Agentfile for OpenHands agent

### Requirement: Built-in codex-cli profile
The profiles package SHALL include a codex-cli profile stub.

#### Scenario: Codex-cli profile exists
- **WHEN** profiles.Get("codex-cli") is called
- **THEN** it returns an embedded Agentfile for Codex CLI agent
