## Why

The current profile system only supports claude-code and has hardcoded paths in the template generation. We need to support multiple agent profiles (openhands, codex-cli) and make it easy to add more in the future. Additionally, `agentbox init` without a profile flag should generate a reference template showing all available configuration options.

## What Changes

- Restructure profiles into separate files with a registry pattern (`internal/config/profiles/`)
- Extend Profile struct to embed full Config (no field drift) plus metadata and comments
- Change `agentbox init` (no flags) to generate a commented reference template showing all sections and available presets
- Change `agentbox init --profile X` to generate working config from profile with explanatory comments
- Add relative path support in Agentfile (resolved relative to Agentfile location)
- Default workspace path to `.` (Agentfile directory) instead of requiring absolute paths
- Add openhands and codex-cli profile stubs (configuration details TBD)

## Capabilities

### New Capabilities
- `profiles`: Modular profile system with registry pattern, self-registering profiles, and template generation

### Modified Capabilities
- `cli`: Init command behavior changes (reference template vs profile-based generation)
- `config`: Path resolution enhanced to support relative paths and `~/` expansion relative to Agentfile location

## Impact

- `internal/config/profiles.go` - Replaced by new `internal/config/profiles/` package
- `cmd/agentbox/cmd/init.go` - Template generation logic refactored
- `internal/config/paths.go` - Enhanced to resolve relative paths
- `internal/config/defaults.go` - Workspace default changes to `.`
- `cmd/agentbox/cmd/common.go` - Pass Agentfile directory to path expansion
