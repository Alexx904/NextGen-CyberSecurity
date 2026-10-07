package validator_test

import (
	"errors"
	"testing"

	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/pkg/validator"
)

func TestValidateTarget_IP(t *testing.T) {
	// IPv4 validi
	validIPv4 := []string{"127.0.0.1", "192.168.1.1", "10.0.0.1", "8.8.8.8"}
	for _, ip := range validIPv4 {
		info, err := validator.ValidateTarget(ip)
		if err != nil {
			t.Errorf("Atteso IP valido per '%s', ottenuto errore: %v", ip, err)
		}
		if info.Type != domain.TargetTypeIP {
			t.Errorf("Atteso Type TargetTypeIP per '%s', ottenuto %v", ip, info.Type)
		}
		if info.IsIPv6 {
			t.Errorf("Atteso IsIPv6 = false per '%s'", ip)
		}
	}

	// IPv6 validi
	validIPv6 := []string{"::1", "2001:db8::1", "fe80::1ff:fe23:4567:890a"}
	for _, ip := range validIPv6 {
		info, err := validator.ValidateTarget(ip)
		if err != nil {
			t.Errorf("Atteso IPv6 valido per '%s', ottenuto errore: %v", ip, err)
		}
		if info.Type != domain.TargetTypeIP || !info.IsIPv6 {
			t.Errorf("Atteso TargetTypeIP e IsIPv6 = true per '%s'", ip)
		}
	}

	// IP non validi
	invalidIP := []string{"256.0.0.1", "192.168.1", "192.168.1.1.1", ":::1:::"}
	for _, ip := range invalidIP {
		_, err := validator.ValidateTarget(ip)
		if err == nil {
			t.Errorf("Atteso errore per IP non valido '%s'", ip)
		}
	}
}

func TestValidateTarget_CIDR(t *testing.T) {
	// CIDR validi
	validCIDR := []struct {
		cidr   string
		isIPv6 bool
	}{
		{"192.168.1.0/24", false},
		{"10.0.0.0/8", false},
		{"172.16.0.0/16", false},
		{"2001:db8::/32", true},
		{"::/0", true},
	}

	for _, tc := range validCIDR {
		info, err := validator.ValidateTarget(tc.cidr)
		if err != nil {
			t.Errorf("Atteso CIDR valido per '%s', ottenuto errore: %v", tc.cidr, err)
		}
		if info.Type != domain.TargetTypeCIDR {
			t.Errorf("Atteso TargetTypeCIDR per '%s', ottenuto %v", tc.cidr, info.Type)
		}
		if info.IsIPv6 != tc.isIPv6 {
			t.Errorf("Atteso isIPv6=%v per '%s', ottenuto %v", tc.isIPv6, tc.cidr, info.IsIPv6)
		}
	}

	// CIDR non validi
	invalidCIDR := []string{"192.168.1.0/33", "10.0.0.0/-1", "10.0.0.0/abc"}
	for _, cidr := range invalidCIDR {
		_, err := validator.ValidateTarget(cidr)
		if err == nil {
			t.Errorf("Atteso errore per CIDR non valido '%s'", cidr)
		}
	}
}

func TestValidateTarget_Range(t *testing.T) {
	// Range validi
	validRanges := []string{
		"192.168.1.1-50",
		"10.0.0.1-254",
		"192.168.1.10-192.168.1.50",
	}

	for _, r := range validRanges {
		info, err := validator.ValidateTarget(r)
		if err != nil {
			t.Errorf("Atteso Range valido per '%s', ottenuto errore: %v", r, err)
		}
		if info.Type != domain.TargetTypeRange {
			t.Errorf("Atteso TargetTypeRange per '%s', ottenuto %v", r, info.Type)
		}
	}

	// Range non validi (invertiti o malformati)
	invalidRanges := []string{
		"192.168.1.50-10",              // invertito
		"192.168.1.1-300",             // ottetto > 255
		"192.168.1.1-",                // parte destra mancante
		"192.168.1.50-192.168.1.10",   // invertito completo
	}

	for _, r := range invalidRanges {
		_, err := validator.ValidateTarget(r)
		if err == nil {
			t.Errorf("Atteso errore per Range non valido '%s'", r)
		}
	}
}

