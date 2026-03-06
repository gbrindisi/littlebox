## 1. Config Changes

- [x] 1.1 Add `BuildScript` field to `AgentConfig` struct in `internal/config/config.go`
- [x] 1.2 Remove `Image` field from `AgentConfig` struct
- [x] 1.3 Remove `Profile` field from `AgentConfig` struct
- [x] 1.4 Update YAML tags for `build_script` field
- [x] 1.5 Update config validation to reject `image` and `profile` fields with migration error message

## 2. Profile Simplification

- [x] 2.1 Remove `Image` field from `Profile` struct in `internal/config/profiles.go`
- [x] 2.2 Remove `aider` profile from `BuiltinProfiles` map
- [x] 2.3 Add `BuildScript` field to `Profile` struct
- [x] 2.4 Update `claude-code` profile with build script: `curl -fsSL https://claude.ai/install.sh | bash`
- [x] 2.5 Remove `ApplyProfile` and `ApplyProfileToConfig` functions (no longer needed at runtime)

## 3. Init Command Update

- [x] 3.1 Update `generateTemplate()` in `cmd/agentbox/cmd/init.go` to output explicit Agentfile
- [x] 3.2 Remove `profile:` field from generated template
- [x] 3.3 Add `build_script:` field to generated template with claude-code install script
- [x] 3.4 Include explicit `command`, `args`, and `env_passthrough` in generated template
- [x] 3.5 Update help text to reflect that profiles are templates only

## 4. Derived Image Building

- [x] 4.1 Create `DerivedImageTag(buildScript string) string` function that computes SHA256 hash
- [x] 4.2 Create `BuildDerivedImage()` function in `internal/container/builder.go`
- [x] 4.3 Generate Dockerfile content: `FROM agentbox/base`, `USER root`, `RUN script`, `USER agent`
- [x] 4.4 Implement image existence check for derived image tag
- [x] 4.5 Add build progress output to stderr during derived image build

## 5. Container Runner Update

- [x] 5.1 Update `CreateContainer()` in `runner.go` to use derived image tag when `BuildScript` is set
- [x] 5.2 Fall back to `ImageTag()` (base image) when no `BuildScript` configured
- [x] 5.3 Call `BuildDerivedImage()` before container creation if image doesn't exist

## 6. CLI Flag for Rebuild

- [x] 6.1 Add `--rebuild` flag to `run` command in `cmd/agentbox/cmd/run.go`
- [x] 6.2 Pass rebuild flag through to container manager
- [x] 6.3 Force rebuild of derived image when flag is set

## 7. Tests

- [x] 7.1 Add unit test for `DerivedImageTag()` hash computation
- [x] 7.2 Add unit test for config parsing with `build_script` field
- [x] 7.3 Add unit test for config rejection of `image` and `profile` fields
- [x] 7.4 Update existing profile tests to reflect new behavior
- [x] 7.5 Add integration test for derived image build and caching

## 8. Cleanup

- [x] 8.1 Remove unused `cfg.Agent.Image` references throughout codebase
- [x] 8.2 Update README with new Agentfile schema (no `profile`, add `build_script`)
- [x] 8.3 Update example Agentfiles in documentation
