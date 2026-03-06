## 1. Library Implementation

- [x] 1.1 Create libsandbox.c with connect() interception via dlsym(RTLD_NEXT)
- [x] 1.2 Implement allowed IP list parsing from /run/sandbox/allowed_ips
- [x] 1.3 Implement CIDR matching for IP ranges
- [x] 1.4 Add stderr message output for blocked connections
- [x] 1.5 Add lazy loading (parse file on first connect, cache in memory)

## 2. Firewall Script Changes

- [x] 2.1 Add /run/sandbox directory creation to init-firewall.sh
- [x] 2.2 Export ipset contents to /run/sandbox/allowed_ips after population
- [x] 2.3 Set file permissions to 444 (world-readable)

## 3. Container Integration

- [x] 3.1 Add libsandbox.c to internal/container/docker/
- [x] 3.2 Update Dockerfile to compile libsandbox.so (gcc -shared -fPIC -ldl)
- [x] 3.3 Update entrypoint.sh to set LD_PRELOAD before exec to agent user

## 4. Testing

- [x] 4.1 Add test verifying blocked connection shows custom error message
- [x] 4.2 Add test verifying allowed connection passes through without message
- [x] 4.3 Add test verifying CIDR range matching works correctly
- [x] 4.4 Verify existing firewall tests still pass
