## 1. Directory Structure

- [x] 1.1 Rename directory `cmd/agent-box/` to `cmd/agentbox/`

## 2. Go Module and Imports

- [x] 2.1 Update module path in `go.mod` from `github.com/gbrindisi/agent-box` to `github.com/gbrindisi/agentbox`
- [x] 2.2 Update all import paths in Go files from `github.com/gbrindisi/agent-box/...` to `github.com/gbrindisi/agentbox/...` (28 files, 37 occurrences)
- [x] 2.3 Run `go mod tidy` to verify imports resolve correctly

## 3. Docker Image Namespace

- [x] 3.1 Update image tag generation in `internal/container/builder.go` from `agent-box/base` to `agentbox/base`
- [x] 3.2 Update Docker image references in test files from `agent-box/base` to `agentbox/base`
- [x] 3.3 Update Docker image references in `internal/container/builderror_integration_test.go`

## 4. CLI Command Metadata

- [x] 4.1 Update cobra `Use` field in `cmd/agentbox/cmd/root.go` from `agent-box` to `agentbox`
- [x] 4.2 Update cobra `Long` description in `cmd/agentbox/cmd/root.go`
- [x] 4.3 Update version template in `cmd/agentbox/cmd/root.go` from `agent-box version` to `agentbox version`
- [x] 4.4 Update command examples in `cmd/agentbox/cmd/validate.go` from `agent-box validate` to `agentbox validate`
- [x] 4.5 Update command examples in `cmd/agentbox/cmd/shell.go` from `agent-box shell` to `agentbox shell`
- [x] 4.6 Update command examples in `cmd/agentbox/cmd/init.go` from `agent-box init` to `agentbox init`
- [x] 4.7 Update command examples in `cmd/agentbox/cmd/run.go` from `agent-box run` to `agentbox run`

## 5. Embedded Files and Scripts

- [x] 5.1 Update error messages in `internal/container/docker/libsandbox.c` from `agent-box:` to `agentbox:`
- [x] 5.2 Update comments in `internal/container/docker/entrypoint.sh` from `agent-box` to `agentbox`
- [x] 5.3 Update comments in `internal/container/docker/init-firewall.sh` from `agent-box` to `agentbox`

## 6. Configuration Files

- [x] 6.1 Update Agentfile header comment from `# agent-box configuration` to `# agentbox configuration`
- [x] 6.2 Update GitHub URL in Agentfile from `github.com/gbrindisi/agent-box` to `github.com/gbrindisi/agentbox`
- [x] 6.3 Update profile templates in `internal/config/profiles/Agentfile.*` header comments
- [x] 6.4 Update GitHub URLs in profile templates from `github.com/gbrindisi/agent-box` to `github.com/gbrindisi/agentbox`

## 7. Test Files

- [x] 7.1 Update test expectations in `cmd/agentbox/cmd/init_test.go` from `agent-box` to `agentbox`
- [x] 7.2 Update test data in `internal/output/output_test.go` from `agent-box/base` to `agentbox/base`
- [x] 7.3 Update test expectations in `internal/container/builder_test.go` from `agent-box/base` to `agentbox/base`
- [x] 7.4 Update test data in `internal/container/builderror/formatter_test.go` from `agent-box/base` to `agentbox/base`

## 8. Code Comments and Documentation Strings

- [x] 8.1 Update package comment in `internal/config/config.go` from `agent-box` to `agentbox`
- [x] 8.2 Update comment in `internal/config/validate.go` warning output from `[agent-box]` to `[agentbox]`
- [x] 8.3 Update comment in `internal/container/builder.go` format documentation from `agent-box/base` to `agentbox/base`

## 9. Main Documentation

- [x] 9.1 Update README.md title from "Agent Box" to "Agentbox"
- [x] 9.2 Update README.md installation instructions (binary name from `agent-box` to `agentbox`)
- [x] 9.3 Update README.md command examples from `agent-box` to `agentbox`
- [x] 9.4 Update README.md GitHub URLs from `github.com/gbrindisi/agent-box` to `github.com/gbrindisi/agentbox`

## 10. OpenSpec Specifications

- [x] 10.1 Update all references in `openspec/specs/` from `agent-box` to `agentbox`
- [x] 10.2 Update image references in spec files from `agent-box/base` to `agentbox/base`

## 11. OpenSpec Archived Changes

- [x] 11.1 Update all references in `openspec/changes/archive/` from `agent-box` to `agentbox`
- [x] 11.2 Update image references in archived changes from `agent-box/base` to `agentbox/base`
- [x] 11.3 Update GitHub URLs in archived changes from `github.com/gbrindisi/agent-box` to `github.com/gbrindisi/agentbox`

## 12. Littlefactory Files

- [x] 12.1 Update references in `.littlefactory/progress.md` from `agent-box` to `agentbox`
- [x] 12.2 Update references in `.littlefactory/tasks.json` from `agent-box` to `agentbox`

## 13. Build and Verification

- [x] 13.1 Run `go build -o agentbox ./cmd/agentbox` and verify success
- [x] 13.2 Run `go test ./...` and verify all tests pass
- [x] 13.3 Run `grep -r "agent-box" . --exclude-dir=.git` and verify no remaining references (except false positives)
- [x] 13.4 Run `agentbox --version` and verify binary works
