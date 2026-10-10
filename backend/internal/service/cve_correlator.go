package service

import (
	"context"
	"fmt"
	"regexp"
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
	CorrelateHostScripts(ctx context.Context, scanID, tenantID uuid.UUID, host scanner.ParsedHost) ([]*domain.Finding, error)
}

// CVECorrelator implementa il motore di correlazione tra servizi rilevati, script Nmap e CVE locali/cloud (NFR-REL-03, FR-17)
type CVECorrelator struct {
	cveRepo      domain.CVERepository
	onlineClient OnlineCVEClient
}

// NewCVECorrelator inizializza una nuova istanza di CVECorrelator con supporto opzionale a client online
func NewCVECorrelator(cveRepo domain.CVERepository, onlineClients ...OnlineCVEClient) *CVECorrelator {
	var client OnlineCVEClient
	if len(onlineClients) > 0 && onlineClients[0] != nil {
		client = onlineClients[0]
	} else {
		client = NewNVDOnlineClient()
	}
	return &CVECorrelator{cveRepo: cveRepo, onlineClient: client}
}

// SetOnlineClient aggiorna o disabilita il client online per le query esterne
func (c *CVECorrelator) SetOnlineClient(client OnlineCVEClient) {
	c.onlineClient = client
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

	// 2.5 Ricerca Online Fallback (NIST NVD API 2.0) se non vi sono match locali (FR-17)
	if len(cveMap) == 0 && c.onlineClient != nil {
		var onlineEntries []*domain.CVEEntry

		// Tentativo A: Ricerca per CPE
		for _, rawCPE := range s.CPEs {
			cpe23 := NormalizeCPETo23(rawCPE)
			if cpe23 != "" {
				entries, err := c.onlineClient.SearchByCPE(ctx, cpe23)
				if err == nil && len(entries) > 0 {
					onlineEntries = append(onlineEntries, entries...)
					break
				}
			}
		}

		// Tentativo B: Ricerca per Keyword Prodotto + Versione
		if len(onlineEntries) == 0 && s.Product != "" {
			kw := s.Product
			if s.Version != "" {
				kw += " " + s.Version
			}
			entries, err := c.onlineClient.SearchByKeyword(ctx, kw)
			if err == nil && len(entries) > 0 {
				onlineEntries = append(onlineEntries, entries...)
			}
		}

		// Se abbiamo trovato CVE online, le integriamo e le salviamo su PostgreSQL
		if len(onlineEntries) > 0 {
			for _, e := range onlineEntries {
				cveMap[e.CVEID] = e
			}
			_ = c.cveRepo.UpsertBatch(ctx, onlineEntries)
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
		// Privilegia il nome del software/prodotto reale (es. dnsmasq, vsftpd) rispetto al nome generico della porta (es. domain, ftp)
		if port.Service.Product != "" {
			serviceName = &port.Service.Product
		} else if port.Service.Name != "" {
			serviceName = &port.Service.Name
		}
		if port.Service.Version != "" {
			serviceVersion = &port.Service.Version
		}
		if port.Service.RawBanner != "" {
			rawBanner = &port.Service.RawBanner
		}
	}

	// Se non vi è alcun servizio associato alla porta, emettiamo un finding base INFO
	portState := port.State
	if portState == "" {
		portState = "open"
	}

	var cpePtr *string
	if port.Service != nil && len(port.Service.CPEs) > 0 {
		norm := NormalizeCPETo23(port.Service.CPEs[0])
		if norm != "" {
			cpePtr = &norm
		}
	}

	if port.Service == nil {
		return []*domain.Finding{
			{
				ID:              uuid.New(),
				ScanID:          scanID,
				TenantID:        tenantID,
				HostIP:          host.IP,
				Hostname:        hostname,
				Port:            port.PortID,
				Protocol:        port.Protocol,
				PortState:       portState,
				Severity:        domain.SeverityInfo,
				Confidence:      domain.ConfidenceConfirmed,
				DetectionMethod: domain.DetectionMethodPortState,
				CreatedAt:       now,
			},
		}, nil
	}

	// Correlazione vulnerabilità
	cves, err := c.CorrelateService(ctx, port.Service)
	if err != nil {
		return nil, fmt.Errorf("correlazione servizio porta %d fallita: %w", port.PortID, err)
	}

	// Correlazione script NSE eseguiti sulla porta (es. --script=vuln)
	scriptFindings := c.correlateScriptList(scanID, tenantID, host, port.PortID, port.Protocol, serviceName, serviceVersion, port.Scripts, now)

	// Se nessuna vulnerabilità nota o script problematica è stata rilevata, verifichiamo la presenza di protocolli insicuri noti
	if len(cves) == 0 && len(scriptFindings) == 0 {
		sev := domain.SeverityInfo
		var remText *string
		svcLower := ""
		if port.Service != nil {
			svcLower = strings.ToLower(port.Service.Name + " " + port.Service.Product)
		}

		switch {
		case port.PortID == 23 || strings.Contains(svcLower, "telnet"):
			sev = domain.SeverityMedium
			msg := "Protocollo Telnet in chiaro attivo: tutte le comunicazioni e le credenziali transitano prive di cifratura. Disabilitare Telnet e migrare a SSH (porta 22) con crittografia end-to-end e autenticazione a chiavi."
			remText = &msg
		case port.PortID == 21 || strings.Contains(svcLower, "ftp"):
			sev = domain.SeverityLow
			msg := "Protocollo FTP in chiaro attivo: le credenziali di autenticazione transitano in chiaro sulla rete. Sostituire il servizio con SFTP (SSH File Transfer Protocol) o FTPS con crittografia TLS obbligatoria."
			remText = &msg
		case port.PortID == 445 || port.PortID == 139 || strings.Contains(svcLower, "netbios") || strings.Contains(svcLower, "microsoft-ds") || strings.Contains(svcLower, "smb"):
			sev = domain.SeverityLow
			msg := "Esposizione servizio di condivisione SMB/NetBIOS: limitare l'accesso alla sola subnet fidata tramite firewall e disabilitare le versioni deprecate del protocollo (SMBv1)."
			remText = &msg
		}

		return []*domain.Finding{
			{
				ID:              uuid.New(),
				ScanID:          scanID,
				TenantID:        tenantID,
				HostIP:          host.IP,
				Hostname:        hostname,
				Port:            port.PortID,
				Protocol:        port.Protocol,
				PortState:       portState,
				ServiceName:     serviceName,
				ServiceVersion:  serviceVersion,
				Severity:        sev,
				Confidence:      domain.ConfidenceConfirmed,
				DetectionMethod: domain.DetectionMethodBanner,
				CPE:             cpePtr,
				RawBanner:       rawBanner,
				Remediation:     remText,
				CreatedAt:       now,
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
			ID:              uuid.New(),
			ScanID:          scanID,
			TenantID:        tenantID,
			HostIP:          host.IP,
			Hostname:        hostname,
			Port:            port.PortID,
			Protocol:        port.Protocol,
			PortState:       portState,
			ServiceName:     serviceName,
			ServiceVersion:  serviceVersion,
			CVEID:           &cveID,
			CVSSScore:       cve.CVSSScore,
			Severity:        sev,
			Confidence:      domain.ConfidenceProbable,
			DetectionMethod: domain.DetectionMethodBanner,
			CPE:             cpePtr,
			RawBanner:       rawBanner,
			Remediation:     &remediation,
			CreatedAt:       now,
		}
		findings = append(findings, f)
	}

	if len(scriptFindings) > 0 {
		findings = append(findings, scriptFindings...)
	}

	return findings, nil
}

// CorrelateHostScripts elabora gli script NSE eseguiti a livello di host (es. controlli SMB su host)
func (c *CVECorrelator) CorrelateHostScripts(
	ctx context.Context,
	scanID, tenantID uuid.UUID,
	host scanner.ParsedHost,
) ([]*domain.Finding, error) {
	if len(host.HostScripts) == 0 {
		return nil, nil
	}
	now := time.Now().UTC()
	return c.correlateScriptList(scanID, tenantID, host, 0, "ip", nil, nil, host.HostScripts, now), nil
}

var cveRegex = regexp.MustCompile(`CVE-\d{4}-\d+`)

func (c *CVECorrelator) correlateScriptList(
	scanID, tenantID uuid.UUID,
	host scanner.ParsedHost,
	portID int,
	protocol string,
	serviceName, serviceVersion *string,
	scripts []scanner.ParsedScript,
	now time.Time,
) []*domain.Finding {
	if len(scripts) == 0 {
		return nil
	}

	var findings []*domain.Finding
	var hostname *string
	if host.PrimaryHostname != "" {
		hostname = &host.PrimaryHostname
	}

	for _, sc := range scripts {
		outUpper := strings.ToUpper(sc.Output)
		isVuln := strings.Contains(outUpper, "VULNERABLE") ||
			strings.Contains(outUpper, "RISK FACTOR: HIGH") ||
			strings.Contains(outUpper, "RISK FACTOR: CRITICAL") ||
			(strings.Contains(sc.ID, "vuln") && !strings.Contains(outUpper, "NOT VULNERABLE") && !strings.Contains(outUpper, "ERROR"))

		if !isVuln {
			continue
		}

		// Estrai codici CVE menzionati
		matches := cveRegex.FindAllString(sc.Output, -1)
		matchedCVEs := make(map[string]bool)
		for _, m := range matches {
			matchedCVEs[m] = true
		}

		if len(matchedCVEs) > 0 {
			for cveIDStr := range matchedCVEs {
				idCopy := cveIDStr
				sev := domain.SeverityHigh
				score := 7.5
				if strings.Contains(outUpper, "CRITICAL") || strings.Contains(outUpper, "REMOTE CODE EXECUTION") {
					sev = domain.SeverityCritical
					score = 9.8
				}
				rem := fmt.Sprintf("Mitigazione prescrittiva per %s rilevata dallo script Nmap '%s': applicare tempestivamente le patch di sicurezza o isolare il servizio con regole firewall.", idCopy, sc.ID)
				bannerCopy := sc.Output
				findings = append(findings, &domain.Finding{
					ID:              uuid.New(),
					ScanID:          scanID,
					TenantID:        tenantID,
					HostIP:          host.IP,
					Hostname:        hostname,
					Port:            portID,
					Protocol:        protocol,
					PortState:       "open",
					ServiceName:     serviceName,
					ServiceVersion:  serviceVersion,
					CVEID:           &idCopy,
					CVSSScore:       &score,
					Severity:        sev,
					Confidence:      domain.ConfidenceConfirmed,
					DetectionMethod: domain.DetectionMethodNSE,
					RawBanner:       &bannerCopy,
					Remediation:     &rem,
					CreatedAt:       now,
				})
			}
		} else {
			sev := domain.SeverityMedium
			score := 5.0
			if strings.Contains(outUpper, "HIGH") || strings.Contains(outUpper, "CRITICAL") {
				sev = domain.SeverityHigh
				score = 7.5
			}
			scriptIDCopy := sc.ID
			bannerCopy := sc.Output
			rem := fmt.Sprintf("Vulnerabilità rilevata dallo script Nmap '%s': verificare la configurazione del servizio e applicare le patch raccomandate dal vendor.", sc.ID)
			findings = append(findings, &domain.Finding{
				ID:              uuid.New(),
				ScanID:          scanID,
				TenantID:        tenantID,
				HostIP:          host.IP,
				Hostname:        hostname,
				Port:            portID,
				Protocol:        protocol,
				PortState:       "open",
				ServiceName:     &scriptIDCopy,
				ServiceVersion:  serviceVersion,
				CVSSScore:       &score,
				Severity:        sev,
				Confidence:      domain.ConfidenceConfirmed,
				DetectionMethod: domain.DetectionMethodNSE,
				RawBanner:       &bannerCopy,
				Remediation:     &rem,
				CreatedAt:       now,
			})
		}
	}

	return findings
}
