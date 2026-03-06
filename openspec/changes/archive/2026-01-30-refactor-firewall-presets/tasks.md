## 1. Config Schema Changes

- [x] 1.1 Update `NetworkConfig` struct: replace `Preset NetworkPreset` with `Presets []string`
- [x] 1.2 Remove `NetworkPreset` type and constants (NetworkPresetStrict, NetworkPresetStandard, NetworkPresetPermissive)
- [x] 1.3 Update config tests in `config_test.go`

## 2. Preset Implementation

- [x] 2.1 Replace `NetworkPresets` map with service-specific presets (anthropic, openai, google-ai, mistral, github, gitlab, bitbucket, npm, pypi, cargo, rubygems, huggingface)
- [x] 2.2 Update `ApplyNetworkPreset` to iterate over `Presets` list and combine all domains
- [x] 2.3 Update preset tests in `presets_test.go`

## 3. Validation

- [x] 3.1 Update `validateNetwork` to validate each preset name in the list
- [x] 3.2 Add detection for old `preset:` format with clear migration error message
- [x] 3.3 Add warning when neither `presets` nor `allow` is configured
- [x] 3.4 Update validation tests in `validate_test.go`

## 4. Allow List Type Detection

- [x] 4.1 Add helper function to detect entry type (domain vs IP vs CIDR)
- [x] 4.2 Update firewall script to handle explicit IPs (already handles CIDR)
- [x] 4.3 Add tests for mixed allow list entries

## 5. Integration

- [x] 5.1 Update runner.go to build ALLOWED_DOMAINS from new presets format
- [x] 5.2 Update test-environment/Agentfile to use new format
- [x] 5.3 Run integration tests to verify firewall works with new presets
