package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/scanner"
)

// Correlator definisce l'interfaccia del motore di correlazione vulnerabilità
type Correlator interface {
	CorrelateService(ctx context.Context, s *scanner.ParsedService) ([]*domain.CVEEntry, error)
	CorrelatePort(ctx context.Context, scanID, tenantID uuid.UUID, host scanner.ParsedHost, port scanner.ParsedPort) ([]*domain.Finding, error)
}

// CVECorrelator implementa il motore di correlazione locale tra servizi rilevati e CVE (NFR-REL-03)
type CVECorrelator struct {
	cveRepo domain.CVERepository
}

// NewCVECorrelator inizializza una nuova istanza di CVECorrelator
func NewCVECorrelator(cveRepo domain.CVERepository) *CVECorrelator {
	return &CVECorrelator{cveRepo: cveRepo}
}

// NormalizeCPETo23 converte un identificatore CPE 2.2 o incompleto nel formato standard CPE 2.3 URI
// Esempio: "cpe:/a:apache:http_server:2.4.49" -> "cpe:2.3:a:apache:http_server:2.4.49:*:*:*:*:*:*:*"
func NormalizeCPETo23(cpe string) string {
	cpe = strings.TrimSpace(cpe)
	if cpe == "" {
		return ""
	}

	// Se è già un URI CPE 2.3
	if strings.HasPrefix(cpe, "cpe:2.3:") {
		parts := strings.Split(cpe, ":")
		if len(parts) >= 13 {
			return cpe
		}
		// Pad a 13 elementi (cpe + 2.3 + 11 attributi standard)
		padded := make([]string, 13)
		copy(padded, parts)
		for i := len(parts); i < 13; i++ {
			padded[i] = "*"
		}
		return strings.Join(padded, ":")
	}

	// Se è in formato CPE 2.2 (prefisso "cpe:/")
	if strings.HasPrefix(cpe, "cpe:/") {
		trimmed := strings.TrimPrefix(cpe, "cpe:/")
		parts := strings.Split(trimmed, ":")

		attrs := make([]string, 11)
		for i := 0; i < 11; i++ {
			if i < len(parts) && parts[i] != "" {
				attrs[i] = parts[i]
			} else {
				attrs[i] = "*"
			}
		}

		return "cpe:2.3:" + strings.Join(attrs, ":")
	}

	return cpe
}

// GenerateRemediation formula una raccomandazione di sicurezza prescrittiva in funzione della gravità e del componente
func GenerateRemediation(serviceName, version, cveID string, severity domain.SeverityLevel) string {
	var target string
	if serviceName != "" {
		target = serviceName
		if version != "" {
			target += " (versione " + version + ")"
		}
	} else {
		target = "il servizio rilevato"
	}

	switch severity {
	case domain.SeverityCritical:
		return fmt.Sprintf("URGENTE: Aggiornare tempestivamente %s all'ultima versione di sicurezza patchata o disabilitare il servizio esposto per mitigare la vulnerabilità critica %s. Applicare restrizioni firewall immediate sulla porta esposta.", target, cveID)
	case domain.SeverityHigh:
		return fmt.Sprintf("Aggiornare %s alla versione supportata più recente per mitigare la vulnerabilità nota %s. Isolare o segmentare l'accesso alla porta se l'esposizione non è strettamente necessaria.", target, cveID)
	case domain.SeverityMedium:
		return fmt.Sprintf("Pianificare l'aggiornamento di %s per correggere la vulnerabilità %s. Verificare la configurazione del servizio e limitare i permessi di esecuzione.", target, cveID)
	default:
		return fmt.Sprintf("Mantenere %s aggiornato e monitorare i bollettini di sicurezza del vendor relativi a %s.", target, cveID)
	}
}

