## Context

The current system has a profile abstraction that sets `Image`, `Command`, `Args`, `EnvPassthrough`, and `Network`. However, `CreateContainer()` in `runner.go` ignores `cfg.Agent.Image` and always uses `ImageTag()` which returns `agentbox/base:0.1.0`. This means:

1. Profile's `Image` field is dead code
2. Users cannot actually use different base images
3. There's no way to install tools like `claude` CLI into the container

The base image (`agentbox/base`) is a Debian image with firewall tools but no agent software. We need a mechanism to install agent tools at build time.

## Goals / Non-Goals

**Goals:**
- Make configuration explicit - what you see in Agentfile is what runs
- Provide a mechanism to install agent tools (claude-code, etc.) at image build time
- Keep fast startup via image caching
- Maintain security model (firewall, privilege dropping)

**Non-Goals:**
- Supporting arbitrary base images (always use `agentbox/base`)
- Runtime tool installation (too slow, network restrictions)
- Supporting multiple profiles beyond `claude-code` (can add later)

## Decisions

### Decision 1: Profiles become init-time templates only

**Choice**: Profiles are used only by `agentbox init` to generate an explicit Agentfile. No runtime profile resolution.

**Rationale**:
- Eliminates hidden behavior - Agentfile contains all configuration
- Simpler mental model - no need to understand profile inheritance
- Easier debugging - what you see is what runs

**Alternatives considered**:
- Keep runtime profiles with explicit overrides - rejected because it maintains the confusion
- Remove profiles entirely - rejected because init templates are valuable for onboarding

### Decision 2: Build script runs as root during image build

**Choice**: The `build_script` runs as root user during `docker build`, then the runtime still runs as `agent` user.

**Rationale**:
- Some tools need root for installation (apt, system-wide paths)
- Build-time root is safe - it's isolated and reproducible
- Runtime remains unprivileged (security preserved)

**Alternatives considered**:
- Run as agent user - rejected because many install scripts need root
- Provide both options - rejected as unnecessary complexity

### Decision 3: Hash-based image tagging for caching

**Choice**: Derived images tagged as `agentbox/build:<sha256(build_script)[:12]>`

**Rationale**:
- Automatic cache invalidation when script changes
- No manual version management needed
- Deterministic - same script = same hash = reuse

**Alternatives considered**:
- User-specified tags - rejected because users would forget to update
- Timestamp-based - rejected because it defeats caching
- Content-hash of entire Agentfile - rejected because non-script changes shouldn't rebuild

### Decision 4: Dockerfile generation for derived images

**Choice**: Generate a Dockerfile dynamically:
```dockerfile
FROM agentbox/base:0.1.0
USER root
RUN <<'SCRIPT'
<build_script content>
SCRIPT
USER agent
```

**Rationale**:
- Leverages Docker's build cache
- Clean layer separation (base + tools)
- Script runs in shell context with heredoc for multi-line support

**Alternatives considered**:
- Commit running container - rejected because it's slower and less reproducible
- Volume-mount script at runtime - rejected because it runs every startup

### Decision 5: Remove aider profile, keep only claude-code

**Choice**: Ship with only `claude-code` profile for now.

**Rationale**:
- Focus on one well-tested path
- Reduces maintenance burden
- Users can still create custom Agentfiles for other tools

## Risks / Trade-offs

**Risk: Build script network access**
- Build runs with full network access (Docker default)
- Mitigation: This is acceptable - build is a one-time trusted operation, runtime is still locked down

**Risk: Build script can do anything as root**
- User-provided scripts run as root during build
- Mitigation: This is the user's own machine and script - same trust model as running `curl | bash` locally

**Risk: Breaking change for existing Agentfiles**
- Agentfiles using `profile:` or `image:` will break
- Mitigation: Clear error messages pointing to migration, document upgrade path

**Trade-off: No arbitrary base images**
- Users cannot use their own base images
- Accepted: The security model depends on `agentbox/base` having the firewall infrastructure. Custom base images would bypass this.

**Trade-off: First run is slower**
- First `agentbox run` with build_script needs to build the derived image
- Accepted: Subsequent runs use cached image. Can add progress output to set expectations.
