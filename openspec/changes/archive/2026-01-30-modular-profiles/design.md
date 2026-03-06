## Context

The current profile implementation in `internal/config/profiles.go` is a single file with a hardcoded map of profiles. Only `claude-code` exists. The `init` command in `cmd/agentbox/cmd/init.go` has template generation logic tightly coupled to this single profile, with hardcoded paths like `~/.claude`.

Path handling in Agentfile currently requires absolute paths for workspace and mounts. The `~/ ` prefix is expanded, but relative paths are not supported.

## Goals / Non-Goals

**Goals:**
- Modular profile architecture that makes adding new profiles straightforward
- Reference template generation showing all config options when `init` is run without `--profile`
- Profile-based template generation with explanatory comments when `--profile` is specified
- Relative path support in Agentfile resolved relative to Agentfile location
- Default workspace to `.` (Agentfile directory) for portability

**Non-Goals:**
- Runtime profile resolution (profiles are init-time templates only)
- Profile inheritance or composition
- Remote profile fetching
- Full openhands/codex-cli configuration (stubs only, details TBD)

## Decisions

### Decision 1: Profile struct embeds Config

**Choice:** Profile struct embeds `config.Config` directly plus metadata fields.

**Alternatives considered:**
- Mirror all Config fields in Profile (rejected: field drift risk, duplication)
- Profile as interface with methods (rejected: over-engineering for static data)

**Rationale:** Embedding ensures Profile can express any valid Config without maintaining parallel field definitions. New Config fields automatically become available in profiles.

```go
type Profile struct {
    Name        string
    Description string
    URL         string
    Comments    ProfileComments
    Config      config.Config
}
```

### Decision 2: Self-registering profiles via init()

**Choice:** Each profile in its own file with `func init() { Register(&Profile{...}) }`.

**Alternatives considered:**
- Central registry file listing all profiles (rejected: requires editing multiple files to add profile)
- Code generation (rejected: unnecessary complexity)

**Rationale:** Go's init() pattern allows adding a profile by creating a single file. Import side effects handle registration.

```
internal/config/profiles/
├── profile.go      # Profile type, registry, Get/All/Names functions
├── claude_code.go  # func init() { Register(...) }
├── openhands.go
└── codex_cli.go
```

### Decision 3: ProfileComments for section-specific documentation

**Choice:** ProfileComments struct with per-section comment fields.

**Rationale:** Different profiles need different explanations. Claude-code needs to explain `~/.claude`, openhands needs its own context. Structured comments allow the template generator to place them appropriately.

```go
type ProfileComments struct {
    Header    string  // Top of generated file
    Agent     string  // Above agent section
    Mounts    string  // Above mounts section
    Network   string  // Above network section
}
```

### Decision 4: Path resolution with Agentfile directory

**Choice:** Enhance `ExpandPaths()` to accept Agentfile directory and resolve relative paths.

**Alternatives considered:**
- Resolve at parse time in Load() (rejected: Load doesn't know usage context)
- Store base directory in Config (rejected: pollutes Config with loader metadata)

**Rationale:** Path expansion is already a separate step (`cfg.ExpandPaths()`). Adding a parameter keeps the interface clean and explicit.

```go
func (c *Config) ExpandPaths(baseDir string) {
    c.Workspace.Path = resolvePath(c.Workspace.Path, baseDir)
    for i := range c.Mounts {
        c.Mounts[i].Source = resolvePath(c.Mounts[i].Source, baseDir)
    }
}

func resolvePath(path, baseDir string) string {
    path = expandTilde(path)
    if filepath.IsAbs(path) {
        return path
    }
    return filepath.Join(baseDir, path)
}
```

### Decision 5: Reference template is dynamically generated

**Choice:** Build reference template at runtime showing all available presets and profiles.

**Alternatives considered:**
- Static template string (rejected: goes stale as presets/profiles change)

**Rationale:** Dynamic generation ensures the reference always reflects current capabilities.

## Risks / Trade-offs

**[Risk] Profile import side effects** - Profiles must be imported for registration to occur.
→ Mitigation: Create `profiles/all.go` with blank imports, import from init.go.

**[Risk] Breaking change to ExpandPaths signature** - Existing callers pass no arguments.
→ Mitigation: Single call site in common.go, straightforward update.

**[Risk] Relative path confusion** - Users might expect relative to cwd, not Agentfile.
→ Mitigation: Document clearly in reference template comments.
