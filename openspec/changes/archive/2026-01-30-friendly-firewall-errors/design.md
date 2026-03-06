## Context

When agentbox blocks network requests via iptables, agents see generic errors like "Connection refused" or "Connection timed out". Terminal agents (e.g., Claude Code in dangerously-skip-permissions mode) don't understand why requests fail and may waste cycles attempting workarounds.

Current flow:
1. Agent runs `curl https://blocked-site.com`
2. DNS resolves (allowed)
3. TCP connect to IP hits iptables DROP rule
4. Agent sees "Connection timed out" - no context

Network protocols (TCP, ICMP) don't support custom error messages. iptables REJECT provides faster failure but still only generic ICMP codes.

## Goals / Non-Goals

**Goals:**
- Provide human-readable error messages when firewall blocks connections
- Message should explain the sandbox restriction and discourage bypass attempts
- Work for all TCP-based protocols (HTTP, HTTPS, SSH, etc.)
- Maintain iptables as the hard security enforcement layer

**Non-Goals:**
- MITM proxy for HTTPS inspection (too complex, security concerns)
- Custom messages for UDP traffic (less common for agents)
- Preventing all possible bypasses of the friendly message (iptables remains the security layer)

## Decisions

### Decision 1: LD_PRELOAD interception over transparent proxy

**Choice:** Use LD_PRELOAD to intercept `connect()` syscalls in userspace.

**Alternatives considered:**
- Transparent proxy: Requires running a proxy process, HTTPS is problematic without MITM
- NFQUEUE: Overkill, still can't inject custom text into network protocols
- DNS-level blocking: HTTPS cert errors obscure the message

**Rationale:** LD_PRELOAD allows printing custom messages to stderr before returning the connection error. Works for all protocols. Simple C library with minimal dependencies.

### Decision 2: File-based allowed IP list sync

**Choice:** Write allowed IPs to `/run/sandbox/allowed_ips` during firewall init, read once by library.

**Alternatives considered:**
- Environment variable: Size limits for many CIDRs
- Query ipset at runtime: Requires elevated permissions
- Shared memory: Overkill for static data

**Rationale:** IPs are determined once at startup and don't change. A file is simple, readable by agent user, and can be cached in memory for O(1) lookups.

### Decision 3: Defense in depth - keep iptables as primary enforcement

**Choice:** LD_PRELOAD provides UX (friendly messages), iptables provides security (hard block).

**Rationale:** LD_PRELOAD can be bypassed (static binaries, direct syscalls, Go runtime). The iptables firewall catches anything that bypasses the library. Agents that bypass LD_PRELOAD still get blocked, just with a generic error.

### Decision 4: Library loads allowed list lazily on first connect

**Choice:** Parse `/run/sandbox/allowed_ips` on first `connect()` call, cache in memory.

**Rationale:** Avoids startup overhead for processes that never make network calls. Single parse, then O(1) hash lookups.

## Risks / Trade-offs

**[Risk] Go binaries bypass LD_PRELOAD (uses direct syscalls)**
- Mitigation: iptables still blocks. Go-based agents get generic errors but are still sandboxed.

**[Risk] Statically linked binaries bypass LD_PRELOAD**
- Mitigation: Same as above - iptables catches them.

**[Risk] Allowed IP list could drift from ipset**
- Mitigation: Both are generated from the same source (ALLOWED_DOMAINS) in the same script. File is written immediately after ipset is populated.

**[Risk] Library adds attack surface**
- Mitigation: Minimal C code, no external dependencies, only intercepts connect(). Library runs as unprivileged agent user.

**[Trade-off] Message visible on stderr, not in HTTP response body**
- For terminal agents reading command output, stderr is visible and useful. For programmatic HTTP clients, they'd need to capture stderr.
