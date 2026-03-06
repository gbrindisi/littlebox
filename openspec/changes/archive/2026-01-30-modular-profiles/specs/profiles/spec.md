## ADDED Requirements

### Requirement: Profile registry with self-registration
The profiles package SHALL provide a registry where profiles register themselves via init().

#### Scenario: Profile self-registers on import
- **WHEN** a profile file with init() is imported
- **THEN** the profile is available via profiles.Get(name)

#### Scenario: Get non-existent profile
- **WHEN** profiles.Get("unknown") is called
- **THEN** it returns nil

#### Scenario: List all profiles
- **WHEN** profiles.All() is called
- **THEN** it returns a map of all registered profiles

#### Scenario: List profile names sorted
- **WHEN** profiles.Names() is called
- **THEN** it returns profile names in alphabetical order

### Requirement: Profile embeds full Config
The Profile struct SHALL embed config.Config to represent any valid Agentfile configuration.

#### Scenario: Profile with all config sections
- **WHEN** a profile is defined with Agent, Workspace, Mounts, Network, and Container config
- **THEN** all sections are accessible via profile.Config

#### Scenario: Profile with metadata
- **WHEN** a profile is defined
- **THEN** it includes Name, Description, and optional URL fields

### Requirement: Profile comments for generated Agentfile
The Profile struct SHALL include ProfileComments for section-specific documentation in generated files.

#### Scenario: Comments in generated output
- **WHEN** a profile with Comments.Agent set is used to generate an Agentfile
- **THEN** the generated file includes that comment above the agent section

#### Scenario: Empty comments omitted
- **WHEN** a profile has empty Comments.Mounts
- **THEN** no comment is added above the mounts section in generated output

### Requirement: Built-in claude-code profile
The profiles package SHALL include a claude-code profile for Claude Code agent.

#### Scenario: Claude-code profile exists
- **WHEN** profiles.Get("claude-code") is called
- **THEN** it returns a profile configured for Claude Code

#### Scenario: Claude-code network presets
- **WHEN** claude-code profile is retrieved
- **THEN** Config.Network.Presets includes anthropic, github, npm, pypi

#### Scenario: Claude-code mounts
- **WHEN** claude-code profile is retrieved
- **THEN** Config.Mounts includes ~/.claude mounted to /home/agent/.claude

### Requirement: Built-in openhands profile
The profiles package SHALL include an openhands profile stub.

#### Scenario: Openhands profile exists
- **WHEN** profiles.Get("openhands") is called
- **THEN** it returns a profile for OpenHands agent

### Requirement: Built-in codex-cli profile
The profiles package SHALL include a codex-cli profile stub.

#### Scenario: Codex-cli profile exists
- **WHEN** profiles.Get("codex-cli") is called
- **THEN** it returns a profile for Codex CLI agent
