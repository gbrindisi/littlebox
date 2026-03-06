## MODIFIED Requirements

### Requirement: Network preset implementation
The network module SHALL implement composable service-specific network presets.

#### Scenario: Composable presets list
- **WHEN** `network.presets` contains multiple preset names (e.g., `[anthropic, github, npm]`)
- **THEN** firewall allows domains from all specified presets combined

#### Scenario: Anthropic preset
- **WHEN** `network.presets` contains "anthropic"
- **THEN** firewall allows: api.anthropic.com, anthropic.com, claude.ai

#### Scenario: OpenAI preset
- **WHEN** `network.presets` contains "openai"
- **THEN** firewall allows: api.openai.com, openai.com, platform.openai.com, cdn.openai.com

#### Scenario: Google AI preset
- **WHEN** `network.presets` contains "google-ai"
- **THEN** firewall allows: generativelanguage.googleapis.com, ai.google.dev, aistudio.google.com

#### Scenario: Mistral preset
- **WHEN** `network.presets` contains "mistral"
- **THEN** firewall allows: api.mistral.ai, mistral.ai

#### Scenario: GitHub preset
- **WHEN** `network.presets` contains "github"
- **THEN** firewall allows: github.com, api.github.com, raw.githubusercontent.com, objects.githubusercontent.com, codeload.github.com, gist.githubusercontent.com, plus IP ranges from api.github.com/meta

#### Scenario: GitLab preset
- **WHEN** `network.presets` contains "gitlab"
- **THEN** firewall allows: gitlab.com, registry.gitlab.com

#### Scenario: Bitbucket preset
- **WHEN** `network.presets` contains "bitbucket"
- **THEN** firewall allows: bitbucket.org, api.bitbucket.org

#### Scenario: npm preset
- **WHEN** `network.presets` contains "npm"
- **THEN** firewall allows: registry.npmjs.org, npmjs.org, npmjs.com

#### Scenario: PyPI preset
- **WHEN** `network.presets` contains "pypi"
- **THEN** firewall allows: pypi.org, files.pythonhosted.org

#### Scenario: Cargo preset
- **WHEN** `network.presets` contains "cargo"
- **THEN** firewall allows: crates.io, static.crates.io, index.crates.io

#### Scenario: RubyGems preset
- **WHEN** `network.presets` contains "rubygems"
- **THEN** firewall allows: rubygems.org

#### Scenario: Hugging Face preset
- **WHEN** `network.presets` contains "huggingface"
- **THEN** firewall allows: huggingface.co, cdn-lfs.huggingface.co

#### Scenario: Unknown preset rejected
- **WHEN** `network.presets` contains an unknown preset name
- **THEN** validation fails with error listing available presets

#### Scenario: Old preset format rejected
- **WHEN** config uses `network.preset` (singular) instead of `network.presets` (plural)
- **THEN** validation fails with clear migration instructions

## ADDED Requirements

### Requirement: Allow list supports IPs and CIDR ranges
The network module SHALL accept IP addresses and CIDR ranges in the allow list.

#### Scenario: Allow single IP address
- **WHEN** `network.allow` contains an IP address (e.g., "10.0.0.1")
- **THEN** firewall adds the IP to the ipset allowlist

#### Scenario: Allow CIDR range
- **WHEN** `network.allow` contains a CIDR range (e.g., "192.168.1.0/24")
- **THEN** firewall adds the entire range to the ipset allowlist

#### Scenario: Allow domain name
- **WHEN** `network.allow` contains a domain name (e.g., "custom.example.com")
- **THEN** firewall resolves the domain and adds resolved IPs to the allowlist

#### Scenario: Mixed allow list
- **WHEN** `network.allow` contains domains, IPs, and CIDR ranges together
- **THEN** firewall processes each entry according to its type

### Requirement: Optional network configuration
The network module SHALL allow both presets and allow to be omitted.

#### Scenario: Only presets specified
- **WHEN** `network.presets` is set but `network.allow` is empty or omitted
- **THEN** firewall allows only domains from the specified presets

#### Scenario: Only allow specified
- **WHEN** `network.allow` is set but `network.presets` is empty or omitted
- **THEN** firewall allows only the explicitly listed domains/IPs/CIDRs

#### Scenario: Both presets and allow specified
- **WHEN** both `network.presets` and `network.allow` are set
- **THEN** firewall combines domains from presets with entries from allow list

#### Scenario: Neither presets nor allow specified
- **WHEN** neither `network.presets` nor `network.allow` is configured
- **THEN** validation warns that container will have no outbound network access
- **AND** execution continues (air-gapped is valid use case)

## REMOVED Requirements

### Requirement: Strict preset domains
**Reason**: Replaced by composable service-specific presets
**Migration**: Use `presets: [anthropic]` instead of `preset: strict`

### Requirement: Standard preset domains
**Reason**: Replaced by composable service-specific presets
**Migration**: Use `presets: [anthropic, github, npm, pypi]` instead of `preset: standard`

### Requirement: Permissive preset
**Reason**: Replaced by composable service-specific presets
**Migration**: Use explicit presets list for needed services, or use `allow` for custom domains
