package scanner

import (
	"testing"
)

func TestInferVendor(t *testing.T) {
	tests := []struct {
		name     string
		hint     DeviceHint
		expected string
	}{
		{
			name: "TP-LINK Router from Banners",
			hint: DeviceHint{
				Hostname:   "",
				OSName:     "Linux",
				Services:   []string{"http", "ssh", "domain"},
				RawBanners: []string{"Dropbear sshd", "dnsmasq", "TP-LINK WAP"},
				HTTPTitle:  "TP-Link Wireless Router",
			},
			expected: "TP-Link",
		},
		{
			name: "Windows PC with MSRPC",
			hint: DeviceHint{
				OSName:   "Windows 11",
				Services: []string{"msrpc", "microsoft-ds", "netbios-ssn"},
			},
			expected: "Microsoft",
		},
		{
			name: "Samsung TV or Phone",
			hint: DeviceHint{
				OSName:    "Tizen OS",
				HTTPTitle: "Samsung Smart TV",
			},
			expected: "Samsung",
		},
		{
			name: "Apple Device via AirPlay",
			hint: DeviceHint{
				Services: []string{"airplay", "raop"},
			},
			expected: "Apple",
		},
		{
			name: "Xiaomi Device",
			hint: DeviceHint{
				HTTPTitle: "Mi Router 4A",
			},
			expected: "Xiaomi",
		},
		{
			name: "Embedded Linux Router",
			hint: DeviceHint{
				RawBanners: []string{"Dropbear sshd 2020.80", "BusyBox telnetd"},
			},
			expected: "Linux Embedded / Router",
		},
		{
			name: "Empty / Unknown Hint",
			hint: DeviceHint{},
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InferVendor(tt.hint)
			if got != tt.expected {
				t.Errorf("InferVendor(%s) = %q; atteso %q", tt.name, got, tt.expected)
			}
		})
	}
}

func TestCleanDeviceHostname(t *testing.T) {
	tests := []struct {
		curr     string
		title    string
		vendor   string
		expected string
	}{
		{"", "TP-Link Wireless Router", "TP-Link", "TP-Link Wireless Router"},
		{"router.lan", "welcome", "TP-Link", "router.lan (welcome)"},
		{"desktop-pc", "", "Microsoft", "desktop-pc"},
		{"", "", "Apple", "Apple Device"},
		{"", "", "", ""},
		{"", "404 Not Found", "Unknown", ""},
	}

	for _, tt := range tests {
		got := CleanDeviceHostname(tt.curr, tt.title, tt.vendor)
		if got != tt.expected {
			t.Errorf("CleanDeviceHostname(%q, %q, %q) = %q; atteso %q", tt.curr, tt.title, tt.vendor, got, tt.expected)
		}
	}
}

