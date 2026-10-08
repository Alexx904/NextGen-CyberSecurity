package scanner

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/pkg/validator"
)

// ProfileType definisce la tipologia di profilo di scansione preconfigurato
type ProfileType string

const (
	ProfileFastDiscovery ProfileType = "Fast Discovery"
	ProfileStandardScan  ProfileType = "Standard Service Scan"
	ProfileFullScan      ProfileType = "Full Deep Scan"
)

// ProfileConfig racchiude le specifiche operative di un profilo di scansione Nmap
type ProfileConfig struct {
	Name           string        `json:"name"`
	Args           []string      `json:"args"`
	Timeout        time.Duration `json:"timeout"`
	Description    string        `json:"description"`
}

// GetDefaultProfiles restituisce i profili di scansione di default preimpostati nel sistema
func GetDefaultProfiles() map[ProfileType]ProfileConfig {
	return map[ProfileType]ProfileConfig{
		ProfileFastDiscovery: {
			Name:        string(ProfileFastDiscovery),
			Args:        []string{"-T4", "-F"},
			Timeout:     15 * time.Minute, // 900 secondi
			Description: "Discovery rapida degli host attivi e delle top 100 porte TCP",
		},
		ProfileStandardScan: {
			Name:        string(ProfileStandardScan),
			Args:        []string{"-sV", "-T4"},
			Timeout:     60 * time.Minute, // 3600 secondi
			Description: "Scansione standard delle top 1000 porte TCP con rilevamento versioni dei servizi",
		},
		ProfileFullScan: {
			Name:        string(ProfileFullScan),
			Args:        []string{"-p-", "-sV", "-O", "-T4"},
			Timeout:     240 * time.Minute, // 14400 secondi
			Description: "Scansione completa su 65535 porte TCP, OS detection e rilevamento servizi approfondito",
		},
	}
}

// Regex per convalidare argomenti con parametri (es: -p80,443 o --top-ports 100 o --max-retries 2)
var (
	validPortArgRegex       = regexp.MustCompile(`^-p([0-9]{1,5}(-[0-9]{1,5})?,?)+$|^-p-$`)
	validTimingArgRegex     = regexp.MustCompile(`^-T[0-5]$`)
	validTopPortsRegex      = regexp.MustCompile(`^[0-9]{1,5}$`)
	validRateRegex          = regexp.MustCompile(`^[0-9]+$`)
	validVersionIntRegex    = regexp.MustCompile(`^[0-9]$`)
)

// DisallowedFlags elenca i flag Nmap severamente vietati per ragioni di sicurezza e conformità architetturale (NFR-SEC-01)
var DisallowedFlags = map[string]bool{
	// Disabilitazione esecuzione script esterni o NSE arbitrari
	"--script":          true,
	"--script-args":     true,
	"--script-help":     true,
	"--script-trace":    true,
	"--script-updatedb": true,

	// Disabilitazione reindirizzamento output su file su disco (l'output deve transitare via stdout XML)
	"-oN":               true,
	"-oX":               true,
	"-oS":               true,
	"-oG":               true,
	"-oA":               true,
	"--stylesheet":      true,
	"--no-stylesheet":   true,

	// Disabilitazione caricamento target da file locali
	"-iL":               true,
	"-iR":               true,
	"--resume":          true,

	// Disabilitazione modalità interattive o privileged bypass
	"--interactive":     true,
	"--privileged":      true,
	"--unprivileged":    true,
	"--datadir":         true,
	"--servicedb":       true,
	"--versiondb":       true,
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
	case "-F", "-sV", "-O", "-Pn", "-sn", "-sS", "-sT", "-sU", "-sA", "-sN", "-sF", "-sX",
		"-n", "-R", "-v", "-vv", "-q", "-d", "-6", "--version-light", "--version-all",
		"--osscan-limit", "--osscan-guess", "--open", "--packet-trace":
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

	return result, nil
}

// BuildSafeArguments convalida il target e genera la slice ordinata di argomenti per l'invocazione diretta di Nmap
// Output generato: ["-oX", "-", ...safeArgs, target]
func BuildSafeArguments(profileArgs []string, target string) ([]string, error) {
	// 1. Validazione preventiva del bersaglio (NFR-SEC-01)
	targetInfo, err := validator.ValidateTarget(target)
	if err != nil {
		return nil, fmt.Errorf("bersaglio non valido per la scansione: %w", err)
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

	// 4. Costruzione della slice finale: output XML forzato su stdout, argomenti del profilo e target finale
	finalArgs := make([]string, 0, len(validatedArgs)+3)
	finalArgs = append(finalArgs, "-oX", "-") // Output XML nativo su stdout
	finalArgs = append(finalArgs, validatedArgs...)
	finalArgs = append(finalArgs, targetInfo.CleanValue)

	return finalArgs, nil
}
