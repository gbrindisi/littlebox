## Why

The current CLI output is verbose by default, showing detailed entrypoint and firewall initialization logs that are useful for debugging but noisy for everyday use. Users want a clean, minimal output that shows progress without implementation details.

## What Changes

- Add `--debug` flag to the `run` command that shows verbose output (current behavior)
- Default to quiet mode with minimal, formatted output using bullet glyphs
- Show "... done" progress pattern for long-running operations (building images, setting up sandbox)
- Suppress container stdout (entrypoint/firewall logs) in quiet mode
- Always show errors regardless of verbosity level

## Capabilities

### New Capabilities
- `output-verbosity`: Controls CLI output verbosity with quiet (default) and debug modes

### Modified Capabilities
- `cli`: Adding `--debug` flag to the run command

## Impact

- `cmd/agentbox/cmd/run.go`: Add `--debug` flag
- `internal/container/builder.go`: Modify output formatting based on verbosity
- Container attach/exec code: Suppress container stdout in quiet mode
- User experience: Cleaner default output, verbose logs available with `--debug`
