## ADDED Requirements

### Requirement: Firewall initialization
The network module SHALL configure iptables rules at container startup.

#### Scenario: Firewall setup on start
- **WHEN** container entrypoint runs
- **THEN** iptables rules are applied before agent process starts

#### Scenario: DNS access allowed
- **WHEN** firewall is initialized
- **THEN** outbound DNS (UDP port 53) is allowed for domain resolution

#### Scenario: Localhost allowed
- **WHEN** firewall is initialized
- **THEN** traffic to/from localhost (127.0.0.1) is always allowed

### Requirement: Domain-based allowlist
The network module SHALL allow traffic only to configured domains.

#### Scenario: Resolve and allow domain
- **WHEN** domain is in allow list (e.g., api.anthropic.com)
- **THEN** firewall resolves domain IPs and adds them to ipset allowlist

#### Scenario: GitHub IP ranges
- **WHEN** github.com is in allow list
- **THEN** firewall fetches GitHub's published IP ranges and adds all to allowlist

#### Scenario: Block unlisted domains
- **WHEN** agent attempts connection to domain not in allowlist
- **THEN** connection is rejected with immediate feedback (ICMP admin-prohibited)

### Requirement: Default deny policy
The network module SHALL deny all traffic not explicitly allowed.

#### Scenario: Drop unknown outbound
- **WHEN** firewall is initialized
- **THEN** OUTPUT chain default policy is set to DROP

#### Scenario: Allow established connections
- **WHEN** connection was initiated to an allowed destination
- **THEN** return traffic (ESTABLISHED, RELATED) is allowed

### Requirement: Docker DNS preservation
The network module SHALL preserve Docker's internal DNS resolution.

#### Scenario: Preserve Docker DNS rules
- **WHEN** firewall flushes iptables
- **THEN** Docker's DNS NAT rules for 127.0.0.11 are preserved and restored

### Requirement: Host network access
The network module SHALL allow communication with the host machine.

#### Scenario: Host network allowed
- **WHEN** firewall is initialized
- **THEN** traffic to host network (detected from default route) is allowed

### Requirement: Firewall verification
The network module SHALL verify firewall effectiveness after setup.

#### Scenario: Verify blocked access
- **WHEN** firewall setup completes
- **THEN** attempt to reach example.com fails (verification that blocking works)

#### Scenario: Verify allowed access
- **WHEN** firewall setup completes
- **THEN** attempt to reach allowed API endpoint succeeds

#### Scenario: Fail on verification error
- **WHEN** firewall verification fails
- **THEN** container startup aborts with clear error message

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
