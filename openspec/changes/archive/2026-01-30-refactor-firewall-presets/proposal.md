## Why

The current preset model uses hierarchical bundles (strict/standard/permissive) with implicit includes, making it hard to understand what's actually allowed. Users think in terms of services (anthropic, github, npm) not abstract security levels. The single-preset design prevents composing exactly the access needed.

## What Changes

- **BREAKING**: Replace `network.preset` (single string) with `network.presets` (list of service names)
- Remove hierarchical presets (strict, standard, permissive)
- Add service-specific presets: `anthropic`, `openai`, `google-ai`, `mistral`, `github`, `gitlab`, `bitbucket`, `npm`, `pypi`, `cargo`, `rubygems`, `huggingface`
- Make both `presets` and `allow` optional - either can be used alone or together
- Extend `allow` to accept domains, IPs, and CIDR ranges (domains already work, add explicit IP/CIDR support)
- Warn at startup when no network access is configured (valid air-gapped use case, but user should be aware)

## Capabilities

### New Capabilities

None - this modifies existing network capability.

### Modified Capabilities

- `network`: Preset configuration changes from single hierarchical value to composable list of service-specific presets

## Impact

- `internal/config/config.go`: Change `Network.Preset` from `NetworkPreset` to `[]string` named `Presets`
- `internal/config/presets.go`: Replace hierarchical presets with service-specific preset map
- `internal/config/validate.go`: Update validation for new preset format, add warning for empty network config
- `internal/container/runner.go`: Update how `ALLOWED_DOMAINS` is built from presets
- `internal/container/docker/init-firewall.sh`: Already supports CIDR, may need minor updates for IP detection
- Existing Agentfiles using `preset: standard` will break and need migration to `presets: [...]`