func TestValidateTarget_FQDN(t *testing.T) {
	// FQDN validi
	validDomains := []string{
		"scanme.nmap.org",
		"internal.corp.local",
		"api.sub.example.com",
		"test-server.example.org",
	}

	for _, d := range validDomains {
		info, err := validator.ValidateTarget(d)
		if err != nil {
			t.Errorf("Atteso FQDN valido per '%s', ottenuto errore: %v", d, err)
		}
		if info.Type != domain.TargetTypeFQDN {
			t.Errorf("Atteso TargetTypeFQDN per '%s', ottenuto %v", d, info.Type)
		}
	}

	// FQDN non validi
	invalidDomains := []string{
		"-invalid.com",
		"domain-.com",
		"singleword",      // nessun TLD
		"domain..com",     // doppio punto
		"domain.c",        // TLD < 2 caratteri
		"domain.123",      // TLD numerico
		"domain_with_underscore.com",
	}

	for _, d := range invalidDomains {
		_, err := validator.ValidateTarget(d)
		if err == nil {
			t.Errorf("Atteso errore per dominio non valido '%s'", d)
		}
	}
}

func TestValidateTarget_SecurityAndInjectionPrevention(t *testing.T) {
	maliciousInputs := []string{
		"-sS",
		"--script=vuln",
		"-oN output.txt",
		"192.168.1.1; rm -rf /",
		"192.168.1.1 && ls",
		"192.168.1.1 | cat",
		"`whoami`.example.com",
		"$(id).example.com",
		"/etc/passwd",
		"192.168.1.1\x00",
		"192.168.1.1 192.168.1.2",
		"",
		"   ",
	}

	for _, input := range maliciousInputs {
		_, err := validator.ValidateTarget(input)
		if !errors.Is(err, domain.ErrTargetInvalid) {
			t.Errorf("Atteso ErrTargetInvalid per input malevolo '%s', ottenuto: %v", input, err)
		}
	}
}

func TestIsTargetBlacklisted(t *testing.T) {
	blacklist := []string{
		"127.0.0.1",
		"10.0.0.0/16",
		"192.168.100.1-50",
		"internal.secret.corp",
	}

	testCases := []struct {
		target        string
		isBlacklisted bool
	}{
		// IP esatto in blacklist
		{"127.0.0.1", true},
		// IP libero
		{"127.0.0.2", false},
		// IP interno al CIDR blacklist 10.0.0.0/16
		{"10.0.5.23", true},
		// IP esterno al CIDR blacklist 10.0.0.0/16
		{"10.1.0.1", false},
		// CIDR che si sovrappone a 10.0.0.0/16
		{"10.0.0.0/24", true},
		{"10.0.0.0/8", true}, // include la blacklist
		// CIDR che include l'IP 127.0.0.1
		{"127.0.0.0/24", true},
		// Range che interseca il range blacklist 192.168.100.1-50
		{"192.168.100.10-20", true},
		{"192.168.100.60-80", false},
		// FQDN esatto
		{"internal.secret.corp", true},
		// Sotto-dominio del FQDN blacklist
		{"dev.internal.secret.corp", true},
		{"sub.dev.internal.secret.corp", true},
		// Dominio diverso
		{"public.corp.com", false},
	}

	for _, tc := range testCases {
		blacklisted, err := validator.IsTargetBlacklisted(tc.target, blacklist)
		if err != nil {
			t.Fatalf("Errore imprevisto durante IsTargetBlacklisted per '%s': %v", tc.target, err)
		}
		if blacklisted != tc.isBlacklisted {
			t.Errorf("Target '%s': atteso blacklisted=%v, ottenuto %v", tc.target, tc.isBlacklisted, blacklisted)
		}
	}
}

