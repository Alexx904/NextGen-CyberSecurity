package validator

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"nextgen-cybersec/internal/domain"
)

// TargetInfo racchiude le informazioni sintattiche estratte dalla validazione
type TargetInfo struct {
	RawValue   string            `json:"raw_value"`
	CleanValue string            `json:"clean_value"`
	Type       domain.TargetType `json:"type"`
	IsIPv6     bool              `json:"is_ipv6"`
}

// ValidateTarget analizza l'input, applica i controlli di sicurezza e ne determina la tipologia
func ValidateTarget(input string) (*TargetInfo, error) {
	clean := strings.TrimSpace(input)
	if clean == "" {
		return nil, domain.ErrTargetInvalid
	}

	// NFR-SEC-01: Prevenzione Flag Injection & Command Injection
	// Un target non può mai iniziare con un trattino o contenere metacaratteri shell
	if strings.HasPrefix(clean, "-") || strings.HasPrefix(clean, "/") {
		return nil, domain.ErrTargetInvalid
	}

	// Rifiuto di spazi interni e caratteri di controllo o shell injection
	for _, ch := range clean {
		if ch <= 32 || ch >= 127 {
			return nil, domain.ErrTargetInvalid
		}
		switch ch {
		case ';', '&', '|', '$', '>', '<', '`', '\\', '"', '\'', '(', ')', '{', '}', '*', '?', '!':
			return nil, domain.ErrTargetInvalid
		}
	}

	// 1. Verifica se è un singolo IP (IPv4 o IPv6)
	if valid, isIPv6 := IsValidIP(clean); valid {
		return &TargetInfo{
			RawValue:   input,
			CleanValue: clean,
			Type:       domain.TargetTypeIP,
			IsIPv6:     isIPv6,
		}, nil
	}

	// 2. Verifica se è una notazione CIDR
	if valid, _, isIPv6 := IsValidCIDR(clean); valid {
		return &TargetInfo{
			RawValue:   input,
			CleanValue: clean,
			Type:       domain.TargetTypeCIDR,
			IsIPv6:     isIPv6,
		}, nil
	}

	// 3. Verifica se è un range di IP (es. 192.168.1.1-50 o 192.168.1.1-192.168.1.50)
	if valid, _, _ := IsValidRange(clean); valid {
		return &TargetInfo{
			RawValue:   input,
			CleanValue: clean,
			Type:       domain.TargetTypeRange,
			IsIPv6:     false,
		}, nil
	}

	// 4. Verifica se è un FQDN (Fully Qualified Domain Name)
	if IsValidFQDN(clean) {
		return &TargetInfo{
			RawValue:   input,
			CleanValue: strings.ToLower(clean),
			Type:       domain.TargetTypeFQDN,
			IsIPv6:     false,
		}, nil
	}

	return nil, domain.ErrTargetInvalid
}

// IsValidIP controlla se la stringa rappresenta un indirizzo IP valido
func IsValidIP(val string) (bool, bool) {
	ip := net.ParseIP(val)
	if ip == nil {
		return false, false
	}
	isIPv6 := strings.Contains(val, ":")
	return true, isIPv6
}

// IsValidCIDR controlla se la stringa rappresenta una subnet CIDR valida
func IsValidCIDR(val string) (bool, *net.IPNet, bool) {
	_, ipNet, err := net.ParseCIDR(val)
	if err != nil || ipNet == nil {
		return false, nil, false
	}
	isIPv6 := strings.Contains(val, ":")
	return true, ipNet, isIPv6
}

// IsValidRange controlla se la stringa rappresenta un intervallo di indirizzi IPv4
func IsValidRange(val string) (bool, net.IP, net.IP) {
	parts := strings.Split(val, "-")
	if len(parts) != 2 {
		return false, nil, nil
	}

	startStr := strings.TrimSpace(parts[0])
	endStr := strings.TrimSpace(parts[1])

	startIP := net.ParseIP(startStr)
	if startIP == nil || startIP.To4() == nil {
		return false, nil, nil
	}

	var endIP net.IP
	// Caso 1: Secondo elemento è un IP completo (es. 192.168.1.1-192.168.1.20)
	if fullEnd := net.ParseIP(endStr); fullEnd != nil && fullEnd.To4() != nil {
		endIP = fullEnd
	} else {
		// Caso 2: Secondo elemento è solo l'ultimo ottetto (es. 192.168.1.1-50)
		lastOctet, err := strconv.Atoi(endStr)
		if err != nil || lastOctet < 0 || lastOctet > 255 {
			return false, nil, nil
		}
		ipv4 := startIP.To4()
		endIP = net.IPv4(ipv4[0], ipv4[1], ipv4[2], byte(lastOctet))
	}

	// Verifica che startIP <= endIP
	if bytes.Compare(startIP.To4(), endIP.To4()) > 0 {
		return false, nil, nil
	}

	return true, startIP.To4(), endIP.To4()
}

