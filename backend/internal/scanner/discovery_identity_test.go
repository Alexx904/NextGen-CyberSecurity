package scanner

import "testing"

func TestNormalizeMACRejectsInvalidIdentity(t *testing.T) {
	for _, mac := range []string{"GG:11:22:33:44:55", "00:00:00:00:00:00", "FF:FF:FF:FF:FF:FF", "01:00:5E:00:00:01"} {
		if got := NormalizeMAC(mac); got != "" {
			t.Errorf("%s accepted as %s", mac, got)
		}
	}
}

func TestSoftwareDoesNotIdentifyHardwareVendor(t *testing.T) {
	for _, hint := range []DeviceHint{{OSName: "Windows 11", Services: []string{"msrpc"}}, {ServerHeader: "nginx"}, {OSName: "Linux"}, {RawBanners: []string{"Samba", "BusyBox"}}} {
		if got := InferVendor(hint); got != "Unknown" {
			t.Errorf("software inferred hardware vendor %q", got)
		}
	}
}
