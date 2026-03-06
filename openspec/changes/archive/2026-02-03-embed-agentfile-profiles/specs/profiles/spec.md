## MODIFIED Requirements

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

## REMOVED Requirements

### Requirement: Profile embeds full Config
**Reason**: Profiles are now embedded YAML files, not Go structs with config.Config
**Migration**: Profile content is accessed via profiles.Get(name) which returns the YAML string

### Requirement: Profile comments for generated Agentfile
**Reason**: Comments are now embedded directly in the Agentfile templates, not stored as ProfileComments struct
**Migration**: Edit the Agentfile.* template files directly to modify comments

### Requirement: Built-in claude-code profile
**Reason**: Replaced by embedded Agentfile.claude-code file
**Migration**: Profile still exists, accessed via profiles.Get("claude-code")

### Requirement: Built-in openhands profile
**Reason**: Replaced by embedded Agentfile.openhands file
**Migration**: Profile still exists, accessed via profiles.Get("openhands")

### Requirement: Built-in codex-cli profile
**Reason**: Replaced by embedded Agentfile.codex-cli file
**Migration**: Profile still exists, accessed via profiles.Get("codex-cli")
