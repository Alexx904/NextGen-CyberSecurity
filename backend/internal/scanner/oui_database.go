package scanner

import (
	"fmt"
	"strings"
	"unicode"
)

// OUIDatabase astrae il motore di traduzione OUI MAC -> Vendor hardware
type OUIDatabase interface {
	Lookup(mac string) string
}

type ouiDatabase struct {
	prefixes map[string]string
}

// NormalizeMACPrefix estrae e normalizza i primi 3 ottetti (OUI a 24 bit) nel formato "XX:XX:XX" maiuscolo
func NormalizeMACPrefix(mac string) (string, error) {
	trimmed := strings.TrimSpace(mac)
	if trimmed == "" {
		return "", fmt.Errorf("indirizzo MAC vuoto")
	}

	var hexChars []rune
	for _, r := range trimmed {
		if r == ':' || r == '-' || r == '.' || unicode.IsSpace(r) {
			continue
		}
		if !unicode.Is(unicode.Hex_Digit, r) {
			return "", fmt.Errorf("carattere non esadecimale '%c' in indirizzo MAC: '%s'", r, mac)
		}
		hexChars = append(hexChars, unicode.ToUpper(r))
	}

	// 2. Un prefisso OUI richiede almeno 6 caratteri esadecimali (3 byte)
	if len(hexChars) < 6 {
		return "", fmt.Errorf("indirizzo MAC troppo corto per estrarre OUI: '%s'", mac)
	}

	// 3. Compone il formato canonico XX:XX:XX
	prefix := fmt.Sprintf("%c%c:%c%c:%c%c",
		hexChars[0], hexChars[1],
		hexChars[2], hexChars[3],
		hexChars[4], hexChars[5],
	)

	return prefix, nil
}

// NewOUIDatabase inizializza il database locale OUI precaricato con vendor enterprise, consumer e IoT
func NewOUIDatabase() OUIDatabase {
	db := &ouiDatabase{
		prefixes: make(map[string]string, len(embeddedOUIMap)),
	}

	for k, v := range embeddedOUIMap {
		db.prefixes[strings.ToUpper(k)] = v
	}

	return db
}

// Lookup restituisce il nome del produttore hardware associato al MAC, o "Unknown" se non censito
func (db *ouiDatabase) Lookup(mac string) string {
	if strings.TrimSpace(mac) == "" {
		return "Unknown"
	}

	prefix, err := NormalizeMACPrefix(mac)
	if err != nil {
		return "Unknown"
	}

	if vendor, ok := db.prefixes[prefix]; ok {
		return vendor
	}

	// Se non è censito, verifica se il MAC è ad amministrazione locale (Randomized/Private MAC)
	if IsLocallyAdministeredMAC(prefix) {
		return "Private MAC (Mobile/Wi-Fi)"
	}

	return "Unknown"
}

// IsLocallyAdministeredMAC verifica se il MAC adotta un indirizzo randomizzato/privato (bit U/L impostato a 1 nel primo byte)
func IsLocallyAdministeredMAC(prefix string) bool {
	if len(prefix) < 2 {
		return false
	}
	var b byte
	if _, err := fmt.Sscanf(prefix[:2], "%X", &b); err == nil {
		// Conforme IEEE 802: Bit 0 deve essere 0 (Unicast) e Bit 1 deve essere 1 (Locally Administered)
		return (b & 0x03) == 0x02
	}
	return false
}

