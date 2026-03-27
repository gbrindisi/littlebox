#!/bin/bash
# littlebox firewall initialization script
# This script sets up iptables rules with ipset for efficient network filtering.
# It runs directly as root from the entrypoint (container starts as root).
#
# Environment variables:
#   ALLOWED_DOMAINS - Comma-separated list of allowed domains
#   DNS_SERVER - Override DNS server (auto-detected if not set)
#
# Features:
#   - Docker DNS auto-detection and preservation
#   - Domain resolution to IP addresses using dig
#   - CIDR range support via ipset hash:net
#   - GitHub IP range fetching from api.github.com/meta
#   - IPv6 firewall rules (default deny)
#   - Firewall verification tests

set -euo pipefail

# Configuration from environment
ALLOWED_DOMAINS="${ALLOWED_DOMAINS:-}"
DNS_SERVER="${DNS_SERVER:-}"
DNS_SERVER_V6="${DNS_SERVER_V6:-}"

log() {
    echo "[firewall] $*"
}

# =============================================================================
# DNS Detection
# =============================================================================

detect_dns() {
    if [ -n "$DNS_SERVER" ]; then
        log "Using configured DNS server: $DNS_SERVER"
        return
    fi

    DNS_SERVER=$(grep nameserver /etc/resolv.conf | head -1 | awk '{print $2}' || true)

    if [[ "$DNS_SERVER" == "127.0.0.11" ]]; then
        log "Docker DNS detected: $DNS_SERVER"
    elif [[ "$DNS_SERVER" =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        log "External DNS detected: $DNS_SERVER"
    else
        log "Warning: Could not detect DNS server, using Docker default"
        DNS_SERVER="127.0.0.11"
    fi
}

# =============================================================================
# ipset Management
# =============================================================================

setup_ipset() {
    log "Creating ipset for allowed IPs..."
    # hash:net supports CIDR ranges
    ipset create allowed_ips hash:net 2>/dev/null || ipset flush allowed_ips

    # Always allow the DNS server
    ipset add allowed_ips "$DNS_SERVER/32" 2>/dev/null || true
}

# =============================================================================
# Domain Resolution
# =============================================================================

resolve_and_add() {
    local domain="$1"
    log "Resolving: $domain"

    # Get all A records using dig
    local ips
    ips=$(dig +short "$domain" A 2>/dev/null | grep -E '^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$' || true)

    if [ -z "$ips" ]; then
        log "  Warning: No IPs found for $domain"
        return
    fi

    for ip in $ips; do
        log "  Adding: $ip"
        ipset add allowed_ips "$ip/32" 2>/dev/null || true
    done
}

add_cidr() {
    local cidr="$1"
    log "Adding CIDR: $cidr"
    ipset add allowed_ips "$cidr" 2>/dev/null || true
}

add_ip() {
    local ip="$1"
    log "Adding IP: $ip"
    ipset add allowed_ips "$ip/32" 2>/dev/null || true
}

# Detect entry type: domain, IP, or CIDR
# Returns: "domain", "ip", or "cidr"
detect_entry_type() {
    local entry="$1"

    # CIDR: contains a slash
    if [[ "$entry" == *"/"* ]]; then
        echo "cidr"
        return
    fi

    # IPv4: matches pattern like "10.0.0.1"
    if [[ "$entry" =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        echo "ip"
        return
    fi

    # Default: domain name
    echo "domain"
}

# =============================================================================
# GitHub IP Range Fetching
# =============================================================================

fetch_github_ips() {
    log "Fetching GitHub IP ranges from api.github.com/meta..."

    local meta
    meta=$(curl -s --max-time 10 https://api.github.com/meta 2>/dev/null || true)

    if [ -z "$meta" ]; then
        log "  Warning: Could not fetch GitHub meta API"
        return
    fi

    # Extract IP ranges for different services (git, web, api, packages)
    local ranges
    ranges=$(echo "$meta" | jq -r '
        .git[]?,
        .web[]?,
        .api[]?,
        .packages[]?
    ' 2>/dev/null | sort -u || true)

    if [ -z "$ranges" ]; then
        log "  Warning: Could not parse GitHub IP ranges"
        return
    fi

    local count=0
    for range in $ranges; do
        ipset add allowed_ips "$range" 2>/dev/null || true
        count=$((count + 1))
    done
    log "  Added $count GitHub IP ranges"
}

# =============================================================================
# iptables Setup (IPv4)
# =============================================================================

setup_iptables() {
    log "Setting up iptables (IPv4) rules..."

    # Flush existing OUTPUT rules
    iptables -F OUTPUT 2>/dev/null || true

    # Set default policy to DROP
    iptables -P OUTPUT DROP

    # Allow loopback
    iptables -A OUTPUT -o lo -j ACCEPT

    # Allow established/related connections
    iptables -A OUTPUT -m state --state ESTABLISHED,RELATED -j ACCEPT

    # Allow DNS queries to the detected DNS server only
    iptables -A OUTPUT -p udp --dport 53 -d "$DNS_SERVER" -j ACCEPT
    iptables -A OUTPUT -p tcp --dport 53 -d "$DNS_SERVER" -j ACCEPT

    # Allow IPs in the allowed_ips ipset
    iptables -A OUTPUT -m set --match-set allowed_ips dst -j ACCEPT

    log "IPv4 iptables rules applied"
}

# =============================================================================
# ip6tables Setup (IPv6)
# =============================================================================

setup_ip6tables() {
    log "Setting up ip6tables (IPv6) rules..."

    # Check if ip6tables is available
    if ! command -v ip6tables &>/dev/null; then
        log "Warning: ip6tables not available, skipping IPv6 firewall setup"
        return
    fi

    # Flush existing OUTPUT rules
    ip6tables -F OUTPUT 2>/dev/null || true

    # Set default policy to DROP (block all IPv6 by default)
    ip6tables -P OUTPUT DROP

    # Allow loopback
    ip6tables -A OUTPUT -o lo -j ACCEPT

    # Allow established/related connections
    ip6tables -A OUTPUT -m state --state ESTABLISHED,RELATED -j ACCEPT

    # Allow ICMPv6 (required for IPv6 neighbor discovery and connectivity)
    ip6tables -A OUTPUT -p icmpv6 -j ACCEPT

    # Allow DNS queries to IPv6 DNS server if configured
    if [ -n "$DNS_SERVER_V6" ]; then
        ip6tables -A OUTPUT -p udp --dport 53 -d "$DNS_SERVER_V6" -j ACCEPT
        ip6tables -A OUTPUT -p tcp --dport 53 -d "$DNS_SERVER_V6" -j ACCEPT
    fi

    # Note: IPv6 addresses for allowed domains are not resolved by default
    # This provides a secure default that blocks IPv6 traffic which could
    # potentially bypass the IPv4 firewall rules.

    log "IPv6 ip6tables rules applied (default deny)"
}

# =============================================================================
# Firewall Verification
# =============================================================================

verify_firewall() {
    log "Verifying firewall rules..."

    # Test allowed domain (if any)
    if [ -n "$ALLOWED_DOMAINS" ]; then
        local first_domain
        first_domain=$(echo "$ALLOWED_DOMAINS" | cut -d',' -f1)

        # Try to reach the allowed domain
        local http_code
        http_code=$(curl -s --max-time 5 -o /dev/null -w "%{http_code}" "https://$first_domain" 2>/dev/null || echo "000")

        if [[ "$http_code" =~ ^(2[0-9][0-9]|3[0-9][0-9])$ ]]; then
            log "  PASS: Can reach $first_domain (HTTP $http_code)"
        elif [ "$http_code" == "000" ]; then
            log "  WARN: Cannot reach $first_domain (connection failed)"
        else
            log "  WARN: $first_domain returned HTTP $http_code"
        fi
    fi

    # Test that a blocked domain is actually blocked
    # Using a well-known domain that should not be in the allow list
    local blocked_test="example.com"
    if curl -s --max-time 3 -o /dev/null "https://$blocked_test" 2>/dev/null; then
        log "  FAIL: Can reach $blocked_test (should be blocked)"
    else
        log "  PASS: $blocked_test is blocked"
    fi

    log "Verification complete"
}

# =============================================================================
# Main
# =============================================================================

main() {
    log "Initializing firewall..."

    # Detect DNS server
    detect_dns

    # Create ipset
    setup_ipset

    # Create directory for sandbox runtime files
    mkdir -p /run/sandbox

    # Create shared DNS cache file for LD_PRELOAD hostname lookups
    touch /run/sandbox/dns_cache
    chmod 666 /run/sandbox/dns_cache

    # Process allowed entries (domains, IPs, or CIDRs)
    if [ -n "$ALLOWED_DOMAINS" ]; then
        IFS=',' read -ra ENTRIES <<< "$ALLOWED_DOMAINS"
        for entry in "${ENTRIES[@]}"; do
            entry=$(echo "$entry" | xargs)  # Trim whitespace
            [ -z "$entry" ] && continue

            # Detect entry type and handle accordingly
            local entry_type
            entry_type=$(detect_entry_type "$entry")

            case "$entry_type" in
                cidr)
                    add_cidr "$entry"
                    ;;
                ip)
                    add_ip "$entry"
                    ;;
                domain)
                    resolve_and_add "$entry"

                    # Special handling for GitHub
                    if [[ "$entry" == "github.com" || "$entry" == "*.github.com" ]]; then
                        fetch_github_ips
                    fi
                    ;;
            esac
        done
    else
        log "No ALLOWED_DOMAINS set, only DNS will be accessible"
    fi

    # Export allowed IPs to file for LD_PRELOAD library
    log "Exporting allowed IPs to /run/sandbox/allowed_ips..."
    ipset list allowed_ips | grep -E '^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+(/[0-9]+)?' > /run/sandbox/allowed_ips || true
    chmod 444 /run/sandbox/allowed_ips
    log "Exported $(wc -l < /run/sandbox/allowed_ips) allowed IPs"

    # Apply IPv4 iptables rules
    setup_iptables

    # Apply IPv6 ip6tables rules (default deny to prevent bypass)
    setup_ip6tables

    # Verify the firewall is working
    verify_firewall

    log "Firewall initialized successfully"

    # =============================================================================
    # Script Cleanup
    # =============================================================================
    # Make this script unreadable so the agent cannot inspect firewall rules.
    # This script runs as root from entrypoint.sh. After completion, the
    # entrypoint drops privileges to the agent user via setpriv.
    chmod 000 /usr/local/bin/init-firewall.sh 2>/dev/null || true

    log "Firewall script hidden from agent user"
}

main "$@"
