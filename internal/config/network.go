package config

import (
	"regexp"
	"strings"
)

// EntryType represents the type of a network allow list entry.
type EntryType int

const (
	// EntryTypeDomain indicates the entry is a domain name (e.g., "github.com").
	EntryTypeDomain EntryType = iota
	// EntryTypeIP indicates the entry is an IPv4 address (e.g., "10.0.0.1").
	EntryTypeIP
	// EntryTypeCIDR indicates the entry is a CIDR range (e.g., "192.168.1.0/24").
	EntryTypeCIDR
)

// String returns a string representation of the entry type.
func (t EntryType) String() string {
	switch t {
	case EntryTypeDomain:
		return "domain"
	case EntryTypeIP:
		return "ip"
	case EntryTypeCIDR:
		return "cidr"
	default:
		return "unknown"
	}
}

// ipv4Regex matches IPv4 addresses (e.g., "10.0.0.1", "192.168.1.100").
var ipv4Regex = regexp.MustCompile(`^\d+\.\d+\.\d+\.\d+$`)

// DetectEntryType determines the type of a network allow list entry.
// Detection logic:
//   - Contains "/" -> CIDR range (e.g., "192.168.1.0/24")
//   - Matches IPv4 pattern -> IP address (e.g., "10.0.0.1")
//   - Otherwise -> domain name (e.g., "github.com")
func DetectEntryType(entry string) EntryType {
	entry = strings.TrimSpace(entry)

	// CIDR detection: contains a slash
	if strings.Contains(entry, "/") {
		return EntryTypeCIDR
	}

	// IPv4 detection: matches pattern like "10.0.0.1"
	if ipv4Regex.MatchString(entry) {
		return EntryTypeIP
	}

	// Default: treat as domain name
	return EntryTypeDomain
}
