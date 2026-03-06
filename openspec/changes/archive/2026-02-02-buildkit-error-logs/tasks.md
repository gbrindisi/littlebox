## 1. Dependencies

- [x] 1.1 Add `github.com/moby/buildkit` dependency for `controlapi.StatusResponse` protobuf types
- [x] 1.2 Verify protobuf decoding works with current Docker API version

## 2. BuildKit Trace Parser

- [x] 2.1 Create `internal/container/builderror` package
- [x] 2.2 Implement `Collector` struct to accumulate `StatusResponse` messages
- [x] 2.3 Implement trace decoding: base64 JSON -> bytes -> protobuf unmarshal
- [x] 2.4 Implement extraction of `Vertex.Error` from collected traces
- [x] 2.5 Implement extraction of `VertexLog.Data` (build output lines) from collected traces
- [x] 2.6 Add unit tests for trace parsing with sample BuildKit messages

## 3. Error Formatter

- [x] 3.1 Implement `Format()` function that takes Dockerfile content and collected traces
- [x] 3.2 Format Dockerfile with line numbers (`  N | <line>`)
- [x] 3.3 Highlight failed line with `>` prefix based on vertex info
- [x] 3.4 Format error message section
- [x] 3.5 Format truncated build output (last 20 lines) with count hint
- [x] 3.6 Add unit tests for error formatting with various Dockerfile/output combinations

## 4. Integration with Builder

- [x] 4.1 Modify `BuildDerivedImage` to use `auxCallback` parameter of `DisplayJSONMessagesStream`
- [x] 4.2 Create collector, pass decode callback to accumulate traces during build
- [x] 4.3 On build error, format and return structured error instead of raw message
- [x] 4.4 Pass generated Dockerfile to formatter for context display
- [x] 4.5 Apply same changes to `BuildImage` (base image builds)
- [x] 4.6 Update `EnsureDerivedImage` to display formatted error in quiet mode
- [x] 4.7 Update `EnsureImage` to display formatted error in quiet mode

## 5. Integration Tests

- [x] 5.1 Create test helper to build derived images with intentionally failing scripts
- [x] 5.2 Test: command not found error (`nonexistent_command_xyz`)
- [x] 5.3 Test: explicit exit code error (`exit 1`)
- [x] 5.4 Test: network error (`curl https://nonexistent.invalid/install.sh`)
- [x] 5.5 Test: multiline script with failure on later line
- [x] 5.6 Verify error output includes Dockerfile with line numbers
- [x] 5.7 Verify error output highlights the correct failed line
- [x] 5.8 Verify truncation works when output exceeds 20 lines
- [x] 5.9 Verify hint message shows correct line count and --debug suggestion
