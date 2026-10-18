package scanner

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/pkg/validator"
)

// ProfileType definisce la tipologia di profilo di scansione preconfigurato
type ProfileType string

const (
	ProfileAllInOne360      ProfileType = "All-in-One 360° Comprehensive Audit"
	ProfileFastDiscovery    ProfileType = "Fast Discovery"
	ProfileStandardScan     ProfileType = "Standard Service Scan"
	ProfileFullScan         ProfileType = "Full Deep Scan"
	ProfileLANDiscovery     ProfileType = "LAN Discovery"
	ProfileStealthSYN       ProfileType = "Stealth SYN Penetration Audit"
	ProfileFirewallBypass   ProfileType = "Firewall & Filter Bypass Check"
	ProfileSystemFlawsOS    ProfileType = "OS & System Daemon Flaws Audit"
	ProfileCriticalServices ProfileType = "Critical Services & RCE Surface"
	ProfileWebSurface       ProfileType = "Web & API Surface Audit"
	ProfileFullWebStack     ProfileType = "Full Web Stack Deep Versioning"
	ProfileFastPingSweep    ProfileType = "Fast Ping & Host Sweep (LAN)"
	ProfileLegacyProtocols  ProfileType = "Legacy & Insecure Protocol Check"
	ProfileSpecializedUDP   ProfileType = "Specialized UDP Critical Services"
)

// ProfileConfig racchiude le specifiche operative di un profilo di scansione Nmap
type ProfileConfig struct {
	OnHost       HostCallback               `json:"-"`
	OnProgress   func(TaskProgress)         `json:"-"`
	Name         string                     `json:"name"`
	Args         []string                   `json:"args"`
	Timeout      time.Duration              `json:"timeout"`
	Description  string                     `json:"description"`
	Category     domain.ScanProfileCategory `json:"category"`
	Invasiveness domain.InvasivenessLevel   `json:"invasiveness"`
}

