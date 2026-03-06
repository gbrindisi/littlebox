## Why

The Git remote is already set to `github.com/gbrindisi/agentbox` (no hyphen), but the codebase still references `agent-box` everywhere (692 occurrences across 78 files). This creates inconsistency between the repository identifier and the code. Additionally, `agentbox` as a single word is simpler and more product-like than the hyphenated form.

## What Changes

- **BREAKING**: Rename Go module from `github.com/gbrindisi/agent-box` to `github.com/gbrindisi/agentbox`
- **BREAKING**: Rename directory `cmd/agent-box/` to `cmd/agentbox/`
- **BREAKING**: Rename binary output from `agent-box` to `agentbox`
- **BREAKING**: Update Docker image namespace from `agent-box/base` to `agentbox/base`
- **BREAKING**: Update all CLI commands from `agent-box run` to `agentbox run`
- Update all import paths in Go files to use new module path
- Update GitHub URLs from `/agent-box` to `/agentbox`
- Update all documentation, comments, and user-facing text
- Update error messages in libsandbox.c and shell scripts
- Update OpenSpec specifications and archived changes

## Capabilities

### New Capabilities
<!-- No new capabilities being introduced -->

### Modified Capabilities
<!-- No requirement changes - this is purely a rename -->

## Impact

- **Users**: Must reinstall binary at new path (`/usr/local/bin/agentbox`), all Docker images will rebuild on first run, all commands change from `agent-box` to `agentbox`
- **Code**: All Go import paths change, affecting 28 files with 37 import occurrences
- **Documentation**: README.md, all OpenSpec specs and archives, code comments
- **Build artifacts**: Docker images cached as `agent-box/base:*` will be replaced with `agentbox/base:*`
- **External dependencies**: None (no external users)
