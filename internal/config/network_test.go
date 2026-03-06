package config

import "testing"

func TestDetectEntryType(t *testing.T) {
	tests := []struct {
		name     string
		entry    string
		expected EntryType
	}{
		// Domain names
		{
			name:     "simple domain",
			entry:    "github.com",
			expected: EntryTypeDomain,
		},
		{
			name:     "subdomain",
			entry:    "api.github.com",
			expected: EntryTypeDomain,
		},
		{
			name:     "multi-level subdomain",
			entry:    "cdn-lfs.huggingface.co",
			expected: EntryTypeDomain,
		},
		{
			name:     "domain with numbers",
			entry:    "ec2.us-east-1.amazonaws.com",
			expected: EntryTypeDomain,
		},

		// IPv4 addresses
		{
			name:     "simple IPv4",
			entry:    "10.0.0.1",
			expected: EntryTypeIP,
		},
		{
			name:     "private IP",
			entry:    "192.168.1.100",
			expected: EntryTypeIP,
		},
		{
			name:     "public IP",
			entry:    "8.8.8.8",
			expected: EntryTypeIP,
		},
		{
			name:     "all zeros",
			entry:    "0.0.0.0",
			expected: EntryTypeIP,
		},
		{
			name:     "broadcast",
			entry:    "255.255.255.255",
			expected: EntryTypeIP,
		},

		// CIDR ranges
		{
			name:     "class C network",
			entry:    "192.168.1.0/24",
			expected: EntryTypeCIDR,
		},
		{
			name:     "class B network",
			entry:    "172.16.0.0/16",
			expected: EntryTypeCIDR,
		},
		{
			name:     "single host CIDR",
			entry:    "10.0.0.1/32",
			expected: EntryTypeCIDR,
		},
		{
			name:     "large CIDR block",
			entry:    "10.0.0.0/8",
			expected: EntryTypeCIDR,
		},

		// Edge cases with whitespace
		{
			name:     "domain with leading space",
			entry:    "  github.com",
			expected: EntryTypeDomain,
		},
		{
			name:     "IP with trailing space",
			entry:    "10.0.0.1  ",
			expected: EntryTypeIP,
		},
		{
			name:     "CIDR with spaces",
			entry:    "  192.168.1.0/24  ",
			expected: EntryTypeCIDR,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectEntryType(tt.entry)
			if result != tt.expected {
				t.Errorf("DetectEntryType(%q) = %v, want %v", tt.entry, result, tt.expected)
			}
		})
	}
}

func TestEntryTypeString(t *testing.T) {
	tests := []struct {
		entryType EntryType
		expected  string
	}{
		{EntryTypeDomain, "domain"},
		{EntryTypeIP, "ip"},
		{EntryTypeCIDR, "cidr"},
		{EntryType(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.entryType.String(); got != tt.expected {
				t.Errorf("EntryType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}