// CorrelateService cerca e restituisce tutte le voci CVE locali applicabili a un servizio rilevato
func (c *CVECorrelator) CorrelateService(ctx context.Context, s *scanner.ParsedService) ([]*domain.CVEEntry, error) {
	if s == nil || c.cveRepo == nil {
		return nil, nil
	}

	cveMap := make(map[string]*domain.CVEEntry)

	// 1. Ricerca tramite CPE specificati nel report Nmap
	for _, rawCPE := range s.CPEs {
		cpe23 := NormalizeCPETo23(rawCPE)
		if cpe23 == "" {
			continue
		}

		// A. Match esatto via GIN Index (@>)
		entries, err := c.cveRepo.SearchByCPEURI(ctx, cpe23)
		if err == nil {
			for _, e := range entries {
				cveMap[e.CVEID] = e
			}
		}

		// B. Se non abbiamo riscontri esatti, tentiamo il matching per prefisso di vendor/prodotto
		if len(entries) == 0 {
			parts := strings.Split(cpe23, ":")
			if len(parts) >= 5 {
				prefix := strings.Join(parts[:5], ":") // cpe:2.3:<part>:<vendor>:<product>
				targetVer := ""
				if len(parts) >= 6 && parts[5] != "*" {
					targetVer = parts[5]
				} else if s.Version != "" {
					targetVer = s.Version
				}

				prefixEntries, err := c.cveRepo.SearchByCPE(ctx, prefix, targetVer)
				if err == nil {
					for _, e := range prefixEntries {
						cveMap[e.CVEID] = e
					}
				}
			}
		}
	}

	// 2. Fallback euristico: se non vi sono CPE o non hanno prodotto match, usiamo Product e Version
	if len(cveMap) == 0 && s.Product != "" && s.Version != "" {
		// Ricerca generica per nome prodotto
		filter := domain.CVEFilter{
			Query: s.Product,
			Limit: 50,
		}
		candidateEntries, _, err := c.cveRepo.Search(ctx, filter)
		if err == nil {
			for _, e := range candidateEntries {
				// Verifica se la versione è menzionata nei criteri CPE o nella descrizione
				for _, crit := range e.CPEMatchCriteria {
					if strings.Contains(crit.CPE23URI, ":"+s.Version+":") || strings.Contains(crit.CPE23URI, ":"+s.Version) {
						cveMap[e.CVEID] = e
						break
					}
				}
			}
		}
	}

	// 3. Conversione in slice e ordinamento decrescente per CVSS Score
	var result []*domain.CVEEntry
	for _, entry := range cveMap {
		result = append(result, entry)
	}

	sort.Slice(result, func(i, j int) bool {
		scoreI := 0.0
		if result[i].CVSSScore != nil {
			scoreI = *result[i].CVSSScore
		}
		scoreJ := 0.0
		if result[j].CVSSScore != nil {
			scoreJ = *result[j].CVSSScore
		}
		if scoreI != scoreJ {
			return scoreI > scoreJ
		}
		return result[i].CVEID < result[j].CVEID
	})

	return result, nil
}

// CorrelatePort genera la collezione di Finding per una determinata porta aperta di un host
func (c *CVECorrelator) CorrelatePort(
	ctx context.Context,
	scanID, tenantID uuid.UUID,
	host scanner.ParsedHost,
	port scanner.ParsedPort,
) ([]*domain.Finding, error) {
	now := time.Now().UTC()

	var hostname *string
	if host.PrimaryHostname != "" {
		hostname = &host.PrimaryHostname
	}

	var serviceName, serviceVersion, rawBanner *string
	if port.Service != nil {
		if port.Service.Name != "" {
			serviceName = &port.Service.Name
		}
		if port.Service.Product != "" && (serviceName == nil || *serviceName == "") {
			serviceName = &port.Service.Product
		}
		if port.Service.Version != "" {
			serviceVersion = &port.Service.Version
		}
		if port.Service.RawBanner != "" {
			rawBanner = &port.Service.RawBanner
		}
	}

	// Se non vi è alcun servizio associato alla porta, emettiamo un finding base INFO
	if port.Service == nil {
		return []*domain.Finding{
			{
				ID:        uuid.New(),
				ScanID:    scanID,
				TenantID:  tenantID,
				HostIP:    host.IP,
				Hostname:  hostname,
				Port:      port.PortID,
				Protocol:  port.Protocol,
				Severity:  domain.SeverityInfo,
				CreatedAt: now,
			},
		}, nil
	}

	// Correlazione vulnerabilità
	cves, err := c.CorrelateService(ctx, port.Service)
	if err != nil {
		return nil, fmt.Errorf("correlazione servizio porta %d fallita: %w", port.PortID, err)
	}

	// Se nessuna vulnerabilità nota è stata rilevata, registriamo la porta aperta come finding informativo (INFO)
	if len(cves) == 0 {
		return []*domain.Finding{
			{
				ID:             uuid.New(),
				ScanID:         scanID,
				TenantID:       tenantID,
				HostIP:         host.IP,
				Hostname:       hostname,
				Port:           port.PortID,
				Protocol:       port.Protocol,
				ServiceName:    serviceName,
				ServiceVersion: serviceVersion,
				Severity:       domain.SeverityInfo,
				RawBanner:      rawBanner,
				CreatedAt:      now,
			},
		}, nil
	}

	// Se sono state individuate vulnerabilità, creiamo un Finding arricchito per ciascuna CVE
	var findings []*domain.Finding
	sNameStr := ""
	if serviceName != nil {
		sNameStr = *serviceName
	}
	sVerStr := ""
	if serviceVersion != nil {
		sVerStr = *serviceVersion
	}

	for _, cve := range cves {
		cveID := cve.CVEID
		sev := cve.Severity
		if !sev.IsValid() {
			if cve.CVSSScore != nil {
				sev = domain.SeverityFromCVSS(*cve.CVSSScore)
			} else {
				sev = domain.SeverityInfo
			}
		}

		remediation := GenerateRemediation(sNameStr, sVerStr, cveID, sev)

		f := &domain.Finding{
			ID:             uuid.New(),
			ScanID:         scanID,
			TenantID:       tenantID,
			HostIP:         host.IP,
			Hostname:       hostname,
			Port:           port.PortID,
			Protocol:       port.Protocol,
			ServiceName:    serviceName,
			ServiceVersion: serviceVersion,
			CVEID:          &cveID,
			CVSSScore:      cve.CVSSScore,
			Severity:       sev,
			RawBanner:      rawBanner,
			Remediation:    &remediation,
			CreatedAt:      now,
		}
		findings = append(findings, f)
	}

	return findings, nil
}
