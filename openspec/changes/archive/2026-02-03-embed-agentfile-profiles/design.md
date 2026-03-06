## Context

The current profile system uses Go structs with a registry pattern:
- Each profile is a Go file (`claude_code.go`, etc.) with an `init()` that calls `Register()`
- Profiles contain `config.Config` (the actual configuration) + `ProfileComments` (for generated output)
- `generateProfileTemplate()` (~100 lines) converts the struct to YAML with comments
- Adding a profile requires writing Go code and understanding the struct hierarchy

This creates friction for contributors and obscures what the actual Agentfile output looks like.

## Goals / Non-Goals

**Goals:**
- Make profile contribution trivial: add a YAML file, done
- Make profile content transparent: what you see is what you get
- Reduce code complexity by removing generation logic
- Maintain feature parity: `init --profile <name>` still works

**Non-Goals:**
- Adding new profiles (that's a follow-up)
- Changing the Agentfile format itself
- Adding profile metadata like descriptions (just names for now)

## Decisions

### Decision 1: Use Go embed with filename-based discovery

**Choice**: Use `//go:embed Agentfile.*` and derive profile names from filenames.

**Rationale**:
- No registration code needed - just add a file
- Filename convention is self-documenting (`Agentfile.claude-code`)
- Go's embed validates at compile time that files exist

**Alternatives considered**:
- Separate metadata file (profiles.yaml) mapping names to files -> Adds another file to maintain
- Parse metadata from YAML comments -> Adds parsing complexity, fragile

### Decision 2: Remove profile descriptions from list output

**Choice**: `init --profile unknown` just lists profile names, not descriptions.

**Rationale**:
- Descriptions add marginal value - users can look at the file
- Avoids needing metadata parsing or a separate metadata file
- Keeps the "just add a file" contribution model pure

**Alternatives considered**:
- Parse first line comment for description -> Fragile, adds parsing code
- Keep a separate profiles.yaml -> Defeats "just add a file" simplicity

### Decision 3: Reference template stays generated

**Choice**: Keep `init` (without profile) generating the reference template dynamically.

**Rationale**:
- Reference template lists available presets from `config.NetworkPresets`
- Embedding would require updating it whenever presets change
- Generation logic is simple (~40 lines) and stable

**Alternatives considered**:
- Embed reference template too -> Would get stale when presets change

## Risks / Trade-offs

- **Loss of descriptions**: Users won't see "Claude Code agent from Anthropic" in profile list -> Acceptable, can look at files
- **No compile-time validation of YAML**: Embedded files could have syntax errors -> Mitigated by unit test that parses all profiles
- **Profile names tied to filenames**: Renaming requires file rename -> This is actually a feature (explicit, obvious)
