package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeMAC(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"98-ba-5f-95-4e-50", "98:BA:5F:95:4E:50"},
		{"98:ba:5f:95:4e:50", "98:BA:5F:95:4E:50"},
		{"0:11:22:33:44:55", "00:11:22:33:44:55"},
		{"invalid-mac", ""},
		{"", ""},
		{"01:02:03:04:05", ""},
	}

	for _, tt := range tests {
		got := NormalizeMAC(tt.input)
		if got != tt.expected {
			t.Errorf("NormalizeMAC(%q) = %q; atteso %q", tt.input, got, tt.expected)
		}
	}
}

func TestARPCache_LoadFromRaw_Windows(t *testing.T) {
	windowsOutput := `
Interfaccia: 192.168.1.10 --- 0x14
  Indirizzo Internet    Indirizzo fisico      Tipo
  192.168.1.1           98-ba-5f-95-4e-50     dinamico  
  192.168.1.3           a0-d2-b1-c1-4a-9e     dinamico  
  192.168.1.14          8c-90-2d-12-b1-54     dinamico  
`
	cache := NewARPCache()
	cache.LoadFromRaw(windowsOutput)

	if got := cache.Lookup("192.168.1.1"); got != "98:BA:5F:95:4E:50" {
		t.Errorf("Lookup(192.168.1.1) = %q; atteso 98:BA:5F:95:4E:50", got)
	}
	if got := cache.Lookup("192.168.1.3"); got != "A0:D2:B1:C1:4A:9E" {
		t.Errorf("Lookup(192.168.1.3) = %q; atteso A0:D2:B1:C1:4A:9E", got)
	}
	if got := cache.Lookup("192.168.1.14"); got != "8C:90:2D:12:B1:54" {
		t.Errorf("Lookup(192.168.1.14) = %q; atteso 8C:90:2D:12:B1:54", got)
	}
	if got := cache.Lookup("192.168.1.99"); got != "" {
		t.Errorf("Lookup(192.168.1.99) non atteso: %q", got)
	}
}

func TestARPCache_LoadFromRaw_Linux(t *testing.T) {
	linuxProcNetArp := `IP address       HW type     Flags       HW address            Mask     Device
192.168.1.1      0x1         0x2         98:ba:5f:95:4e:50     *        eth0
192.168.1.5      0x1         0x2         ee:5b:26:72:a1:e1     *        eth0
`
	cache := NewARPCache()
	cache.LoadFromRaw(linuxProcNetArp)

	if got := cache.Lookup("192.168.1.1"); got != "98:BA:5F:95:4E:50" {
		t.Errorf("Lookup(192.168.1.1) = %q; atteso 98:BA:5F:95:4E:50", got)
	}
	if got := cache.Lookup("192.168.1.5"); got != "EE:5B:26:72:A1:E1" {
		t.Errorf("Lookup(192.168.1.5) = %q; atteso EE:5B:26:72:A1:E1", got)
	}
}

func TestARPCache_LoadFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "arp_test.txt")

	content := `{"192.168.1.20": "00:14:22:AA:BB:CC"}`
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("errore scrittura file temp: %v", err)
	}

	cache := NewARPCache()
	if err := cache.LoadFromFile(filePath); err != nil {
		t.Fatalf("LoadFromFile fallito: %v", err)
	}

	if got := cache.Lookup("192.168.1.20"); got != "00:14:22:AA:BB:CC" {
		t.Errorf("Lookup(192.168.1.20) = %q; atteso 00:14:22:AA:BB:CC", got)
	}
}

