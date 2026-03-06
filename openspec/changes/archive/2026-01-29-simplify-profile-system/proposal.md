## Why

The current profile system is confusing and misleading. Profiles define an `Image` field (e.g., `ghcr.io/anthropics/claude-code:latest`) that is configured but never used at runtime - containers always use the hardcoded `agentbox/base` image. This creates a disconnect between configuration and behavior. Additionally, there's no mechanism to install agent tools (like claude-code CLI) into the base image.

## What Changes

- **BREAKING**: Remove `Image` field from profiles and AgentConfig - all containers use `agentbox/base`
- **BREAKING**: Remove `aider` profile - focus on `claude-code` only for now
- **BREAKING**: Remove `profile` field from Agentfile - profiles are now init-time templates only
- Add `build_script` field to AgentConfig for installing tools at image build time
- Change profile purpose: profiles become templates that generate explicit Agentfile content during `init`
- Add derived image building: when `build_script` is present, build a derived image (`agentbox/build:<hash>`) from the script
- Add `--rebuild` flag to force image rebuild

## Capabilities

### New Capabilities

- `build-script`: Build-time script execution for installing agent tools into the container image

### Modified Capabilities

- `config`: Remove `Image` field, remove `profile` field, add `build_script` field
- `container`: Build derived images from `build_script`, use hash-based image tagging for caching

## Impact

- **Breaking changes**: Existing Agentfiles using `profile:` or `image:` fields will need updating
- **CLI**: `init` command generates fully explicit Agentfile (no profile reference)
- **Config**: AgentConfig struct loses `Image` and `Profile` fields, gains `BuildScript`
- **Container**: Runner uses derived image tag instead of hardcoded `ImageTag()`
- **Profiles**: Reduced to template data for `init` command only
