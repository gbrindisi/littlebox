## 1. Create Embedded Agentfile Templates

- [x] 1.1 Create `Agentfile.claude-code` with current claude-code profile content (copy from generated output)
- [x] 1.2 Create `Agentfile.codex-cli` with current codex-cli profile content
- [x] 1.3 Create `Agentfile.openhands` with current openhands profile content

## 2. Implement Embed-Based Profile Discovery

- [x] 2.1 Create `embed.go` with `//go:embed Agentfile.*` directive
- [x] 2.2 Implement `Names()` function that parses embedded filenames and returns sorted profile names
- [x] 2.3 Implement `Get(name string) (string, bool)` function that returns embedded file content

## 3. Update CLI Init Command

- [x] 3.1 Update `init.go` to use `profiles.Get()` returning string content
- [x] 3.2 Remove `generateProfileTemplate()` function
- [x] 3.3 Update error message for unknown profile to list names only (no descriptions)
- [x] 3.4 Update reference template generation to list profile names only (no descriptions)

## 4. Delete Old Profile Code

- [x] 4.1 Delete `profile.go` (Profile struct, registry, Register function)
- [x] 4.2 Delete `claude_code.go`
- [x] 4.3 Delete `codex_cli.go`
- [x] 4.4 Delete `openhands.go`
- [x] 4.5 Delete `all.go`

## 5. Update Tests

- [x] 5.1 Create single test that iterates over `profiles.Names()` and validates each profile
- [x] 5.2 Test validates: profile exists, content is non-empty, content parses as valid YAML config
- [x] 5.3 Delete old profile-specific tests
- [x] 5.4 Update init_test.go to match new behavior
