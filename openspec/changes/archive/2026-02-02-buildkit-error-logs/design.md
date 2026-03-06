## Context

Currently, `builder.go` uses `jsonmessage.DisplayJSONMessagesStream` to handle Docker build output. In quiet mode, output goes to `io.Discard`, so when a build fails, users see only "build failed" with no details about what went wrong.

Docker Desktop's "Builds > Error" page shows structured error information by parsing BuildKit protobuf traces. The same data is available via the Docker API - it's encoded in `JSONMessage.Aux` fields with `ID = "moby.buildkit.trace"`.

## Goals / Non-Goals

**Goals:**
- Parse BuildKit protobuf traces to extract structured error information
- Display build errors with Dockerfile context and line highlighting
- Show truncated build output with hint to use `--debug` for full logs
- Work for both base image builds and derived image builds

**Non-Goals:**
- Real-time progress display (keep existing "Building... done" pattern)
- Warning aggregation (only care about errors)
- Custom error formatting for debug mode (it already shows full output)

## Decisions

### Decision 1: Use `moby/buildkit` for protobuf types

**Choice**: Import `github.com/moby/buildkit/api/services/control` for `StatusResponse` type

**Alternatives considered**:
- Generate our own proto types: More work, harder to maintain
- Parse JSON manually: Brittle, would break on schema changes
- Capture raw output on error: Simpler but loses structure

**Rationale**: The buildkit types are stable, well-documented, and give us exactly what we need - `Vertex.Error` for the error message and `VertexLog.Data` for build output.

### Decision 2: Capture trace during build, format on error

**Choice**: Collect `StatusResponse` messages during build via `auxCallback`, only format if build fails

```
Build flow:
┌─────────────────────────────────────────────────────────────┐
│  ImageBuild() ──▶ JSONMessage stream                        │
│                       │                                     │
│         ┌─────────────┴─────────────┐                      │
│         ▼                           ▼                      │
│   jm.Aux != nil              jm.Error != nil               │
│   (buildkit trace)           (build failed)                │
│         │                           │                      │
│         ▼                           ▼                      │
│   Decode & store             Format error from             │
│   StatusResponse             collected traces              │
│                                                            │
└─────────────────────────────────────────────────────────────┘
```

**Rationale**: We need the trace data to format errors, but we don't want to parse protobuf on every message in the success case. Storing intermediate state and formatting on error is clean.

### Decision 3: Error output format

**Choice**: Show Dockerfile with line numbers, highlight failed line with `>`, show truncated output

```
● Building derived image: agentbox/build:abc123... failed

  1 | FROM agentbox/base:0.1.0
  2 | USER root
> 3 | RUN <<'SCRIPT'
  4 | set -ex
  5 | curl -fsSL https://example.com/install.sh | bash
  6 | SCRIPT

Error: exit code 1

  + curl -fsSL https://example.com/install.sh | bash
  curl: (6) Could not resolve host: example.com

... (15 more lines, use --debug for full output)
```

**Rationale**: Matches Docker Desktop UX. The Dockerfile is short for derived images (we generate it), so showing full context is fine. Truncated output keeps quiet mode quiet while giving enough info to diagnose.

### Decision 4: Line limit for truncated output

**Choice**: Show last 20 lines of build output by default

**Rationale**: Errors typically appear at the end. 20 lines is enough to see the failing command and its immediate context without overwhelming the terminal.

### Decision 5: New package for build error handling

**Choice**: Create `internal/container/builderror` package with:
- `Collector` - accumulates StatusResponse messages during build
- `Format()` - formats error with Dockerfile context when build fails

**Rationale**: Keeps `builder.go` clean, makes error handling testable in isolation.

## Risks / Trade-offs

**Risk**: BuildKit protobuf schema changes
→ **Mitigation**: Pin `moby/buildkit` version, update as needed. Schema is stable.

**Risk**: Increased binary size from buildkit dependency
→ **Mitigation**: Only importing the control API types, not the full buildkit client. Acceptable trade-off for better UX.

**Risk**: Line number highlighting might be wrong if Dockerfile format changes
→ **Mitigation**: We control Dockerfile generation in `generateDerivedDockerfile()`. Keep them in sync.

**Trade-off**: Memory usage from collecting all StatusResponse messages
→ Acceptable for build operations which are infrequent. Could optimize later by only keeping last N messages if needed.
