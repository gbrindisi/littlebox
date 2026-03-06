## ADDED Requirements

### Requirement: DNS function interception
The dns-caching module SHALL intercept DNS resolution functions via LD_PRELOAD.

#### Scenario: getaddrinfo interception
- **WHEN** a process calls getaddrinfo() to resolve a hostname
- **THEN** libsandbox.so intercepts the call before passing to the real function

#### Scenario: gethostbyname interception
- **WHEN** a process calls gethostbyname() to resolve a hostname
- **THEN** libsandbox.so intercepts the call before passing to the real function

### Requirement: Hostname-to-IP caching
The dns-caching module SHALL cache hostname-to-IP mappings for recently resolved domains.

#### Scenario: Cache successful IPv4 resolution
- **WHEN** getaddrinfo() or gethostbyname() successfully resolves a hostname to an IPv4 address
- **THEN** libsandbox.so stores the hostname and IP address in thread-local storage
- **AND** the cache entry includes a timestamp

#### Scenario: Non-IPv4 resolutions pass through
- **WHEN** getaddrinfo() returns a non-IPv4 result (e.g., IPv6, Unix socket)
- **THEN** libsandbox.so does not cache the result
- **AND** the function proceeds normally

#### Scenario: Failed DNS lookups not cached
- **WHEN** getaddrinfo() or gethostbyname() fails to resolve a hostname
- **THEN** libsandbox.so does not update the cache
- **AND** the function returns the error normally

### Requirement: Cache TTL management
The dns-caching module SHALL expire cache entries after a short TTL.

#### Scenario: Cache entry within TTL
- **WHEN** a cached entry is less than 2 seconds old
- **THEN** the entry is considered valid for hostname lookup

#### Scenario: Cache entry expired
- **WHEN** a cached entry is 2 seconds or older
- **THEN** the entry is considered expired and not used for hostname lookup

### Requirement: Thread-local cache storage
The dns-caching module SHALL maintain separate caches per thread.

#### Scenario: Thread-isolated cache
- **WHEN** different threads perform DNS lookups
- **THEN** each thread maintains its own independent cache entry
- **AND** cache entries do not interfere across threads

#### Scenario: Single entry per thread
- **WHEN** a thread performs multiple DNS lookups
- **THEN** only the most recent lookup is stored in the cache
- **AND** previous entries for that thread are overwritten

### Requirement: Hostname lookup by IP
The dns-caching module SHALL provide a mechanism to retrieve the hostname for a given IP address.

#### Scenario: IP found in cache
- **WHEN** an IP address matches a cached entry within the TTL
- **THEN** the cached hostname is returned

#### Scenario: IP not in cache
- **WHEN** an IP address is not found in the cache or the entry is expired
- **THEN** no hostname is returned (null or empty result)
