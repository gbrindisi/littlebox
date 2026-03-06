#define _GNU_SOURCE
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <dlfcn.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <errno.h>
#include <pthread.h>
#include <stdint.h>
#include <time.h>
#include <netdb.h>
#include <sys/file.h>
#include <fcntl.h>
#include <unistd.h>

// Function pointer to the real connect() syscall
static int (*real_connect)(int, const struct sockaddr *, socklen_t) = NULL;

// Function pointers for DNS resolution functions
static int (*real_getaddrinfo)(const char *, const char *, const struct addrinfo *, struct addrinfo **) = NULL;
static struct hostent *(*real_gethostbyname)(const char *) = NULL;

// Process-wide DNS cache structure (shared across threads)
typedef struct {
    char hostname[256];
    uint32_t ip;
    time_t timestamp;
} dns_cache_entry_t;

#define DNS_CACHE_SIZE 64
#define DNS_CACHE_TTL_SECONDS 2
#define DNS_CACHE_FILE "/run/sandbox/dns_cache"
#define DNS_CACHE_FILE_MAX_ENTRIES 256

static dns_cache_entry_t dns_cache[DNS_CACHE_SIZE];
static int dns_cache_next = 0;
static pthread_mutex_t dns_cache_mutex = PTHREAD_MUTEX_INITIALIZER;
static __thread char dns_cache_hostname_buf[256];

// Initialization flag
static pthread_once_t init_once = PTHREAD_ONCE_INIT;
static int initialized = 0;

// Allowed IP list structure
typedef struct {
    uint32_t network;  // Network address in host byte order
    uint32_t mask;     // Network mask in host byte order
} cidr_entry_t;

#define MAX_ALLOWED_IPS 1024
static cidr_entry_t allowed_ips[MAX_ALLOWED_IPS];
static int num_allowed_ips = 0;

// Initialize the real connect function pointer
static void init_real_connect(void) {
    real_connect = dlsym(RTLD_NEXT, "connect");
    if (!real_connect) {
        fprintf(stderr, "agentbox: Failed to load real connect(): %s\n", dlerror());
        exit(1);
    }
}

// Initialize the real getaddrinfo function pointer
static void init_real_getaddrinfo(void) {
    real_getaddrinfo = dlsym(RTLD_NEXT, "getaddrinfo");
    if (!real_getaddrinfo) {
        fprintf(stderr, "agentbox: Failed to load real getaddrinfo(): %s\n", dlerror());
        exit(1);
    }
}

// Initialize the real gethostbyname function pointer
static void init_real_gethostbyname(void) {
    real_gethostbyname = dlsym(RTLD_NEXT, "gethostbyname");
    if (!real_gethostbyname) {
        fprintf(stderr, "agentbox: Failed to load real gethostbyname(): %s\n", dlerror());
        exit(1);
    }
}

// Parse a CIDR notation IP (e.g., "192.168.1.0/24" or "10.0.0.1")
static int parse_cidr(const char *cidr_str, cidr_entry_t *entry) {
    char buf[256];
    char *slash;
    uint32_t ip_addr;
    int prefix_len = 32;  // Default to /32 for single IPs
    struct in_addr addr;

    strncpy(buf, cidr_str, sizeof(buf) - 1);
    buf[sizeof(buf) - 1] = '\0';

    slash = strchr(buf, '/');
    if (slash) {
        *slash = '\0';
        prefix_len = atoi(slash + 1);
        if (prefix_len < 0 || prefix_len > 32) {
            return -1;
        }
    }

    if (inet_pton(AF_INET, buf, &addr) != 1) {
        return -1;
    }

    ip_addr = ntohl(addr.s_addr);

    // Calculate network mask
    if (prefix_len == 0) {
        entry->mask = 0;
    } else {
        entry->mask = ~((1U << (32 - prefix_len)) - 1);
    }

    entry->network = ip_addr & entry->mask;

    return 0;
}

