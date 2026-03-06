## Why

When running multiple agentbox instances on different workspaces with identical build scripts, they incorrectly share the same derived image. This causes issues when agents have workspace-specific state or expectations. Each workspace should have its own dedicated image to ensure isolation.

## What Changes

- Derived image tags will include both the build script hash AND the workspace path
- Running agentbox in a different workspace with the same build script will trigger a rebuild
- Image tag format changes from `agentbox/build:<script-hash>` to `agentbox/build:<combined-hash>`

## Capabilities

### New Capabilities

None - this modifies existing behavior.

### Modified Capabilities

- `build-script`: Hash-based image caching will include workspace path in the hash computation, not just the build script content.

## Impact

- `internal/container/builder.go` - `DerivedImageTag()` signature changes to accept workspace path
- `internal/container/builder.go` - `EnsureDerivedImage()` passes workspace path through
- `cmd/agentbox/cmd/common.go` - Passes workspace path to image building
- Existing cached images will not be reused (new hash includes workspace path)
- More disk space used (one image per workspace instead of shared)
