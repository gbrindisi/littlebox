## Context

The repository remote is `github.com/gbrindisi/agentbox` but all code, documentation, and artifacts use `agent-box` (hyphenated). This affects:
- Go module path and all import statements
- Directory structure (`cmd/agent-box/`)
- Binary name and CLI invocations
- Docker image tags (`agent-box/base:*`)
- 692 text occurrences across 78 files

There are no external users or backward compatibility requirements, allowing a clean break.

## Goals / Non-Goals

**Goals:**
- Align codebase with Git remote naming
- Update all references to use `agentbox` consistently
- Maintain functionality throughout (tests pass, binary builds)

**Non-Goals:**
- Backward compatibility (no dual-name support)
- Preserving historical accuracy in archives (consistency over history)
- Supporting old Docker images (clean rebuild required)

## Decisions

### Decision 1: Clean Break vs Gradual Migration

**Chosen**: Clean break - change everything at once in a single commit.

**Rationale**:
- No external users to disrupt
- Dual-name support adds complexity with no benefit
- Single atomic change is easier to review and reason about

**Alternatives considered**:
- Gradual migration with both names supported: Rejected due to unnecessary complexity
- Keeping archives unchanged: Rejected for consistency (user prefers updating archives)

### Decision 2: Archive Handling

**Chosen**: Update all archives to use `agentbox`.

**Rationale**: User preference for consistency across the codebase over historical accuracy.

**Alternatives considered**:
- Leave archives as historical record: Rejected per user preference

### Decision 3: Sequencing of Changes

**Chosen**: Follow this sequence:
1. Rename directory structure first (`cmd/agent-box/` → `cmd/agentbox/`)
2. Update Go module path in `go.mod`
3. Update all import paths in Go files (must happen together with step 2)
4. Update Docker image namespace
5. Update text references (CLI metadata, docs, comments)
6. Run `go mod tidy` and verify build

**Rationale**: Structural changes first, then propagate through code, finally update documentation. This minimizes intermediate broken states during manual application.

**Alternatives considered**:
- Text-first approach: Would leave imports broken temporarily
- Random order: Harder to track progress and verify completeness

## Risks / Trade-offs

**[Risk]** Missing occurrences during bulk replacement
**→ Mitigation**: Use grep to verify zero remaining `agent-box` occurrences before commit. Search patterns: `agent-box`, `github.com/gbrindisi/agent-box`.

**[Risk]** Docker image cache invalidation surprises users
**→ Mitigation**: Expected behavior, no external users. User is aware images will rebuild.

**[Trade-off]** Historical archives will reference `agentbox` even though it was `agent-box` at that time
**→ Accepted**: Consistency over historical accuracy per user preference.

**[Risk]** Go import paths might be overlooked in less obvious places
**→ Mitigation**: After updating, run `go build` and `go test ./...` to catch any missed imports via compilation errors.

## Migration Plan

**Deployment**:
1. Apply all changes in single commit
2. Rebuild binary: `go build -o agentbox ./cmd/agentbox`
3. Reinstall: `sudo mv agentbox /usr/local/bin/`
4. Docker images rebuild automatically on next run

**Rollback**:
Not applicable - this is the source repository. If needed, use `git revert`.

**Verification**:
- `go build` succeeds
- `go test ./...` passes
- `grep -r "agent-box" .` (excluding .git) returns only false positives
- Binary runs: `agentbox --version`
