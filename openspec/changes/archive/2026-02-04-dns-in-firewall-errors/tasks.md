## 1. Add DNS Interception Infrastructure

- [x] 1.1 Add thread-local storage structure for DNS cache (hostname, IP, timestamp)
- [x] 1.2 Add function pointer declarations for real getaddrinfo() and gethostbyname()
- [x] 1.3 Add helper function to store DNS lookup in thread-local cache
- [x] 1.4 Add helper function to lookup hostname by IP from thread-local cache

## 2. Implement getaddrinfo() Interception

- [x] 2.1 Implement getaddrinfo() wrapper function
- [x] 2.2 Initialize real_getaddrinfo function pointer using dlsym()
- [x] 2.3 Call real getaddrinfo() and check return value
- [x] 2.4 Extract IPv4 address from result if successful
- [x] 2.5 Cache hostname and IPv4 address with current timestamp
- [x] 2.6 Return result from real getaddrinfo()

## 3. Implement gethostbyname() Interception

- [x] 3.1 Implement gethostbyname() wrapper function
- [x] 3.2 Initialize real_gethostbyname function pointer using dlsym()
- [x] 3.3 Call real gethostbyname() and check return value
- [x] 3.4 Extract IPv4 address from h_addr_list if successful
- [x] 3.5 Cache hostname and IPv4 address with current timestamp
- [x] 3.6 Return result from real gethostbyname()

## 4. Update connect() Error Messages

- [x] 4.1 In connect() function, lookup hostname by IP before blocking
- [x] 4.2 Check if cached entry exists and is within 2-second TTL
- [x] 4.3 Format error message with hostname and IP if hostname available
- [x] 4.4 Format error message with IP only if hostname not available
- [x] 4.5 Update error message format to match spec requirements

## 5. Testing

- [x] 5.1 Add test case for DNS lookup followed by blocked connection with hostname display
- [x] 5.2 Add test case for blocked connection without prior DNS lookup (IP only)
- [x] 5.3 Add test case for cache TTL expiration (2+ seconds between DNS and connect)
- [x] 5.4 Add test case for multiple threads performing independent DNS lookups
- [x] 5.5 Verify existing tests still pass with new interception code
