package scanner

import (
	"testing"
)

func TestIsSafeArgument(t *testing.T) {
	tests := []struct {
		arg      string
		expected bool
	}{
		{"-sn", true},
		{"-PE", true},
		{"-PR", true},
		{"-PS22,80,443", true},
		{"-PS22,80,443,445,8080,3389", true},
		{"-PU53", true},
		{"-PU53,137", true},
		{"-T4", true},
		{"-T5", true},
		{"-n", true},
		{"-F", true},
		{"-sV", true},
		{"-p80,443", true},
		{"-p1-1000", true},
		{"-p-", true},
		// Unsafe
		{"--script", false},
		{"--script=vuln", false},
		{"-oN", false},
		{"-oX", false},
		{"; rm -rf /", false},
		{"-PS80;whoami", false},
		{"-PU53|calc", false},
		{"", false},
		{"-T9", false},
	}

	for _, tt := range tests {
		t.Run(tt.arg, func(t *testing.T) {
			got := IsSafeArgument(tt.arg)
			if got != tt.expected {
				t.Errorf("IsSafeArgument(%q) = %v; atteso %v", tt.arg, got, tt.expected)
			}
		})
	}
}

func TestLANDiscoveryProfile_ArgsAreSafe(t *testing.T) {
	profiles := GetDefaultProfiles()
	lanProfile, ok := profiles[ProfileLANDiscovery]
	if !ok {
		t.Fatal("ProfileLANDiscovery non trovato nei profili di default")
	}

	for _, arg := range lanProfile.Args {
		if !IsSafeArgument(arg) {
			t.Errorf("argomento in ProfileLANDiscovery '%s' non è considerato sicuro", arg)
		}
	}

	safeArgs, err := BuildSafeArguments(lanProfile.Args, "192.168.1.0/24")
	if err != nil {
		t.Fatalf("BuildSafeArguments fallito: %v", err)
	}

	if len(safeArgs) < 3 {
		t.Fatalf("troppi pochi argomenti generati: %v", safeArgs)
	}
	if safeArgs[len(safeArgs)-1] != "192.168.1.0/24" {
		t.Errorf("ultimo argomento atteso '192.168.1.0/24', ottenuto '%s'", safeArgs[len(safeArgs)-1])
	}
}