// GetDefaultProfiles restituisce i profili di scansione di default preimpostati nel sistema
func GetDefaultProfiles() map[ProfileType]ProfileConfig {
	return map[ProfileType]ProfileConfig{
		ProfileAllInOne360: {
			Name:         string(ProfileAllInOne360),
			Args:         []string{"-sS", "-sV", "-O", "--script=vuln", "-T4", "--open"},
			Timeout:      90 * time.Minute,
			Description:  "Scansione generalista completa a 360°: mappatura host vivi, scansione porte standard, identificazione versioni demoni (sV), fingerprinting OS (O) e controllo vulnerabilità note (--script=vuln)",
			Category:     domain.ProfileCategoryAllInOne,
			Invasiveness: domain.InvasivenessIntrusive,
		},
		ProfileFastDiscovery: {
			Name:         string(ProfileFastDiscovery),
			Args:         []string{"-T4", "-F"},
			Timeout:      15 * time.Minute, // 900 secondi
			Description:  "Discovery rapida degli host attivi e delle top 100 porte TCP",
			Category:     domain.ProfileCategoryDiscovery,
			Invasiveness: domain.InvasivenessSafe,
		},
		ProfileStandardScan: {
			Name:         string(ProfileStandardScan),
			Args:         []string{"-sV", "-T4"},
			Timeout:      60 * time.Minute, // 3600 secondi
			Description:  "Scansione standard delle top 1000 porte TCP con rilevamento versioni dei servizi",
			Category:     domain.ProfileCategorySystemFlaws,
			Invasiveness: domain.InvasivenessSafe,
		},
		ProfileFullScan: {
			Name:         string(ProfileFullScan),
			Args:         []string{"-p-", "-sV", "-O", "-T4"},
			Timeout:      240 * time.Minute, // 14400 secondi
			Description:  "Scansione completa su 65535 porte TCP, OS detection e rilevamento servizi approfondito",
			Category:     domain.ProfileCategoryPenTest,
			Invasiveness: domain.InvasivenessIntrusive,
		},
		ProfileLANDiscovery: {
			Name:         string(ProfileLANDiscovery),
			Args:         []string{"-sn", "-PE", "-PS21,22,80,443,445,8080,8443,3389,53,8000", "-PU53,137,161,5353", "-T4"},
			Timeout:      2 * time.Minute, // 120 secondi
			Description:  "Ricognizione LAN ultra-veloce con ping sweep affidabile, probe ICMP/SYN/UDP su porte standard e risoluzione reverse DNS",
			Category:     domain.ProfileCategoryDiscovery,
			Invasiveness: domain.InvasivenessSafe,
		},
		ProfileStealthSYN: {
			Name:         string(ProfileStealthSYN),
			Args:         []string{"-sS", "-Pn", "-T2", "--open"},
			Timeout:      60 * time.Minute,
			Description:  "Ricognizione stealth a basso rumore con pacchetti SYN senza handshake TCP completo, elusione baseline firewall e solo porte aperte",
			Category:     domain.ProfileCategoryPenTest,
			Invasiveness: domain.InvasivenessSafe,
		},
		ProfileFirewallBypass: {
			Name:         string(ProfileFirewallBypass),
			Args:         []string{"-sA", "-Pn", "-T3"},
			Timeout:      30 * time.Minute,
			Description:  "Mappatura filtri firewall perimetrali e apparati IDS con pacchetti TCP ACK per verificare se i firewall sono stateful o stateless",
			Category:     domain.ProfileCategoryPenTest,
			Invasiveness: domain.InvasivenessIntrusive,
		},
		ProfileSystemFlawsOS: {
			Name:         string(ProfileSystemFlawsOS),
			Args:         []string{"-sV", "-O", "--version-intensity", "7", "--script=vuln", "-T4"},
			Timeout:      60 * time.Minute,
			Description:  "Rilevamento dettagliato del sistema operativo, fingerprinting demoni di sistema e script attivi di vulnerabilità Nmap (--script=vuln)",
			Category:     domain.ProfileCategorySystemFlaws,
			Invasiveness: domain.InvasivenessIntrusive,
		},
		ProfileCriticalServices: {
			Name:         string(ProfileCriticalServices),
			Args:         []string{"-sV", "-p", "21,22,23,25,53,80,110,135,139,143,443,445,1433,1521,3306,3389,5432,5900,6379,8080,8443", "--script=vuln", "-T4"},
			Timeout:      30 * time.Minute,
			Description:  "Audit mirato sui vettori critici più colpite da exploit remoti e vulnerabilità note RCE/PrivEsc (--script=vuln)",
			Category:     domain.ProfileCategorySystemFlaws,
			Invasiveness: domain.InvasivenessIntrusive,
		},
		ProfileWebSurface: {
			Name:         string(ProfileWebSurface),
			Args:         []string{"-sV", "-p", "80,443,8000,8080,8443,8888,3000,5000,9000", "-T4", "--open"},
			Timeout:      20 * time.Minute,
			Description:  "Rilevamento mirato di server HTTP/HTTPS, API REST, backend microservizi e reverse proxy applicativi (Nginx, Apache, Traefik)",
			Category:     domain.ProfileCategoryWebApp,
			Invasiveness: domain.InvasivenessSafe,
		},
		ProfileFullWebStack: {
			Name:         string(ProfileFullWebStack),
			Args:         []string{"-sV", "--version-all", "-p", "80,443,8008,8080,8443,9090,9443", "-T4"},
			Timeout:      40 * time.Minute,
			Description:  "Versioning approfondito con interrogazione intensiva dei banner HTTP e intestazioni TLS per scovare componenti obsoleti",
			Category:     domain.ProfileCategoryWebApp,
			Invasiveness: domain.InvasivenessSafe,
		},
		ProfileFastPingSweep: {
			Name:         string(ProfileFastPingSweep),
			Args:         []string{"-sn", "-PE", "-PS22,80,443,445", "-T4"},
			Timeout:      2 * time.Minute,
			Description:  "Ricognizione LAN ultra-rapida con ping sweep, ICMP echo e probe SYN senza port scan per mappare tutti gli host accesi e resolver reverse DNS",
			Category:     domain.ProfileCategoryDiscovery,
			Invasiveness: domain.InvasivenessSafe,
		},
		ProfileLegacyProtocols: {
			Name:         string(ProfileLegacyProtocols),
			Args:         []string{"-sV", "-p", "21,23,79,110,143,512,513,514", "-T3"},
			Timeout:      15 * time.Minute,
			Description:  "Verifica conformità e sicurezza per rilevare servizi in chiaro e protocolli legacy obsoleti (Telnet, FTP, Finger, POP3, IMAP, Rexec, Rsh, Rlogin)",
			Category:     domain.ProfileCategoryCompliance,
			Invasiveness: domain.InvasivenessSafe,
		},
		ProfileSpecializedUDP: {
			Name:         string(ProfileSpecializedUDP),
			Args:         []string{"-sU", "-p", "U:53,69,123,161,514,623,5353", "-T4"},
			Timeout:      20 * time.Minute,
			Description:  "Audit mirato sui servizi UDP ad alto rischio: DNS (53), TFTP (69), NTP (123), SNMP (161), Syslog (514), IPMI (623) e mDNS (5353)",
			Category:     domain.ProfileCategoryPenTest,
			Invasiveness: domain.InvasivenessSafe,
		},
	}
}

