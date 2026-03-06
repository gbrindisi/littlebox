## Why

When a `build_script` fails during `--rebuild`, the error output is discarded in quiet mode, leaving users with no information about what went wrong. Docker Desktop shows structured build errors with Dockerfile context and line numbers - we should provide similar UX in agentbox.

## What Changes

- Parse BuildKit protobuf traces from `JSONMessage.Aux` to extract structured error information
- Display build errors with Dockerfile context showing line numbers and highlighting the failed step
- Show truncated build output (last N lines) with hint to use `--debug` for full logs
- Add `moby/buildkit` dependency for protobuf type definitions

## Capabilities

### New Capabilities
- `build-errors`: Structured build error parsing and display for BuildKit-based image builds

### Modified Capabilities
- `output-verbosity`: Add requirement for structured build error display in quiet mode (currently just says "display error message" without specifying format)

## Impact

- **Code**: `internal/container/builder.go` - add BuildKit trace parser and error formatter
- **Dependencies**: Add `github.com/moby/buildkit` for `controlapi.StatusResponse` protobuf types
- **Output**: Changes error output format for build failures (improvement, not breaking)
- **Testing**: Need integration tests with intentionally failing build scripts
