## Context

Currently, derived image tags are computed solely from the SHA256 hash of the build script content. This means two different workspaces with identical Agentfiles share the same cached image. When running multiple agentbox instances simultaneously on different workspaces, this shared image can cause conflicts if the agent has workspace-specific expectations.

## Goals / Non-Goals

**Goals:**
- Each workspace gets its own dedicated derived image
- Automatic rebuild when running in a new workspace
- Maintain caching within the same workspace (same script + same workspace = reuse)

**Non-Goals:**
- Portable images across machines (absolute paths are machine-specific anyway)
- Optimizing disk space usage (we accept more images)
- Image sharing between similar workspaces

## Decisions

### Decision 1: Include absolute workspace path in hash

**Choice**: Hash `buildScript + "\x00" + absoluteWorkspacePath`

**Rationale**:
- Absolute path guarantees uniqueness across all workspaces on the machine
- Null byte separator prevents collision attacks (e.g., script "A" + path "B" vs script "AB")
- Simple to implement - single line change to hash computation

**Alternatives considered**:
- Workspace basename only: Could collide if two directories share the same name
- Workspace hash in label (runtime check): More complex, still need to trigger rebuild
- Configurable behavior: Over-engineering for a clear use case

### Decision 2: Use resolved absolute path

**Choice**: Use the fully resolved workspace path from config (after `ExpandPaths()`)

**Rationale**:
- Config already resolves relative paths to absolute
- Moving a workspace directory correctly triggers a rebuild (it's semantically a different workspace)
- Symlink resolution happens at config level, not in hash

## Risks / Trade-offs

- **More disk usage**: Each workspace has its own image even with identical build scripts
  - Mitigation: Acceptable trade-off for isolation; users can prune old images

- **Rebuild on workspace move**: Moving a directory triggers rebuild
  - Mitigation: This is arguably correct behavior; the workspace identity changed

- **Breaking change for cached images**: Existing images won't be found with new hash scheme
  - Mitigation: Automatic rebuild on first run; no data loss