// Regex per convalidare argomenti con parametri (es: -p80,443 o -p U:53,69 o --top-ports 100 o --max-retries 2)
var (
	validPortArgRegex    = regexp.MustCompile(`^-p(([TU]:)?[0-9]{1,5}(-[0-9]{1,5})?,?)+$|^-p-$`)
	validTimingArgRegex  = regexp.MustCompile(`^-T[0-5]$`)
	validTopPortsRegex   = regexp.MustCompile(`^[0-9]{1,5}$`)
	validRateRegex       = regexp.MustCompile(`^[0-9]+$`)
	validVersionIntRegex = regexp.MustCompile(`^[0-9]$`)
	validSYNArgRegex     = regexp.MustCompile(`^-PS([0-9]{1,5},?)+$`)
	validUDPArgRegex     = regexp.MustCompile(`^-PU([0-9]{1,5},?)+$`)
)

// DisallowedFlags elenca i flag Nmap severamente vietati per ragioni di sicurezza e conformità architetturale (NFR-SEC-01)
var DisallowedFlags = map[string]bool{
	// Disabilitazione flag ausiliari script o parametri non convalidati
	"--script-args":     true,
	"--script-help":     true,
	"--script-trace":    true,
	"--script-updatedb": true,

	// Disabilitazione reindirizzamento output su file su disco (l'output deve transitare via stdout XML)
	"-oN":             true,
	"-oX":             true,
	"-oS":             true,
	"-oG":             true,
	"-oA":             true,
	"--stylesheet":    true,
	"--no-stylesheet": true,

	// Disabilitazione caricamento target da file locali
	"-iL":      true,
	"-iR":      true,
	"--resume": true,

	// Disabilitazione modalità interattive o privileged bypass
	"--interactive":  true,
	"--privileged":   true,
	"--unprivileged": true,
	"--datadir":      true,
	"--servicedb":    true,
	"--versiondb":    true,
}

// IsSafeArgument verifica se un singolo token/flag Nmap rispetta le policy di sicurezza
func IsSafeArgument(arg string) bool {
	argClean := strings.TrimSpace(arg)
	if argClean == "" {
		return false
	}

	// Verifica presenza di metacaratteri shell malevoli
	for _, ch := range argClean {
		switch ch {
		case ';', '&', '|', '$', '>', '<', '`', '\\', '"', '\'', '(', ')', '{', '}', '*', '?', '!':
			return false
		}
	}

	// Controlla contro la blacklist dei flag vietati
	for forbidden := range DisallowedFlags {
		if argClean == forbidden || strings.HasPrefix(argClean, forbidden+"=") {
			return false
		}
	}

	// Flag esatti consentiti
	switch argClean {
	case "-F", "-sV", "-O", "-Pn", "-sn", "-PR", "-PE", "-PP", "-sS", "-sT", "-sU", "-sA", "-sN", "-sF", "-sX",
		"-n", "-R", "-v", "-vv", "-q", "-d", "-6", "--version-light", "--version-all",
		"--osscan-limit", "--osscan-guess", "--open", "--packet-trace",
		"--script=vuln", "--script=vulners":
		return true
	}

	// Validazione timing -T0 .. -T5
	if validTimingArgRegex.MatchString(argClean) {
		return true
	}

	// Validazione specifica porte (es: -p-, -p80, -p80,443, -p1-1000)
	if validPortArgRegex.MatchString(argClean) {
		return true
	}

	// Validazione probe SYN ping (es: -PS22,80,443)
	if validSYNArgRegex.MatchString(argClean) {
		return true
	}

	// Validazione probe UDP ping (es: -PU53,137)
	if validUDPArgRegex.MatchString(argClean) {
		return true
	}

	return false
}

