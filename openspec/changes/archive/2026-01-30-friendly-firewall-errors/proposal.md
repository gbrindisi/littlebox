## Why

When the agentbox firewall blocks a network request, the agent sees generic errors like "Connection refused" or "Connection timed out" with no context. Terminal agents (like Claude Code in yolo mode) don't understand why the request failed and may waste cycles attempting workarounds. A human-readable error message explaining the sandbox restriction would help agents understand the constraint immediately.

## What Changes

- Add an LD_PRELOAD shared library (`libsandbox.so`) that intercepts `connect()` syscalls
- When a connection to a blocked IP is attempted, print a clear error message to stderr before returning the error
- Export the allowed IP list from `init-firewall.sh` to a file readable by the library
- Keep iptables firewall as the hard enforcement layer (defense in depth)

## Capabilities

### New Capabilities

- `sandbox-connect`: Intercepts outbound connections via LD_PRELOAD to provide human-readable error messages when the firewall blocks a request

### Modified Capabilities

- `security`: Add requirement for exporting allowed IPs to a file for the LD_PRELOAD library to read

## Impact

- New C shared library (`libsandbox.so`) to build and include in container image
- Dockerfile changes to compile the library and set LD_PRELOAD
- `init-firewall.sh` changes to export allowed IPs to `/run/sandbox/allowed_ips`
- Entrypoint changes to set LD_PRELOAD environment variable before dropping privileges
