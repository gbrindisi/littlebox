## 1. Profile Package Structure

- [x] 1.1 Create `internal/config/profiles/` directory
- [x] 1.2 Create `profile.go` with Profile struct (embedding config.Config), ProfileComments struct, and registry functions (Register, Get, All, Names)
- [x] 1.3 Create `all.go` with blank imports for all profile files to ensure registration

## 2. Profile Implementations

- [x] 2.1 Create `claude_code.go` with claude-code profile (migrate from existing profiles.go)
- [x] 2.2 Create `openhands.go` with openhands profile stub
- [x] 2.3 Create `codex_cli.go` with codex-cli profile stub
- [x] 2.4 Delete old `internal/config/profiles.go` after migration (blocked: requires init.go migration in task 4)

## 3. Path Resolution

- [x] 3.1 Add `resolvePath(path, baseDir string) string` function to `paths.go`
- [x] 3.2 Update `ExpandPaths()` signature to accept `baseDir string` parameter
- [x] 3.3 Update `ExpandPaths()` to resolve relative paths for workspace.path and mount sources
- [x] 3.4 Update `common.go` to pass Agentfile directory to `ExpandPaths()`
- [x] 3.5 Add tests for relative path resolution in `paths_test.go`

## 4. Init Command Changes

- [x] 4.1 Update imports in `init.go` to use new profiles package
- [x] 4.2 Implement `generateReferenceTemplate()` for init without --profile
- [x] 4.3 Implement `generateProfileTemplate(profile)` for init with --profile
- [x] 4.4 Update init command logic to dispatch between reference and profile templates
- [x] 4.5 Remove old `generateTemplate()` function and helpers after migration

## 5. Default Workspace Path

- [x] 5.1 Update `applyWorkspaceDefaults()` in `defaults.go` to default workspace.path to "." instead of cwd
- [x] 5.2 Update related tests in `defaults_test.go` if any (actually in config_test.go)

## 6. Testing

- [x] 6.1 Add tests for profile registry (register, get, all, names)
- [x] 6.2 Add tests for claude-code profile content
- [x] 6.3 Add tests for reference template generation
- [x] 6.4 Add tests for profile template generation
- [x] 6.5 Verify existing config tests still pass with path resolution changes
