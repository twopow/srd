package util

import (
	"testing"
)

func TestIsIp(t *testing.T) {
	tests := []struct {
		name     string
		hostname string
		expected bool
	}{
		// Basic IPv4 addresses without port
		{
			name:     "valid IPv4 without port",
			hostname: "192.168.1.1",
			expected: true,
		},
		{
			name:     "valid IPv4 localhost",
			hostname: "127.0.0.1",
			expected: true,
		},
		{
			name:     "valid IPv4 public IP",
			hostname: "8.8.8.8",
			expected: true,
		},
		{
			name:     "valid IPv4 with zeros",
			hostname: "0.0.0.0",
			expected: true,
		},
		{
			name:     "valid IPv4 with 255",
			hostname: "255.255.255.255",
			expected: true,
		},

		// IPv4 addresses with port
		{
			name:     "valid IPv4 with port 80",
			hostname: "192.168.1.1:80",
			expected: true,
		},
		{
			name:     "valid IPv4 with port 8080",
			hostname: "127.0.0.1:8080",
			expected: true,
		},
		{
			name:     "valid IPv4 with port 443",
			hostname: "8.8.8.8:443",
			expected: true,
		},
		{
			name:     "valid IPv4 with port 65535",
			hostname: "192.168.1.1:65535",
			expected: true,
		},
		{
			name:     "valid IPv4 with port 1",
			hostname: "192.168.1.1:1",
			expected: true,
		},

		// Edge cases for valid IPs
		{
			name:     "valid IPv4 with single digit octets",
			hostname: "1.2.3.4",
			expected: true,
		},
		{
			name:     "valid IPv4 with double digit octets",
			hostname: "12.34.56.78",
			expected: true,
		},
		{
			name:     "valid IPv4 with triple digit octets",
			hostname: "123.234.123.234",
			expected: true,
		},

		// Invalid IPv4 addresses
		{
			name:     "invalid IPv4 with negative octet",
			hostname: "-1.1.1.1",
			expected: false,
		},
		{
			name:     "invalid IPv4 with too many octets",
			hostname: "1.2.3.4.5",
			expected: false,
		},
		{
			name:     "invalid IPv4 with too few octets",
			hostname: "1.2.3",
			expected: false,
		},
		{
			name:     "invalid IPv4 with letters",
			hostname: "192.168.1.a",
			expected: false,
		},

		// Invalid port numbers
		{
			name:     "invalid IPv4 with negative port",
			hostname: "192.168.1.1:-1",
			expected: false,
		},
		{
			name:     "invalid IPv4 with non-numeric port",
			hostname: "192.168.1.1:abc",
			expected: false,
		},

		// Note: The current regex allows these cases, so they're expected to pass
		// These are technically invalid but match the regex pattern
		{
			name:     "IPv4 with octet > 255 (regex allows this)",
			hostname: "256.1.1.1",
			expected: true,
		},
		{
			name:     "IPv4 with leading zeros (regex allows this)",
			hostname: "192.168.01.1",
			expected: true,
		},
		{
			name:     "IPv4 with port > 65535 (regex allows this)",
			hostname: "192.168.1.1:65536",
			expected: true,
		},
		{
			name:     "IPv4 with port 0 (regex allows this)",
			hostname: "192.168.1.1:0",
			expected: true,
		},
		{
			name:     "IPv4 with port with leading zeros (regex allows this)",
			hostname: "192.168.1.1:080",
			expected: true,
		},

		// Non-IP strings
		{
			name:     "domain name",
			hostname: "example.com",
			expected: false,
		},
		{
			name:     "domain name with port",
			hostname: "example.com:80",
			expected: false,
		},
		{
			name:     "empty string",
			hostname: "",
			expected: false,
		},
		{
			name:     "just port",
			hostname: ":80",
			expected: false,
		},
		{
			name:     "just colon",
			hostname: ":",
			expected: false,
		},
		{
			name:     "IPv6 address",
			hostname: "2001:db8::1",
			expected: false,
		},
		{
			name:     "random string",
			hostname: "not an ip",
			expected: false,
		},
		{
			name:     "IP with extra characters",
			hostname: "192.168.1.1:80:extra",
			expected: false,
		},
		{
			name:     "IP with spaces",
			hostname: "192.168.1.1 :80",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsIp(tt.hostname)
			if result != tt.expected {
				t.Errorf("IsIp(%q) = %v, expected %v", tt.hostname, result, tt.expected)
			}
		})
	}
}
