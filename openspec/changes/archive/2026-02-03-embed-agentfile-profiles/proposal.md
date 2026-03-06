## Why

The current profile system requires writing Go code to define profiles - a Profile struct with config.Config, ProfileComments, and init() registration. This makes it hard to contribute new profiles and difficult to see what an Agentfile will look like without running the code. By embedding actual Agentfile templates directly, contributors can just add a YAML file and the profile is automatically available.

## What Changes

- Replace Go struct-based profile definitions with embedded Agentfile templates
- Auto-discover profiles from embedded `Agentfile.*` filenames (e.g., `Agentfile.claude-code` -> profile name `claude-code`)
- Remove dynamic Agentfile generation - just print the embedded file verbatim
- Remove ProfileComments, registry pattern, and generateProfileTemplate() code
- Single unit test validates all embedded profiles load correctly

## Capabilities

### New Capabilities

None - this is a simplification of existing functionality.

### Modified Capabilities

- `profiles`: Profile system changes from Go struct registration to embedded file discovery. Profile names derived from filenames. No more dynamic generation.
- `cli`: The `init --profile` behavior changes to output embedded files verbatim instead of generating from structs.

## Impact

- **Deleted code**: `internal/config/profiles/profile.go` (Profile struct, registry), individual profile Go files (`claude_code.go`, `codex_cli.go`, `openhands.go`), `generateProfileTemplate()` in init.go
- **New files**: `Agentfile.claude-code`, `Agentfile.codex-cli`, `Agentfile.openhands` (YAML templates), `embed.go` (embed directive and discovery)
- **Simplified tests**: Single test iterates over discovered profiles instead of individual profile tests
- **Contribution model**: Add an `Agentfile.<name>` file to add a new profile - no Go code required