// Load allowed IPs from file
static void load_allowed_ips(void) {
    FILE *fp;
    char line[256];
    cidr_entry_t entry;

    fp = fopen("/run/sandbox/allowed_ips", "r");
    if (!fp) {
        // Fail-open for compatibility
        fprintf(stderr, "agentbox: Warning - /run/sandbox/allowed_ips not found, allowing all connections\n");
        initialized = -1;  // Special flag: allow all
        return;
    }

    num_allowed_ips = 0;
    while (fgets(line, sizeof(line), fp) && num_allowed_ips < MAX_ALLOWED_IPS) {
        // Remove trailing newline
        line[strcspn(line, "\n")] = '\0';

        // Skip empty lines and comments
        if (line[0] == '\0' || line[0] == '#') {
            continue;
        }

        if (parse_cidr(line, &entry) == 0) {
            allowed_ips[num_allowed_ips++] = entry;
        }
    }

    fclose(fp);
    initialized = 1;
}

// Initialize the library (called once via pthread_once)
static void init_library(void) {
    init_real_connect();
    init_real_getaddrinfo();
    init_real_gethostbyname();
    load_allowed_ips();
}

// Check if an IP is in the allowed list
static int is_ip_allowed(uint32_t ip_addr) {
    int i;

    // If we failed to open the file, allow all connections
    if (initialized == -1) {
        return 1;
    }

    for (i = 0; i < num_allowed_ips; i++) {
        if ((ip_addr & allowed_ips[i].mask) == allowed_ips[i].network) {
            return 1;
        }
    }

    return 0;
}

// Convert IP to string
static int ip_to_string(uint32_t ip, char *buf, size_t size) {
    struct in_addr addr;
    addr.s_addr = htonl(ip);
    return inet_ntop(AF_INET, &addr, buf, size) != NULL;
}

// Convert IP string to uint32
static int ip_from_string(const char *buf, uint32_t *ip_out) {
    struct in_addr addr;
    if (inet_pton(AF_INET, buf, &addr) != 1) {
        return 0;
    }
    *ip_out = ntohl(addr.s_addr);
    return 1;
}

// Store DNS lookup in process-wide cache
static void cache_dns_lookup_memory(const char *hostname, uint32_t ip) {
    if (!hostname || hostname[0] == '\0') {
        return;
    }

    pthread_mutex_lock(&dns_cache_mutex);
    dns_cache_entry_t *entry = &dns_cache[dns_cache_next];
    dns_cache_next = (dns_cache_next + 1) % DNS_CACHE_SIZE;

    strncpy(entry->hostname, hostname, sizeof(entry->hostname) - 1);
    entry->hostname[sizeof(entry->hostname) - 1] = '\0';
    entry->ip = ip;
    entry->timestamp = time(NULL);
    pthread_mutex_unlock(&dns_cache_mutex);
}

static void cache_file_entries_add(dns_cache_entry_t *entries, int *count, const char *hostname, uint32_t ip, time_t timestamp) {
    if (!hostname || hostname[0] == '\0') {
        return;
    }

    if (*count >= DNS_CACHE_FILE_MAX_ENTRIES) {
        memmove(entries, entries + 1, sizeof(dns_cache_entry_t) * (DNS_CACHE_FILE_MAX_ENTRIES - 1));
        *count = DNS_CACHE_FILE_MAX_ENTRIES - 1;
    }

    dns_cache_entry_t *entry = &entries[*count];
    strncpy(entry->hostname, hostname, sizeof(entry->hostname) - 1);
    entry->hostname[sizeof(entry->hostname) - 1] = '\0';
    entry->ip = ip;
    entry->timestamp = timestamp;
    (*count)++;
}

