package scanner

import (
	"strings"
)

// DeviceHint racchiude gli indizi raccolti su un dispositivo per dedurne produttore e tipologia
type DeviceHint struct {
	Hostname     string
	OSName       string
	Services     []string
	RawBanners   []string
	HTTPTitle    string
	ServerHeader string
}

// InferVendor deduce il produttore hardware o la tipologia del dispositivo in base agli indizi raccolti
func InferVendor(hint DeviceHint) string {
	combined := strings.ToLower(
		hint.Hostname + " " +
			hint.OSName + " " +
			strings.Join(hint.Services, " ") + " " +
			strings.Join(hint.RawBanners, " ") + " " +
			hint.HTTPTitle + " " +
			hint.ServerHeader,
	)

	switch {
	case strings.Contains(combined, "tp-link") || strings.Contains(combined, "tplink"):
		return "TP-Link"
	case strings.Contains(combined, "apple") || strings.Contains(combined, "airplay") || strings.Contains(combined, "darwin") || strings.Contains(combined, "macos") || strings.Contains(combined, "iphone") || strings.Contains(combined, "ipad") || strings.Contains(combined, "macbook"):
		return "Apple"
	case strings.Contains(combined, "samsung") || strings.Contains(combined, "tizen"):
		return "Samsung"
	case strings.Contains(combined, "cisco"):
		return "Cisco Systems"
	case strings.Contains(combined, "xiaomi") || strings.Contains(combined, "mi router"):
		return "Xiaomi"
	case strings.Contains(combined, "huawei"):
		return "Huawei"
	case strings.Contains(combined, "raspberry") || strings.Contains(combined, "raspbian"):
		return "Raspberry Pi Foundation"
	case strings.Contains(combined, "dell"):
		return "Dell"
	case strings.Contains(combined, "hp ") || strings.Contains(combined, "hewlett-packard") || strings.Contains(combined, "laserjet"):
		return "HP"
	case strings.Contains(combined, "canon"):
		return "Canon"
	case strings.Contains(combined, "epson"):
		return "Epson"
	case strings.Contains(combined, "brother"):
		return "Brother"
	case strings.Contains(combined, "synology") || strings.Contains(combined, "diskstation"):
		return "Synology"
	case strings.Contains(combined, "qnap"):
		return "QNAP"
	case strings.Contains(combined, "espressif") || strings.Contains(combined, "esp32") || strings.Contains(combined, "esp8266"):
		return "Espressif IoT"
	case strings.Contains(combined, "fritz!box") || strings.Contains(combined, "avm"):
		return "AVM FRITZ!"
	case strings.Contains(combined, "ubiquiti") || strings.Contains(combined, "unifi"):
		return "Ubiquiti Networks"
	case strings.Contains(combined, "mikrotik") || strings.Contains(combined, "routeros"):
		return "MikroTik"
	case strings.Contains(combined, "vmware"):
		return "VMware"
	default:
		return "Unknown"
	}
}

// CleanDeviceHostname suggerisce un hostname leggibile partendo da PrimaryHostname, Titolo Web o indicazioni di servizio
func CleanDeviceHostname(currentHostname, httpTitle, vendor string) string {
	curr := strings.TrimSpace(currentHostname)
	title := strings.TrimSpace(httpTitle)

	// Se il titolo HTTP è informativo (non generico come "404" o "error" o vuoto)
	if title != "" && !strings.EqualFold(title, "error") && !strings.EqualFold(title, "404 not found") && !strings.EqualFold(title, "403 forbidden") {
		// Se non abbiamo hostname o l'hostname attuale è generico
		if curr == "" || strings.HasPrefix(curr, "ip-") {
			return title
		}
		// Se il titolo aggiunge contesto utile
		if !strings.EqualFold(curr, title) {
			return curr + " (" + title + ")"
		}
	}

	if curr != "" {
		return curr
	}

	if vendor != "" && vendor != "Unknown" {
		if strings.HasPrefix(vendor, "Private MAC") {
			return "Mobile Device"
		}
		return vendor + " Device"
	}

	return ""
}
