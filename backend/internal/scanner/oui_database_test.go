package scanner

import (
	"testing"
)

func TestNormalizeMACPrefix(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  string
		expectErr bool
	}{
		{
			name:      "Standard colon format lowercase",
			input:     "b8:27:eb:12:34:56",
			expected:  "B8:27:EB",
			expectErr: false,
		},
		{
			name:      "Standard colon format uppercase",
			input:     "B8:27:EB:AA:BB:CC",
			expected:  "B8:27:EB",
			expectErr: false,
		},
		{
			name:      "Hyphen format lowercase",
			input:     "00-14-22-1a-2b-3c",
			expected:  "00:14:22",
			expectErr: false,
		},
		{
			name:      "Cisco dot format",
			input:     "0001.42ab.cdef",
			expected:  "00:01:42",
			expectErr: false,
		},
		{
			name:      "Compact hex string",
			input:     "3c0754112233",
			expected:  "3C:07:54",
			expectErr: false,
		},
		{
			name:      "3-octet prefix only",
			input:     "DC:A6:32",
			expected:  "DC:A6:32",
			expectErr: false,
		},
		{
			name:      "Too short string",
			input:     "00:11",
			expected:  "",
			expectErr: true,
		},
		{
			name:      "Empty string",
			input:     "",
			expected:  "",
			expectErr: true,
		},
		{
			name:      "Invalid hex characters",
			input:     "ZZ:YY:XX:11:22:33",
			expected:  "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeMACPrefix(tt.input)
			if tt.expectErr && err == nil {
				t.Fatalf("NormalizeMACPrefix(%q) atteso errore, ricevuto nil", tt.input)
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("NormalizeMACPrefix(%q) errore inatteso: %v", tt.input, err)
			}
			if got != tt.expected {
				t.Errorf("NormalizeMACPrefix(%q) = %q, atteso %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestOUIDatabase_Lookup(t *testing.T) {
	db := NewOUIDatabase()

	tests := []struct {
		name     string
		mac      string
		expected string
	}{
		{
			name:     "Raspberry Pi Foundation",
			mac:      "b8:27:eb:55:66:77",
			expected: "Raspberry Pi Foundation",
		},
		{
			name:     "Apple device",
			mac:      "3C:07:54:AA:BB:CC",
			expected: "Apple",
		},
		{
			name:     "Cisco Systems router",
			mac:      "00-01-42-11-22-33",
			expected: "Cisco Systems",
		},
		{
			name:     "Dell workstation",
			mac:      "18:03:73:de:ad:be",
			expected: "Dell",
		},
		{
			name:     "Espressif IoT (ESP32)",
			mac:      "24:0a:c4:12:34:56",
			expected: "Espressif Systems",
		},
		{
			name:     "VMware virtual machine",
			mac:      "00:50:56:c0:00:08",
			expected: "VMware",
		},
		{
			name:     "Microsoft Hyper-V NIC",
			mac:      "00:15:5D:01:02:03",
			expected: "Microsoft",
		},
		{
			name:     "Ubiquiti Networks AP",
			mac:      "dc:9f:db:aa:bb:cc",
			expected: "Ubiquiti Networks",
		},
		{
			name:     "Unknown valid MAC",
			mac:      "FF:EE:DD:00:11:22",
			expected: "Unknown",
		},
		{
			name:     "Empty string",
			mac:      "",
			expected: "Unknown",
		},
		{
			name:     "Malformed input",
			mac:      "invalid-mac-address",
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := db.Lookup(tt.mac)
			if got != tt.expected {
				t.Errorf("Lookup(%q) = %q, atteso %q", tt.mac, got, tt.expected)
			}
		})
	}
}