// IsValidFQDN verifica la conformità del dominio a RFC 1123 / RFC 952
func IsValidFQDN(val string) bool {
	if len(val) == 0 || len(val) > 253 {
		return false
	}

	val = strings.TrimSuffix(val, ".")
	labels := strings.Split(val, ".")
	if len(labels) < 2 {
		return false // Richiede almeno dominio + TLD (es. scanme.nmap.org, example.com)
	}

	for i, label := range labels {
		if len(label) == 0 || len(label) > 63 {
			return false
		}
		if label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}

		// Ultimo label (TLD): deve essere puramente alfabetico di almeno 2 caratteri
		if i == len(labels)-1 {
			if len(label) < 2 {
				return false
			}
			for _, ch := range label {
				if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')) {
					return false
				}
			}
			continue
		}

		// Label intermedi: alfanumerici e trattino
		for _, ch := range label {
			if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '-') {
				return false
			}
		}
	}

	return true
}

// IPNetOverlaps verifica se due subnet CIDR hanno indirizzi in comune
func IPNetOverlaps(n1, n2 *net.IPNet) bool {
	return n1.Contains(n2.IP) || n2.Contains(n1.IP)
}

// ipInRange verifica se un IP si trova all'interno di un range [startIP, endIP]
func ipInRange(ip, startIP, endIP net.IP) bool {
	ip4 := ip.To4()
	start4 := startIP.To4()
	end4 := endIP.To4()
	if ip4 == nil || start4 == nil || end4 == nil {
		return false
	}
	return bytes.Compare(ip4, start4) >= 0 && bytes.Compare(ip4, end4) <= 0
}

// IsTargetBlacklisted verifica se il target specificato collide con una regola di blacklist
func IsTargetBlacklisted(targetValue string, blacklistedEntries []string) (bool, error) {
	targetInfo, err := ValidateTarget(targetValue)
	if err != nil {
		return false, fmt.Errorf("validazione target fallita: %w", err)
	}

	for _, entry := range blacklistedEntries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		// Caso FQDN
		if targetInfo.Type == domain.TargetTypeFQDN {
			targetLower := strings.ToLower(targetInfo.CleanValue)
			entryLower := strings.ToLower(entry)

			if targetLower == entryLower || strings.HasSuffix(targetLower, "."+entryLower) {
				return true, nil
			}
			continue
		}

		// Regola blacklist è un singolo IP
		if bIP := net.ParseIP(entry); bIP != nil {
			switch targetInfo.Type {
			case domain.TargetTypeIP:
				tIP := net.ParseIP(targetInfo.CleanValue)
				if tIP != nil && tIP.Equal(bIP) {
					return true, nil
				}
			case domain.TargetTypeCIDR:
				_, tNet, _ := net.ParseCIDR(targetInfo.CleanValue)
				if tNet != nil && tNet.Contains(bIP) {
					return true, nil
				}
			case domain.TargetTypeRange:
				_, startIP, endIP := IsValidRange(targetInfo.CleanValue)
				if ipInRange(bIP, startIP, endIP) {
					return true, nil
				}
			}
			continue
		}

		// Regola blacklist è una subnet CIDR
		if _, bNet, errCIDR := net.ParseCIDR(entry); errCIDR == nil && bNet != nil {
			switch targetInfo.Type {
			case domain.TargetTypeIP:
				tIP := net.ParseIP(targetInfo.CleanValue)
				if tIP != nil && bNet.Contains(tIP) {
					return true, nil
				}
			case domain.TargetTypeCIDR:
				_, tNet, _ := net.ParseCIDR(targetInfo.CleanValue)
				if tNet != nil && IPNetOverlaps(tNet, bNet) {
					return true, nil
				}
			case domain.TargetTypeRange:
				_, startIP, endIP := IsValidRange(targetInfo.CleanValue)
				if bNet.Contains(startIP) || bNet.Contains(endIP) {
					return true, nil
				}
			}
			continue
		}

		// Regola blacklist è un Range
		if isR, bStart, bEnd := IsValidRange(entry); isR {
			switch targetInfo.Type {
			case domain.TargetTypeIP:
				tIP := net.ParseIP(targetInfo.CleanValue)
				if ipInRange(tIP, bStart, bEnd) {
					return true, nil
				}
			case domain.TargetTypeRange:
				_, tStart, tEnd := IsValidRange(targetInfo.CleanValue)
				// Overlap tra due range
				if !(bytes.Compare(tEnd, bStart) < 0 || bytes.Compare(tStart, bEnd) > 0) {
					return true, nil
				}
			}
			continue
		}
	}

	return false, nil
}