// embeddedOUIMap cataloga i prefissi MAC più diffusi per reti aziendali, virtualization e dispositivi IoT
var embeddedOUIMap = map[string]string{
	// Sagemcom / Router / ISP Gateways
	"98:BA:5F": "Sagemcom Broadband / Router",
	"00:19:70": "Sagemcom",
	"DC:53:7C": "Sagemcom",

	// Samsung
	"A0:D2:B1": "Samsung",

	// Xiaomi
	"8C:90:2D": "Xiaomi",
	"04:CF:8C": "Xiaomi",
	"14:F6:5A": "Xiaomi",
	"28:6C:07": "Xiaomi",
	"34:80:B3": "Xiaomi",

	// Componenti Hardware / Wi-Fi
	"50:07:C3": "Murata Manufacturing",
	"8C:17:59": "Liteon Technology",
	"0C:EE:99": "Apple",
	// Apple
	"00:17:F2": "Apple",
	"00:1C:B3": "Apple",
	"3C:07:54": "Apple",
	"AC:DE:48": "Apple",
	"F0:18:98": "Apple",
	"F4:0F:24": "Apple",
	"00:03:93": "Apple",
	"00:05:02": "Apple",
	"00:0A:95": "Apple",
	"00:11:24": "Apple",
	"00:14:51": "Apple",
	"00:16:CB": "Apple",
	"00:19:E3": "Apple",
	"00:1B:63": "Apple",
	"00:1E:52": "Apple",
	"00:1F:5B": "Apple",
	"00:1F:F3": "Apple",
	"00:21:E9": "Apple",
	"00:22:41": "Apple",
	"00:23:12": "Apple",
	"00:23:32": "Apple",
	"00:23:6C": "Apple",
	"00:24:36": "Apple",
	"00:25:00": "Apple",
	"00:25:4B": "Apple",
	"00:25:BC": "Apple",
	"00:26:08": "Apple",
	"00:26:4A": "Apple",
	"00:26:B0": "Apple",
	"18:65:90": "Apple",
	"28:CF:E9": "Apple",
	"34:15:9E": "Apple",
	"40:6C:8F": "Apple",
	"44:2A:60": "Apple",
	"48:D7:05": "Apple",
	"68:5B:35": "Apple",
	"70:56:81": "Apple",
	"74:E1:B6": "Apple",
	"78:31:C1": "Apple",
	"88:66:5A": "Apple",
	"A4:83:E7": "Apple",
	"BC:52:B7": "Apple",
	"CC:29:F5": "Apple",
	"D8:96:95": "Apple",
	"E0:B9:BA": "Apple",
	"F8:1E:DF": "Apple",

	// Cisco Systems
	"00:00:0C": "Cisco Systems",
	"00:01:42": "Cisco Systems",
	"00:01:43": "Cisco Systems",
	"00:01:63": "Cisco Systems",
	"00:01:64": "Cisco Systems",
	"00:1B:54": "Cisco Systems",
	"00:40:96": "Cisco Systems",
	"70:81:05": "Cisco Systems",
	"F8:C0:01": "Cisco Systems",
	"00:08:2F": "Cisco Systems",
	"00:08:7C": "Cisco Systems",
	"00:0E:D7": "Cisco Systems",
	"00:13:80": "Cisco Systems",
	"00:14:F2": "Cisco Systems",
	"00:17:59": "Cisco Systems",
	"00:18:74": "Cisco Systems",
	"00:19:07": "Cisco Systems",
	"00:1A:6D": "Cisco Systems",
	"00:1C:58": "Cisco Systems",
	"00:1D:45": "Cisco Systems",
	"00:22:BD": "Cisco Systems",
	"00:24:97": "Cisco Systems",
	"00:26:0B": "Cisco Systems",
	"00:27:0D": "Cisco Systems",
	"54:75:D0": "Cisco Systems",
	"58:97:1E": "Cisco Systems",
	"88:F0:31": "Cisco Systems",

	// Dell
	"00:14:22": "Dell",
	"00:18:8B": "Dell",
	"18:03:73": "Dell",
	"24:B6:FD": "Dell",
	"74:86:7A": "Dell",
	"84:2B:2B": "Dell",
	"B8:AC:6F": "Dell",
	"F8:DB:88": "Dell",
	"00:06:5B": "Dell",
	"00:08:74": "Dell",
	"00:0B:DB": "Dell",
	"00:0D:56": "Dell",
	"00:0F:1F": "Dell",
	"00:11:43": "Dell",
	"00:12:3F": "Dell",
	"00:13:72": "Dell",
	"00:15:C5": "Dell",
	"00:16:44": "Dell",
	"00:19:B9": "Dell",
	"00:1A:A0": "Dell",
	"00:1C:23": "Dell",
	"00:1D:09": "Dell",
	"00:1E:4F": "Dell",
	"00:21:9B": "Dell",
	"00:22:19": "Dell",
	"00:23:AE": "Dell",
	"00:24:E8": "Dell",
	"00:25:64": "Dell",
	"14:FE:B5": "Dell",
	"54:48:10": "Dell",
	"90:B1:1C": "Dell",
	"D4:BE:D9": "Dell",

	// Raspberry Pi Foundation
	"B8:27:EB": "Raspberry Pi Foundation",
	"DC:A6:32": "Raspberry Pi Foundation",
	"E4:5F:01": "Raspberry Pi Foundation",
	"28:CD:C1": "Raspberry Pi Foundation",
	"D8:3A:DD": "Raspberry Pi Foundation",

	// Espressif Systems (ESP32 / ESP8266)
	"24:0A:C4": "Espressif Systems",
	"24:62:AB": "Espressif Systems",
	"30:AE:A4": "Espressif Systems",
	"84:F3:EB": "Espressif Systems",
	"A4:CF:12": "Espressif Systems",
	"C4:4F:33": "Espressif Systems",
	"DC:4F:22": "Espressif Systems",
	"3C:61:05": "Espressif Systems",
	"3C:71:BF": "Espressif Systems",
	"68:C6:3A": "Espressif Systems",
	"80:7D:3A": "Espressif Systems",
	"94:B9:7E": "Espressif Systems",
	"BC:FF:4D": "Espressif Systems",
	"EC:62:60": "Espressif Systems",

	// HP / Hewlett Packard
	"00:08:02": "Hewlett Packard",
	"00:0B:CD": "Hewlett Packard",
	"00:0E:7F": "Hewlett Packard",
	"00:11:85": "Hewlett Packard",
	"00:15:60": "Hewlett Packard",
	"3C:D9:2B": "Hewlett Packard",
	"9C:8E:99": "Hewlett Packard",
	"FC:15:B4": "Hewlett Packard",
	"00:1A:4B": "Hewlett Packard",
	"00:1C:C4": "Hewlett Packard",
	"00:1E:0B": "Hewlett Packard",
	"00:21:5A": "Hewlett Packard",
	"00:23:7D": "Hewlett Packard",
	"00:24:81": "Hewlett Packard",
	"00:25:B3": "Hewlett Packard",
	"10:1F:74": "Hewlett Packard",
	"70:5A:0F": "Hewlett Packard",

	// Intel
	"00:02:B3": "Intel",
	"00:03:47": "Intel",
	"00:04:23": "Intel",
	"00:0E:0C": "Intel",
	"00:13:E8": "Intel",
	"00:15:00": "Intel",
	"34:13:E8": "Intel",
	"84:3A:4B": "Intel",
	"00:1B:21": "Intel",
	"00:1C:C0": "Intel",
	"00:1E:67": "Intel",
	"00:21:6A": "Intel",
	"00:22:FB": "Intel",
	"00:24:D7": "Intel",
	"00:26:C6": "Intel",
	"68:05:CA": "Intel",
	"A0:36:9F": "Intel",

	// VMware
	"00:50:56": "VMware",
	"00:0C:29": "VMware",
	"00:05:69": "VMware",
	"00:1C:14": "VMware",

	// Microsoft Hyper-V & Devices
	"00:15:5D": "Microsoft",
	"00:03:FF": "Microsoft",
	"00:12:5A": "Microsoft",
	"00:17:FA": "Microsoft",
	"7C:1E:52": "Microsoft",
	"DC:B4:C4": "Microsoft",
	"00:1D:D8": "Microsoft",
	"00:50:F2": "Microsoft",
	"28:18:78": "Microsoft",
	"60:45:BD": "Microsoft",

	// Google
	"00:1A:11": "Google",
	"3C:5A:37": "Google",
	"54:60:09": "Google",
	"70:3A:CB": "Google",
	"F4:F5:D8": "Google",
	"08:9E:08": "Google",
	"14:C1:4E": "Google",
	"20:DF:B9": "Google",
	"30:FD:38": "Google",
	"F8:8F:C2": "Google",

	// Samsung
	"00:07:AB": "Samsung",
	"00:12:47": "Samsung",
	"00:15:99": "Samsung",
	"00:16:32": "Samsung",
	"34:23:BA": "Samsung",
	"50:56:BF": "Samsung",
	"68:EB:AE": "Samsung",
	"94:65:2D": "Samsung",
	"00:1D:25": "Samsung",
	"00:21:D1": "Samsung",
	"00:23:D7": "Samsung",
	"00:26:5D": "Samsung",
	"18:22:7E": "Samsung",

	// Ubiquiti Networks
	"00:15:6D": "Ubiquiti Networks",
	"00:27:22": "Ubiquiti Networks",
	"24:A4:3C": "Ubiquiti Networks",
	"68:72:51": "Ubiquiti Networks",
	"74:83:C2": "Ubiquiti Networks",
	"78:8A:20": "Ubiquiti Networks",
	"80:2A:A8": "Ubiquiti Networks",
	"B4:FB:E4": "Ubiquiti Networks",
	"DC:9F:DB": "Ubiquiti Networks",
	"F0:9F:C2": "Ubiquiti Networks",
	"04:18:D6": "Ubiquiti Networks",
	"18:E8:29": "Ubiquiti Networks",
	"44:D9:E7": "Ubiquiti Networks",

	// TP-Link
	"00:27:19": "TP-Link",
	"14:CC:20": "TP-Link",
	"50:3E:AA": "TP-Link",
	"60:E3:27": "TP-Link",
	"C0:25:E9": "TP-Link",
	"E8:48:B8": "TP-Link",
	"EC:08:6B": "TP-Link",
	"F4:EC:38": "TP-Link",
	"30:DE:4B": "TP-Link",
	"54:AF:97": "TP-Link",
	"70:4F:57": "TP-Link",
	"98:DA:C4": "TP-Link",
	"A0:F3:C1": "TP-Link",

	// Netgear
	"00:09:5B": "Netgear",
	"00:0F:B5": "Netgear",
	"00:14:6C": "Netgear",
	"00:18:4D": "Netgear",
	"20:4E:7F": "Netgear",
	"84:1B:5E": "Netgear",
	"A0:04:60": "Netgear",
	"C0:3F:0E": "Netgear",
	"00:1E:2A": "Netgear",
	"00:24:B2": "Netgear",
	"00:26:F2": "Netgear",
	"10:DA:43": "Netgear",
	"28:80:88": "Netgear",

	// D-Link
	"00:05:5D": "D-Link",
	"00:0D:88": "D-Link",
	"00:15:E9": "D-Link",
	"00:17:9A": "D-Link",
	"1C:7E:E5": "D-Link",
	"28:10:7B": "D-Link",
	"00:1B:11": "D-Link",
	"00:1E:58": "D-Link",
	"00:22:B0": "D-Link",
	"14:D6:4D": "D-Link",
	"20:CF:30": "D-Link",

	// MikroTik
	"00:0C:42": "MikroTik",
	"48:8F:5A": "MikroTik",
	"64:D1:54": "MikroTik",
	"CC:2D:E0": "MikroTik",
	"D4:CA:6D": "MikroTik",
	"E4:8D:8C": "MikroTik",
	"08:55:31": "MikroTik",
	"18:FD:74": "MikroTik",
	"2C:C8:1B": "MikroTik",
	"4C:5E:0C": "MikroTik",
	"B8:69:F4": "MikroTik",

	// Huawei
	"00:18:82": "Huawei",
	"00:1E:10": "Huawei",
	"00:25:9E": "Huawei",
	"20:F3:A3": "Huawei",
	"70:54:F5": "Huawei",
	"E4:68:A3": "Huawei",
	"F8:4A:BF": "Huawei",
	"00:22:A1": "Huawei",
	"00:25:68": "Huawei",
	"08:19:A6": "Huawei",
	"10:47:80": "Huawei",

	// Storage & NAS
	"00:11:32": "Synology",
	"00:08:9B": "QNAP",
	"24:5E:BE": "QNAP",

	// Virtualization & Hypervisors
	"52:54:00": "QEMU / KVM",
	"BC:24:11": "Proxmox",
	"00:07:0E": "Realtek",
	"00:E0:4C": "Realtek",

	// PC Vendors
	"00:59:07": "Lenovo",
	"60:67:20": "Lenovo",
	"70:72:3C": "Lenovo",
	"AC:74:09": "Lenovo",
	"EC:8E:B5": "Lenovo",
	"00:15:F2": "Asus",
	"00:18:F3": "Asus",
	"04:D9:F5": "Asus",
	"08:60:6E": "Asus",
	"10:7B:44": "Asus",
	"2C:4D:54": "Asus",
	"70:4D:7B": "Asus",
	"AC:9E:17": "Asus",

	// Consumer Electronics & Smart Home
	"00:FC:8B": "Amazon Technologies",
	"38:F7:3D": "Amazon Technologies",
	"40:B4:CD": "Amazon Technologies",
	"68:37:E9": "Amazon Technologies",
	"74:75:48": "Amazon Technologies",
	"AC:63:BE": "Amazon Technologies",
	"F0:D2:F1": "Amazon Technologies",
	"FC:65:DE": "Amazon Technologies",
	"00:01:4A": "Sony",
	"00:04:1F": "Sony",
	"00:13:15": "Sony",
	"00:1A:80": "Sony",
	"28:0D:FC": "Sony",
	"70:9E:29": "Sony",
	"00:1C:62": "LG Electronics",
	"00:1E:75": "LG Electronics",
	"10:68:3F": "LG Electronics",
	"34:FC:EF": "LG Electronics",
	"64:99:5D": "LG Electronics",
	"00:0E:58": "Sonos",
	"48:A6:B8": "Sonos",
	"5C:AA:FD": "Sonos",
	"94:9F:3E": "Sonos",
	"B8:E9:37": "Sonos",
	"00:17:88": "Philips Signify Hue",
	"EC:B5:FA": "Philips Signify Hue",

	// Printers
	"00:00:85": "Canon",
	"00:1E:8F": "Canon",
	"18:0C:AC": "Canon",
	"70:85:C2": "Canon",
	"00:80:77": "Brother Industries",
	"30:05:5C": "Brother Industries",
	"40:49:0F": "Brother Industries",

	// Enterprise Security & Routing
	"00:09:0F": "Fortinet",
	"08:5B:0E": "Fortinet",
	"70:4C:A5": "Fortinet",
	"90:6C:AC": "Fortinet",
	"00:1C:73": "Arista Networks",
	"28:99:3A": "Arista Networks",
	"74:83:EF": "Arista Networks",
	"00:05:85": "Juniper Networks",
	"00:10:DB": "Juniper Networks",
	"00:12:1E": "Juniper Networks",
	"00:14:F6": "Juniper Networks",
	"00:19:E2": "Juniper Networks",
	"00:1D:B5": "Juniper Networks",
	"00:21:59": "Juniper Networks",
	"00:24:DC": "Juniper Networks",
	"00:26:88": "Juniper Networks",
	"84:18:88": "Juniper Networks",
}