// ParseProfileArgs converte una stringa di argomenti Nmap (es. "-sV -T4 -p 80,443") in una slice sicura
func ParseProfileArgs(argsStr string) ([]string, error) {
	fields := strings.Fields(strings.TrimSpace(argsStr))
	if len(fields) == 0 {
		return nil, nil
	}

	result := make([]string, 0, len(fields))
	for i := 0; i < len(fields); i++ {
		token := fields[i]

		// Gestione flag composti con parametro successivo (es: -p 80,443 oppure --top-ports 100)
		if token == "-p" {
			if i+1 < len(fields) {
				i++
				combined := "-p" + fields[i]
				if !validPortArgRegex.MatchString(combined) {
					return nil, fmt.Errorf("%w: porta non valida '%s'", domain.ErrDisallowedScanOption, fields[i])
				}
				result = append(result, combined)
				continue
			}
			return nil, fmt.Errorf("%w: parametro mancante per flag -p", domain.ErrDisallowedScanOption)
		}

		if token == "--script" {
			if i+1 < len(fields) {
				scriptCat := fields[i+1]
				if scriptCat == "vuln" || scriptCat == "vulners" {
					result = append(result, "--script="+scriptCat)
					i++
					continue
				}
			}
			return nil, fmt.Errorf("%w: categoria script non consentita per %s", domain.ErrDisallowedScanOption, token)
		}

		if token == "--script=vuln" || token == "--script=vulners" {
			result = append(result, token)
			continue
		}

		if token == "--top-ports" {
			if i+1 < len(fields) && validTopPortsRegex.MatchString(fields[i+1]) {
				result = append(result, token, fields[i+1])
				i++
				continue
			}
			return nil, fmt.Errorf("%w: valore non valido per --top-ports", domain.ErrDisallowedScanOption)
		}

		if token == "--max-retries" || token == "--min-rate" || token == "--max-rate" {
			if i+1 < len(fields) && validRateRegex.MatchString(fields[i+1]) {
				result = append(result, token, fields[i+1])
				i++
				continue
			}
			return nil, fmt.Errorf("%w: valore numerico atteso per %s", domain.ErrDisallowedScanOption, token)
		}

		if token == "--version-intensity" {
			if i+1 < len(fields) && validVersionIntRegex.MatchString(fields[i+1]) {
				result = append(result, token, fields[i+1])
				i++
				continue
			}
			return nil, fmt.Errorf("%w: intensità versione non valida per %s (0-9)", domain.ErrDisallowedScanOption, token)
		}

		if !IsSafeArgument(token) {
			return nil, fmt.Errorf("%w: flag '%s'", domain.ErrDisallowedScanOption, token)
		}
		result = append(result, token)
	}

	if err := validateScanSemantics(result); err != nil {
		return nil, err
	}
	return result, nil
}