// Store DNS lookup in shared cache file
static void cache_dns_lookup_file(const char *hostname, uint32_t ip) {
    int fd;
    FILE *fp;
    char line[512];
    dns_cache_entry_t entries[DNS_CACHE_FILE_MAX_ENTRIES];
    int count = 0;
    time_t now = time(NULL);

    if (!hostname || hostname[0] == '\0') {
        return;
    }

    fd = open(DNS_CACHE_FILE, O_RDWR | O_CREAT, 0644);
    if (fd < 0) {
        return;
    }

    if (flock(fd, LOCK_EX) != 0) {
        close(fd);
        return;
    }

    fp = fdopen(fd, "r+");
    if (!fp) {
        flock(fd, LOCK_UN);
        close(fd);
        return;
    }

    while (fgets(line, sizeof(line), fp)) {
        long ts;
        char ip_str[64];
        char host[256];
        uint32_t parsed_ip;
        if (sscanf(line, "%ld %63s %255s", &ts, ip_str, host) != 3) {
            continue;
        }
        if (ts <= 0) {
            continue;
        }
        time_t age = now - (time_t)ts;
        if (age < 0 || age > DNS_CACHE_TTL_SECONDS) {
            continue;
        }
        if (!ip_from_string(ip_str, &parsed_ip)) {
            continue;
        }
        cache_file_entries_add(entries, &count, host, parsed_ip, (time_t)ts);
    }

    cache_file_entries_add(entries, &count, hostname, ip, now);

    rewind(fp);
    if (ftruncate(fd, 0) != 0) {
        // Ignore truncation errors
    }
    rewind(fp);

    for (int i = 0; i < count; i++) {
        char ip_buf[INET_ADDRSTRLEN];
        if (!ip_to_string(entries[i].ip, ip_buf, sizeof(ip_buf))) {
            continue;
        }
        fprintf(fp, "%ld %s %s\n", (long)entries[i].timestamp, ip_buf, entries[i].hostname);
    }
    fflush(fp);

    flock(fd, LOCK_UN);
    fclose(fp);
}

// Store DNS lookup in both memory and shared cache
static void cache_dns_lookup(const char *hostname, uint32_t ip) {
    cache_dns_lookup_memory(hostname, ip);
    cache_dns_lookup_file(hostname, ip);
}

// Lookup hostname by IP from process-wide cache (with TTL)
static const char *lookup_hostname_by_ip_memory(uint32_t ip) {
    time_t now = time(NULL);
    time_t best_ts = 0;
    const char *result = NULL;

    pthread_mutex_lock(&dns_cache_mutex);
    for (int i = 0; i < DNS_CACHE_SIZE; i++) {
        dns_cache_entry_t *entry = &dns_cache[i];
        if (entry->hostname[0] == '\0') {
            continue;
        }

        time_t age = now - entry->timestamp;
        if (age >= 0 && age <= DNS_CACHE_TTL_SECONDS && entry->ip == ip && entry->timestamp >= best_ts) {
            best_ts = entry->timestamp;
            strncpy(dns_cache_hostname_buf, entry->hostname, sizeof(dns_cache_hostname_buf) - 1);
            dns_cache_hostname_buf[sizeof(dns_cache_hostname_buf) - 1] = '\0';
            result = dns_cache_hostname_buf;
        }
    }
    pthread_mutex_unlock(&dns_cache_mutex);

    return result;
}

// Lookup hostname by IP from shared cache file (with TTL)
static const char *lookup_hostname_by_ip_file(uint32_t ip) {
    int fd;
    FILE *fp;
    char line[512];
    time_t now = time(NULL);
    time_t best_ts = 0;
    const char *result = NULL;

    fd = open(DNS_CACHE_FILE, O_RDONLY);
    if (fd < 0) {
        return NULL;
    }

    if (flock(fd, LOCK_SH) != 0) {
        close(fd);
        return NULL;
    }

    fp = fdopen(fd, "r");
    if (!fp) {
        flock(fd, LOCK_UN);
        close(fd);
        return NULL;
    }

    while (fgets(line, sizeof(line), fp)) {
        long ts;
        char ip_str[64];
        char host[256];
        uint32_t parsed_ip;
        if (sscanf(line, "%ld %63s %255s", &ts, ip_str, host) != 3) {
            continue;
        }
        if (ts <= 0) {
            continue;
        }
        time_t age = now - (time_t)ts;
        if (age < 0 || age > DNS_CACHE_TTL_SECONDS) {
            continue;
        }
        if (!ip_from_string(ip_str, &parsed_ip)) {
            continue;
        }
        if (parsed_ip != ip) {
            continue;
        }
        if ((time_t)ts >= best_ts) {
            best_ts = (time_t)ts;
            strncpy(dns_cache_hostname_buf, host, sizeof(dns_cache_hostname_buf) - 1);
            dns_cache_hostname_buf[sizeof(dns_cache_hostname_buf) - 1] = '\0';
            result = dns_cache_hostname_buf;
        }
    }

    flock(fd, LOCK_UN);
    fclose(fp);

    return result;
}

