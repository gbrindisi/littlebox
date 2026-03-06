## Context

The current firewall preset system uses three hierarchical presets (strict, standard, permissive) where each level implicitly includes domains from lower levels. This creates confusion - users don't know what "standard" actually allows without reading the code. The single-preset design forces users into predefined bundles rather than composing exactly the access they need.

Current state:
- `network.preset`: single string enum (strict|standard|permissive)
- `network.allow`: additional domains appended to preset
- Presets are cumulative: permissive includes standard includes strict

Target users are running autonomous coding agents (Claude Code, Codex CLI, OpenHands) that need access to AI APIs + code hosting + package registries.

## Goals / Non-Goals

**Goals:**
- Enable users to compose network access from discrete service presets
- Make preset names match how users think (service names, not security levels)
- Support all three input types in allow list: domains, IPs, CIDR ranges
- Warn users when no network access is configured (valid but potentially unintentional)
- Provide presets for common coding agent services

**Non-Goals:**
- Cloud provider presets (AWS, GCP, Azure) - too broad, users should use `allow` for specific endpoints
- Backward compatibility with old `preset:` format - clean break, clear error message
- IPv6 support in presets - current firewall blocks IPv6 by default, keep that behavior

## Decisions

### Decision 1: Config field naming
Use `presets` (plural) as a list instead of `preset` (singular) as a string.

**Rationale**: Plural name makes it obvious it's a list. Clean break from old format enables clear migration error message rather than silent behavior change.

**Alternatives considered**:
- Support both `preset` and `presets` with deprecation warning - adds complexity, delays migration pain
- Use `preset` as list (same name, different type) - confusing YAML behavior

### Decision 2: Preset granularity
One preset per logical service (anthropic, github, npm) rather than per-domain or per-use-case.

**Rationale**: Matches user mental model. "I need github access" not "I need github.com, api.github.com, raw.githubusercontent.com, ...". Presets are additive - users combine them.

**Preset list**:
| Category | Preset | Domains |
|----------|--------|---------|
| AI | `anthropic` | api.anthropic.com, anthropic.com, claude.ai |
| AI | `openai` | api.openai.com, openai.com, platform.openai.com, cdn.openai.com |
| AI | `google-ai` | generativelanguage.googleapis.com, ai.google.dev, aistudio.google.com |
| AI | `mistral` | api.mistral.ai, mistral.ai |
| Source | `github` | github.com, api.github.com, raw.githubusercontent.com, objects.githubusercontent.com, codeload.github.com, gist.githubusercontent.com |
| Source | `gitlab` | gitlab.com, registry.gitlab.com |
| Source | `bitbucket` | bitbucket.org, api.bitbucket.org |
| Packages | `npm` | registry.npmjs.org, npmjs.org, npmjs.com |
| Packages | `pypi` | pypi.org, files.pythonhosted.org |
| Packages | `cargo` | crates.io, static.crates.io, index.crates.io |
| Packages | `rubygems` | rubygems.org |
| ML | `huggingface` | huggingface.co, cdn-lfs.huggingface.co |

### Decision 3: Allow list type detection
Detect entry type at runtime by pattern matching rather than requiring typed entries.

**Rationale**: Simpler YAML, less verbose. Detection is unambiguous:
- Contains `/` → CIDR range (e.g., `192.168.1.0/24`)
- Matches IPv4 pattern `^\d+\.\d+\.\d+\.\d+$` → IP address
- Otherwise → domain name

**Alternatives considered**:
- Typed entries (`{domain: x}`, `{ip: y}`, `{cidr: z}`) - verbose, no practical benefit

### Decision 4: Empty network config behavior
Warn but allow execution when neither `presets` nor `allow` is configured.

**Rationale**: Air-gapped execution is a valid use case (local-only processing, compliance requirements). Warning ensures users don't accidentally run without network access.

Warning message: "No network presets or allow list configured. Container will have no outbound network access (except DNS)."

### Decision 5: GitHub IP ranges
Keep special handling for GitHub - fetch IP ranges from api.github.com/meta when github preset is used.

**Rationale**: GitHub uses dynamic IP ranges that can't be enumerated as static domains. Existing implementation works well.

## Risks / Trade-offs

**Breaking change for existing users** → Clear error message with migration example when old `preset:` format detected. Documentation update.

**Preset list maintenance burden** → Services change domains infrequently. Start minimal, add presets based on user requests rather than speculating.

**storage.googleapis.com not included** → Go modules use this domain, but it's shared with many GCP services. Users needing Go modules add it via `allow`. Documented in examples.
