## Context

libsandbox.so currently intercepts connect() syscalls using LD_PRELOAD to enforce sandbox firewall rules. When a connection is blocked, it prints the destination IP address. However, by the time connect() is called, DNS resolution has already occurred, so the original hostname is lost. This makes it difficult for users to identify which domain to add to the firewall allowlist.

The DNS resolution typically happens milliseconds before the connect() call in the same thread, providing a narrow window where we can correlate hostnames with IPs.

## Goals / Non-Goals

**Goals:**
- Display domain names in firewall error messages when available
- Maintain thread-safe operation for multi-threaded programs
- Gracefully fall back to IP-only display when hostname unavailable
- Keep implementation simple and low-overhead

**Non-Goals:**
- Guaranteed hostname availability for all connections (best-effort only)
- Support for IPv6 (future enhancement)
- Long-term DNS caching or network-layer DNS interception
- Reverse DNS lookups (unreliable and adds latency)

## Decisions

### Decision 1: Intercept getaddrinfo() and gethostbyname()

**Rationale:** These are the two primary DNS resolution functions used by programs. getaddrinfo() is modern POSIX, gethostbyname() covers legacy code.

**Alternatives considered:**
- Reverse DNS lookup at block time: Unreliable, adds latency, often returns generic PTR records
- Kernel-level tracking (eBPF): Too complex, requires root privileges
- HTTP proxy interception: Only works for HTTP traffic, requires TLS interception

**Trade-off:** We won't catch programs that use alternative DNS methods or internal caching, but we'll cover 90%+ of cases.

### Decision 2: Thread-local storage with 1-2 second TTL

**Rationale:** DNS lookup and connect() typically occur within milliseconds in the same thread. Thread-local storage avoids locking overhead and race conditions. A 1-2 second TTL handles edge cases where programs have logic between DNS and connect.

**Alternatives considered:**
- Global cache with mutex: Adds locking overhead, more complex
- No TTL: Slightly simpler, but risks stale associations if thread does multiple DNS lookups
- Longer TTL (5+ seconds): Higher chance of incorrect hostname associations

**Trade-off:** Thread-local means we only track one recent lookup per thread. If a thread looks up multiple hosts rapidly, we'll only remember the most recent one. This is acceptable given the typical DNS→connect pattern.

### Decision 3: Store single hostname per thread

**Rationale:** Most programs follow a pattern of: lookup hostname → connect to result. Storing only the most recent lookup keeps the implementation simple and avoids memory management complexity.

**Alternatives considered:**
- Hash map of IP→hostname: More accurate but requires dynamic allocation, locking, and cleanup
- Fixed-size circular buffer: More complex, minimal benefit

**Trade-off:** If a thread looks up multiple hosts before connecting, we'll lose earlier entries. Acceptable for typical usage patterns.

## Risks / Trade-offs

**[Risk]** Programs that cache DNS results internally will bypass our interception
→ **Mitigation:** Fall back to IP-only display, which is current behavior

**[Risk]** Multi-threaded programs with complex DNS patterns may show incorrect hostnames
→ **Mitigation:** 1-2 second TTL reduces stale entries, thread-local storage prevents cross-thread contamination

**[Risk]** Memory usage for storing hostnames in thread-local storage
→ **Mitigation:** Fixed 256-byte buffer per thread, deallocated when thread exits

**[Trade-off]** Additional function interception adds minimal overhead to DNS resolution
→ **Acceptable:** DNS is already I/O-bound, the cache write is negligible

**[Trade-off]** IPv6 connections will not show hostnames
→ **Acceptable:** Can be added in future enhancement if needed
