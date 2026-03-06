## Why

When the sandbox firewall blocks an outbound connection, libsandbox.so currently only displays the IP address in the error message. This makes debugging difficult because users must manually determine which domain was blocked before they can update the firewall allowlist.

## What Changes

- Intercept DNS resolution functions (getaddrinfo, gethostbyname) in libsandbox.so
- Cache hostname-to-IP mappings with a short TTL (1-2 seconds)
- Enhance firewall error messages to display domain names alongside IP addresses when available
- Fall back to IP-only display when hostname information is unavailable

## Capabilities

### New Capabilities
- `dns-caching`: Track DNS lookups to correlate hostnames with IP addresses at connection time

### Modified Capabilities
- `sandbox-connect`: Update blocked connection error message format to include domain names

## Impact

- Modified: `internal/container/docker/libsandbox.c` - Add DNS interception and caching logic
- Modified: `openspec/specs/sandbox-connect/spec.md` - Update error message requirements to include domain names