// Lookup hostname by IP from memory, then shared file
static const char *lookup_hostname_by_ip(uint32_t ip) {
    const char *hostname = lookup_hostname_by_ip_memory(ip);
    if (hostname) {
        return hostname;
    }

    hostname = lookup_hostname_by_ip_file(ip);
    if (hostname) {
        cache_dns_lookup_memory(hostname, ip);
    }

    return hostname;
}

// Intercepted connect() function
int connect(int sockfd, const struct sockaddr *addr, socklen_t addrlen) {
    struct sockaddr_in *addr_in;
    uint32_t ip_addr;
    char ip_str[INET_ADDRSTRLEN];

    // Lazy initialization
    pthread_once(&init_once, init_library);

    // Pass through if not initialized properly
    if (!real_connect) {
        errno = ENOSYS;
        return -1;
    }

    // Only check TCP/IPv4 connections
    if (addr->sa_family != AF_INET) {
        return real_connect(sockfd, addr, addrlen);
    }

    addr_in = (struct sockaddr_in *)addr;
    ip_addr = ntohl(addr_in->sin_addr.s_addr);

    // Check if IP is allowed
    if (!is_ip_allowed(ip_addr)) {
        const char *hostname;

        // Format IP for error message
        inet_ntop(AF_INET, &addr_in->sin_addr, ip_str, sizeof(ip_str));

        // Lookup hostname by IP from cache
        hostname = lookup_hostname_by_ip(ip_addr);

        // Format error message based on whether hostname is available
        if (hostname) {
            fprintf(stderr, "agentbox: Connection to %s (%s) blocked by sandbox firewall. This is not bypassable.\n", hostname, ip_str);
        } else {
            fprintf(stderr, "agentbox: Connection to %s blocked by sandbox firewall. This is not bypassable.\n", ip_str);
        }

        errno = ECONNREFUSED;
        return -1;
    }

    // Pass through to real connect
    return real_connect(sockfd, addr, addrlen);
}

// Intercepted getaddrinfo() function
int getaddrinfo(const char *node, const char *service, const struct addrinfo *hints, struct addrinfo **res) {
    int ret;
    struct addrinfo *result;

    // Lazy initialization
    pthread_once(&init_once, init_library);

    // Pass through if not initialized properly
    if (!real_getaddrinfo) {
        return EAI_SYSTEM;
    }

    // Call the real getaddrinfo()
    ret = real_getaddrinfo(node, service, hints, res);

    // If successful, extract IPv4 addresses and cache them
    if (ret == 0 && node != NULL && res != NULL) {
        for (result = *res; result != NULL; result = result->ai_next) {
            // Only cache IPv4 addresses
            if (result->ai_family == AF_INET && result->ai_addr != NULL) {
                struct sockaddr_in *addr_in = (struct sockaddr_in *)result->ai_addr;
                uint32_t ip_addr = ntohl(addr_in->sin_addr.s_addr);
                cache_dns_lookup(node, ip_addr);
            }
        }
    }

    return ret;
}

// Intercepted gethostbyname() function
struct hostent *gethostbyname(const char *name) {
    struct hostent *result;

    // Lazy initialization
    pthread_once(&init_once, init_library);

    // Pass through if not initialized properly
    if (!real_gethostbyname) {
        return NULL;
    }

    // Call the real gethostbyname()
    result = real_gethostbyname(name);

    // If successful and result is IPv4, cache all returned addresses
    if (result != NULL && result->h_addrtype == AF_INET && name != NULL) {
        for (char **addr = result->h_addr_list; addr != NULL && *addr != NULL; addr++) {
            uint32_t ip_addr = ntohl(*(uint32_t *)(*addr));
            cache_dns_lookup(name, ip_addr);
        }
    }

    return result;
}
