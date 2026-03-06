## Context

Currently, agentbox outputs verbose logs by default during startup:
- Image build/cache messages from `builder.go` via `fmt.Fprintf(output, ...)`
- Entrypoint logs (`[entrypoint] ...`) from `entrypoint.sh` via `echo`
- Firewall logs (`[firewall] ...`) from `init-firewall.sh` via `echo` (35+ messages)

All output flows through the container's stdout which is attached to the user's terminal. There's no mechanism to control verbosity.

The `runContainer` function in `common.go` orchestrates the flow:
1. `EnsureImage()` - outputs to `os.Stdout`
2. `EnsureDerivedImage()` - outputs to `os.Stderr`
3. `CreateContainer()` + `AttachContainer()` - streams container stdout/stderr

## Goals / Non-Goals

**Goals:**
- Default to quiet mode with minimal, user-friendly output
- Provide `--debug` flag to show verbose output (current behavior)
- Use consistent "... done" pattern for long-running operations
- Always show errors regardless of verbosity

**Non-Goals:**
- Modifying shell scripts (entrypoint.sh, init-firewall.sh)
- Adding granular log levels (info, warn, debug, trace)
- Adding persistent configuration for verbosity

## Decisions

### Decision: Control verbosity at Go CLI layer only

**Choice:** Suppress container stdout in quiet mode by not displaying it, rather than passing env vars to shell scripts.

**Rationale:**
- Simpler implementation - shell scripts remain unchanged
- Single point of control - all verbosity logic in Go
- Container still logs everything internally (useful for debugging if needed)

**Alternative considered:** Pass `AGENT_BOX_DEBUG=1` env var to container and modify shell scripts to check it. Rejected because it adds complexity across two layers.

### Decision: Use io.Writer abstraction for output control

**Choice:** Pass different `io.Writer` implementations based on verbosity:
- Quiet mode: custom writer that prints "... done" after operation completes
- Debug mode: `os.Stdout`/`os.Stderr` (current behavior)

**Rationale:** The existing code already uses `io.Writer` parameter in `EnsureImage`/`EnsureDerivedImage`, making this a natural extension.

### Decision: Suppress container stdout during sandbox setup in quiet mode

**Choice:** In quiet mode, discard container stdout during the entrypoint/firewall phase, showing only "Setting up the sandbox... done".

**Rationale:**
- The entrypoint/firewall logs are implementation details not useful to users
- A single status line provides sufficient feedback
- Errors still propagate via exit code and stderr

### Decision: Output format with bullet glyph

**Choice:** Use `●` (U+25CF BLACK CIRCLE) as prefix for status lines in quiet mode.

```
● Using cached image: agentbox/base:0.1.0
● Building image: agentbox/base:0.1.0, this may take a while... done
● Setting up the sandbox... done
● Running the agent
```

**Rationale:** Simple, works in all terminals, visually clean.

## Risks / Trade-offs

**Risk:** Users may miss important warnings in quiet mode.
- Mitigation: Errors always print to stderr regardless of verbosity.

**Risk:** Debugging becomes harder in quiet mode.
- Mitigation: `--debug` flag restores full output. Users can re-run with debug.

**Trade-off:** Container stdout is discarded in quiet mode during setup.
- Acceptable because errors propagate via exit code/stderr, and `--debug` is available.