// RestrictedSubnetsCIDR elenca le subnet riservate e dell'infrastruttura interna (NFR-SEC-01, FR-01.4)
var RestrictedSubnetsCIDR = []string{
	"0.0.0.0/8",       // Indirizzi rete corrente / host locale
	"127.0.0.0/8",     // IPv4 Loopback
	"::1/128",         // IPv6 Loopback
	"169.254.0.0/16",  // IPv4 Link-Local / Cloud Instance Metadata Service (AWS/GCP/Azure 169.254.169.254)
	"fe80::/10",       // IPv6 Link-Local
	"172.17.0.0/16",   // Docker default bridge
	"172.18.0.0/16",   // Docker compose networks
	"172.19.0.0/16",
	"172.20.0.0/16",
}

// RestrictedHostnames elenca gli hostname interni riservati dei container della piattaforma
var RestrictedHostnames = []string{
	"localhost",
	"postgres",
	"nextgen-postgres",
	"backend",
	"nextgen-backend",
	"frontend",
	"nextgen-frontend",
}

// CheckTargetScope verifica che il target non ricada nell'infrastruttura interna protetta
// e non ecceda la dimensione massima ammessa (FR-01.4).
func CheckTargetScope(targetValue string, resolveDNS bool) error {
	// Consente lo sblocco per sviluppo locale o testing autorizzato se espressamente configurato via ambiente
	if strings.EqualFold(os.Getenv("ALLOW_INTERNAL_TARGETS"), "true") || strings.EqualFold(os.Getenv("DEV_ALLOW_LOCALHOST"), "true") {
		return nil
	}

	clean := strings.ToLower(strings.TrimSpace(targetValue))
	for _, forbiddenHost := range RestrictedHostnames {
		if clean == forbiddenHost || strings.HasSuffix(clean, "."+forbiddenHost) {
			return domain.ErrTargetInternalRestricted
		}
	}

	targetInfo, err := ValidateTarget(targetValue)
	if err != nil {
		return err
	}

	// 1. Controllo dimensione massima ammissibile (un /8 non deve partire)
	if targetInfo.Type == domain.TargetTypeCIDR {
		_, ipNet, isIPv6 := IsValidCIDR(targetInfo.CleanValue)
		if ipNet != nil {
			ones, _ := ipNet.Mask.Size()
			if !isIPv6 && ones < 16 {
				return fmt.Errorf("%w: prefisso /%d inferiore al minimo consentito /16 (max 65536 host)", domain.ErrTargetScopeTooBroad, ones)
			}
			if isIPv6 && ones < 64 {
				return fmt.Errorf("%w: prefisso IPv6 /%d inferiore al minimo consentito /64", domain.ErrTargetScopeTooBroad, ones)
			}
		}
	}

	// 2. Controllo contro hostname riservati
	if targetInfo.Type == domain.TargetTypeFQDN {
		lowerFQDN := strings.ToLower(targetInfo.CleanValue)
		for _, forbiddenHost := range RestrictedHostnames {
			if lowerFQDN == forbiddenHost || strings.HasSuffix(lowerFQDN, "."+forbiddenHost) {
				return domain.ErrTargetInternalRestricted
			}
		}

		if strings.HasSuffix(lowerFQDN, ".internal") || strings.HasSuffix(lowerFQDN, ".local") {
			return domain.ErrTargetInternalRestricted
		}

		// Controllo anti-DNS Rebinding se abilitata la risoluzione
		if resolveDNS {
			resolvedIPs, lookupErr := net.LookupIP(targetInfo.CleanValue)
			if lookupErr == nil {
				for _, ip := range resolvedIPs {
					if isIPRestricted(ip) {
						return fmt.Errorf("%w: dominio '%s' risolve su IP riservato %s", domain.ErrTargetInternalRestricted, targetInfo.CleanValue, ip.String())
					}
				}
			}
		}
		return nil
	}

	// 3. Controllo IP / CIDR / Range contro le subnet riservate
	isRestricted, err := IsTargetBlacklisted(targetInfo.CleanValue, RestrictedSubnetsCIDR)
	if err != nil {
		return err
	}
	if isRestricted {
		return domain.ErrTargetInternalRestricted
	}

	return nil
}

// isIPRestricted controlla se un IP net.IP appartiene alle subnet protette
func isIPRestricted(ip net.IP) bool {
	for _, cidr := range RestrictedSubnetsCIDR {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err == nil && ipNet != nil && ipNet.Contains(ip) {
			return true
		}
	}
	return false
}

