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
		{"--script=vuln", true},
		{"--script=vulners", true},
		// Unsafe
		{"--script", false},
		{"--script=dangerous.nse", false},
		{"--script-args", false},
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

func TestAllInOne360Profile_ArgsAreSafe(t *testing.T) {
	profiles := GetDefaultProfiles()
	profile, ok := profiles[ProfileAllInOne360]
	if !ok {
		t.Fatal("ProfileAllInOne360 non trovato nei profili di default")
	}

	for _, arg := range profile.Args {
		if !IsSafeArgument(arg) {
			t.Errorf("argomento in ProfileAllInOne360 '%s' non è considerato sicuro", arg)
		}
	}

	safeArgs, err := BuildSafeArguments(profile.Args, "10.0.0.5")
	if err != nil {
		t.Fatalf("BuildSafeArguments fallito per All-in-One: %v", err)
	}

	if len(safeArgs) < 5 {
		t.Fatalf("numero argomenti inatteso per All-in-One: %v", safeArgs)
	}
}

func TestParseProfileArgs_VulnerabilityScript(t *testing.T) {
	parsed, err := ParseProfileArgs("-sV -T4 --script vuln")
	if err != nil {
		t.Fatalf("ParseProfileArgs fallito con --script vuln: %v", err)
	}
	found := false
	for _, a := range parsed {
		if a == "--script=vuln" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("atteso flag '--script=vuln' negli argomenti, ottenuto: %v", parsed)
	}

	// Arbitrary script must fail
	_, err = ParseProfileArgs("-sV --script malicious.nse")
	if err == nil {
		t.Error("atteso errore per script arbitrario 'malicious.nse', ma nessun errore restituito")
	}
}

func TestSpecializedUDPProfile_ArgsAreSafe(t *testing.T) {
	profiles := GetDefaultProfiles()
	udpProfile, ok := profiles[ProfileSpecializedUDP]
	if !ok {
		t.Fatal("ProfileSpecializedUDP non trovato nei profili di default")
	}

	for _, arg := range udpProfile.Args {
		if arg == "-p" {
			continue
		}
		if !IsSafeArgument(arg) && !validPortArgRegex.MatchString("-p"+arg) {
			t.Errorf("argomento in ProfileSpecializedUDP '%s' non è considerato sicuro", arg)
		}
	}

	parsed, err := ParseProfileArgs("-sU -p U:53,69,123,161,514,623,5353 -T4")
	if err != nil {
		t.Fatalf("ParseProfileArgs fallito con argomenti UDP: %v", err)
	}

	safeArgs, err := BuildSafeArguments(parsed, "192.168.1.10")
	if err != nil {
		t.Fatalf("BuildSafeArguments fallito per UDP Profile: %v", err)
	}

	if len(safeArgs) < 5 {
		t.Fatalf("troppi pochi argomenti generati: %v", safeArgs)
	}
}

func TestDefaultProfiles_Invasiveness(t *testing.T) {
	profiles := GetDefaultProfiles()
	for pType, cfg := range profiles {
		if !cfg.Invasiveness.IsValid() {
			t.Errorf("profilo %s ha livello di invasività non valido: %s", pType, cfg.Invasiveness)
		}
	}
}


