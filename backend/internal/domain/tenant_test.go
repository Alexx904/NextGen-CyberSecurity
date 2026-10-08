package domain_test

import (
	"testing"

	"nextgen-cybersec/internal/domain"
)

func TestIsValidHexColor(t *testing.T) {
	validColors := []string{
		"#fff",
		"#FFF",
		"#000",
		"#ffffff",
		"#FFFFFF",
		"#0f172a",
		"#2563eb",
		"#38bdf8",
		"#123456",
		"#abcdef",
		"#ABCDEF",
		" #ffffff ", // Con spazi da trimmare
	}

	for _, color := range validColors {
		if !domain.IsValidHexColor(color) {
			t.Errorf("atteso colore valido per '%s', ma restituito false", color)
		}
	}

	invalidColors := []string{
		"",
		"fff",
		"#ff",
		"#ffff",
		"#fffff",
		"#fffffff",
		"#gggggg",
		"rgb(255,255,255)",
		"red",
		"#12345z",
		"# 123456",
	}

	for _, color := range invalidColors {
		if domain.IsValidHexColor(color) {
			t.Errorf("atteso colore non valido per '%s', ma restituito true", color)
		}
	}
}

func TestBrandColors_IsValid(t *testing.T) {
	valid := domain.BrandColors{
		Primary:   "#0f172a",
		Secondary: "#2563eb",
		Accent:    "#38bdf8",
	}
	if !valid.IsValid() {
		t.Errorf("atteso BrandColors valido, ottenuto false")
	}

	invalidCases := []domain.BrandColors{
		{Primary: "invalid", Secondary: "#2563eb", Accent: "#38bdf8"},
		{Primary: "#0f172a", Secondary: "", Accent: "#38bdf8"},
		{Primary: "#0f172a", Secondary: "#2563eb", Accent: "#zzz"},
	}

	for i, c := range invalidCases {
		if c.IsValid() {
			t.Errorf("caso %d: atteso non valido, ottenuto valido per %+v", i, c)
		}
	}
}

func TestBrandColors_DefaultValidation(t *testing.T) {
	defaults := domain.DefaultBrandColors()
	if !defaults.IsValid() {
		t.Errorf("i default brand colors devono essere validi: %+v", defaults)
	}
	if defaults.Primary == "" || defaults.Secondary == "" || defaults.Accent == "" {
		t.Errorf("i colori di default non devono essere vuoti: %+v", defaults)
	}
}
