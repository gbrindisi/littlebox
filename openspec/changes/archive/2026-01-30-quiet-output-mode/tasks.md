## 1. Add Debug Flag to CLI

- [x] 1.1 Add `--debug` flag to `run` command in `cmd/agentbox/cmd/run.go`
- [x] 1.2 Add `--debug` flag to `shell` command in `cmd/agentbox/cmd/shell.go`
- [x] 1.3 Add `debug` field to `runOptions` struct in `cmd/agentbox/cmd/common.go`

## 2. Create Output Formatting Utilities

- [x] 2.1 Create `internal/output/output.go` with verbosity-aware output functions
- [x] 2.2 Implement `StatusWriter` that prints "message... done" pattern for long operations
- [x] 2.3 Implement `BulletPrint` function for `● message` format in quiet mode

## 3. Update Image Building Output

- [x] 3.1 Modify `EnsureImage` in `builder.go` to accept verbosity parameter
- [x] 3.2 Modify `EnsureDerivedImage` in `builder.go` to accept verbosity parameter
- [x] 3.3 In quiet mode: show bullet-prefixed cache messages or "building... done" pattern
- [x] 3.4 In debug mode: show current verbose output including docker build progress

## 4. Update Container Startup Output

- [x] 4.1 Modify `runContainer` in `common.go` to use verbosity-aware output
- [x] 4.2 In quiet mode: print "Setting up the sandbox..." before attach, "done" after entrypoint completes
- [x] 4.3 In quiet mode: suppress container stdout during sandbox setup phase
- [x] 4.4 In quiet mode: print "Running the agent" before handing control to agent
- [x] 4.5 In debug mode: stream all container output as currently done

## 5. Testing

- [x] 5.1 Manual test: verify quiet mode output matches expected format
- [x] 5.2 Manual test: verify `--debug` shows verbose output
- [x] 5.3 Manual test: verify errors display in quiet mode