// validateScanSemantics rejects configurations that Nmap cannot execute coherently.
func validateScanSemantics(args []string) error {
	flags := make(map[string]bool)
	tcpTypes, portSelectors := 0, 0
	invalid := func(reason string) error { return fmt.Errorf("%w: %s", domain.ErrDisallowedScanOption, reason) }
	for i, arg := range args {
		flags[arg] = true
		switch arg {
		case "-sS", "-sT", "-sA", "-sN", "-sF", "-sX":
			tcpTypes++
		case "-F":
			portSelectors++
		case "--top-ports":
			portSelectors++
			n, _ := strconv.Atoi(args[i+1]) // ParseProfileArgs already checked arity and syntax.
			if n < 1 || n > 65535 {
				return invalid("--top-ports richiede un valore tra 1 e 65535")
			}
		}
		if strings.HasPrefix(arg, "-p") {
			portSelectors++
			if arg == "-p-" {
				continue
			}
			for _, entry := range strings.Split(strings.TrimPrefix(arg, "-p"), ",") {
				entry = strings.TrimPrefix(strings.TrimPrefix(entry, "T:"), "U:")
				bounds := strings.Split(entry, "-")
				start, err := strconv.Atoi(bounds[0])
				if err != nil || start < 0 || start > 65535 {
					return invalid("porta fuori intervallo 0-65535")
				}
				if len(bounds) == 2 {
					end, err := strconv.Atoi(bounds[1])
					if err != nil || end < start || end > 65535 {
						return invalid("intervallo porte non valido")
					}
				}
			}
		}
	}
	if tcpTypes > 1 || portSelectors > 1 {
		return invalid("scegli una sola tecnica TCP e un solo selettore porte")
	}
	if flags["-sn"] && (tcpTypes > 0 || portSelectors > 0 || flags["-sU"] || flags["-sV"] || flags["-O"]) {
		return invalid("-sn esegue solo host discovery: rimuovi opzioni di port scan, versioni e OS")
	}
	if flags["-sA"] && (flags["--open"] || flags["-sV"] || flags["-O"]) {
		return invalid("ACK misura filtered/unfiltered: incompatibile con --open, -sV e -O")
	}
	return nil
}

// ClassifyInvasiveness is a conservative operational label, never a safety guarantee.
func ClassifyInvasiveness(args []string) domain.InvasivenessLevel {
	for _, arg := range args {
		switch arg {
		case "--script=vuln", "--script=vulners", "-p-", "-T4", "-T5", "--version-all", "-sA", "--min-rate":
			return domain.InvasivenessIntrusive
		}
		if arg == "-p1-65535" || arg == "-p0-65535" {
			return domain.InvasivenessIntrusive
		}
	}
	return domain.InvasivenessSafe
}

// BuildSafeArguments convalida il target e genera la slice ordinata di argomenti per l'invocazione diretta di Nmap
// Output generato: ["-oX", "-", ...safeArgs, "--", target]
func BuildSafeArguments(profileArgs []string, target string) ([]string, error) {
	// 1. Validazione preventiva del bersaglio (NFR-SEC-01)
	targetInfo, err := validator.ValidateTarget(target)
	if err != nil {
		return nil, fmt.Errorf("bersaglio non valido per la scansione: %w", err)
	}

	// 1.1 Controllo perimetro di sicurezza interno e deny-list (FR-01.4)
	if err := validator.CheckTargetScope(target, false); err != nil {
		return nil, fmt.Errorf("perimetro di scansione non consentito: %w", err)
	}

	// 2. Controllo di sicurezza su tutti gli argomenti forniti
	var validatedArgs []string
	if len(profileArgs) > 0 {
		// Se gli argomenti sono passati come stringhe singole o aggregate, li analizziamo
		combined := strings.Join(profileArgs, " ")
		parsed, parseErr := ParseProfileArgs(combined)
		if parseErr != nil {
			return nil, parseErr
		}
		validatedArgs = parsed
	}

	// 3. Se il target è IPv6 e il flag -6 non è presente, aggiungiamolo automaticamente
	if targetInfo.IsIPv6 {
		hasIPv6 := false
		for _, a := range validatedArgs {
			if a == "-6" {
				hasIPv6 = true
				break
			}
		}
		if !hasIPv6 {
			validatedArgs = append([]string{"-6"}, validatedArgs...)
		}
	}

	// 4. Costruzione della slice finale: output XML forzato su stdout, argomenti del profilo,
	// delimitatore standard POSIX '--' contro flag injection ostile e target finale
	finalArgs := make([]string, 0, len(validatedArgs)+4)
	finalArgs = append(finalArgs, "-oX", "-") // Output XML nativo su stdout
	finalArgs = append(finalArgs, validatedArgs...)
	finalArgs = append(finalArgs, "--") // Delimitatore di sicurezza POSIX
	finalArgs = append(finalArgs, targetInfo.CleanValue)

	return finalArgs, nil
}
