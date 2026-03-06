## REMOVED Requirements

### Requirement: Agent profile configuration
**Reason**: Profiles are now init-time templates only, not runtime configuration. The `profile` field is removed from Agentfile schema.
**Migration**: Run `agentbox init --profile claude-code` to generate an explicit Agentfile, then remove any `profile:` field from existing Agentfiles.

## MODIFIED Requirements

### Requirement: Environment variable passthrough
The config system SHALL support passing host environment variables to container via explicit configuration.

#### Scenario: Explicit env passthrough
- **WHEN** Agentfile specifies `agent.env_passthrough: [ANTHROPIC_API_KEY]`
- **THEN** ANTHROPIC_API_KEY from host is passed to container

#### Scenario: Glob pattern env passthrough
- **WHEN** Agentfile specifies `agent.env_passthrough: [CLAUDE_CODE_*]`
- **THEN** all environment variables matching the pattern are passed to container

#### Scenario: Missing optional env var
- **WHEN** passthrough variable is not set on host
- **THEN** variable is silently omitted from container environment
