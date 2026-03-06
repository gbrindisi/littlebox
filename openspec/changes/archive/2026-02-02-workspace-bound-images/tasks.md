## 1. Update DerivedImageTag Function

- [x] 1.1 Modify `DerivedImageTag` in `internal/container/builder.go` to accept `workspacePath` parameter
- [x] 1.2 Update hash computation to include workspace path with null-byte separator
- [x] 1.3 Update `DerivedImageTag` tests in `internal/container/builder_test.go`

## 2. Update EnsureDerivedImage

- [x] 2.1 Modify `EnsureDerivedImage` in `internal/container/builder.go` to accept `workspacePath` parameter
- [x] 2.2 Pass `workspacePath` through to `DerivedImageTag` call

## 3. Wire Up Workspace Path

- [x] 3.1 Update `runContainer` in `cmd/agentbox/cmd/common.go` to pass `cfg.Workspace.Path` to `EnsureDerivedImage`

## 4. Verification

- [x] 4.1 Run existing tests to ensure no regressions
- [x] 4.2 Manual test: build image in workspace1, verify different hash when running from workspace2
