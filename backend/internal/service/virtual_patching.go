package service

import (
	"context"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"nextgen-cybersec/internal/domain"
)

type VirtualPatchFindingReader interface {
	GetByID(context.Context, uuid.UUID, uuid.UUID) (*domain.Finding, error)
}

type VirtualPatchingService struct{ findings VirtualPatchFindingReader }

func NewVirtualPatchingService(findings VirtualPatchFindingReader) *VirtualPatchingService {
	return &VirtualPatchingService{findings: findings}
}

func (s *VirtualPatchingService) Generate(ctx context.Context, tenant, id uuid.UUID) (*domain.VirtualPatchBundle, error) {
	if tenant == uuid.Nil || id == uuid.Nil {
		return nil, domain.ErrValidation
	}
	finding, err := s.findings.GetByID(ctx, tenant, id)
	if err != nil {
		return nil, err
	}
	if finding == nil || finding.TenantID != tenant || finding.ID != id {
		return nil, domain.ErrNotFound
	}
	return BuildVirtualPatchBundle(*finding), nil
}

// The same raw-URI indicator is exported across engines. It intentionally flags
// some benign encoded paths too. It is not a complete exploit detection rule.
const traversalIndicator = `([.][.]/|[.]%2e|%2e[.]|%2e%2e|%252e|%%32%65)`

func BuildVirtualPatchBundle(f domain.Finding) *domain.VirtualPatchBundle {
	b := &domain.VirtualPatchBundle{
		FindingID: f.ID, CatalogRevision: "2026-10-18.2",
		Reason: "Nessun template revisionato per questa CVE e questo servizio.",
		Files:  []domain.VirtualPatchFile{}, Sources: []string{},
		Limitations: []string{
			"Snippet da revisionare e provare in staging; nessuna installazione o reload automatico.",
			"La correlazione CVE non prova vulnerabilità o sfruttabilità. Aggiornare il software resta necessario.",
		},
	}
	for _, entry := range virtualPatchCatalog {
		b.CoveredCVEs = append(b.CoveredCVEs, entry.CVE)
	}
	if f.CVEID != nil {
		b.CVEID = *f.CVEID
	}
	entry, known := findVirtualPatchEntry(b.CVEID)
	if !known || f.Protocol != "tcp" || f.Port <= 0 || f.Port > 65535 || f.ServiceName == nil {
		return b
	}
	service := strings.ToLower(*f.ServiceName)
	if service != "http" && service != "https" && service != "http-proxy" && service != "ssl/http" {
		return b
	}
	b.Supported = true
	b.Product, b.Strategy, b.OperationalImpact = entry.Product, entry.Strategy, entry.Impact
	b.Reason = "Template per " + entry.Product + "; verificare prodotto, versione e configurazione."
	b.Sources = []string{
		entry.Source,
		"https://fail2ban.readthedocs.io/en/latest/filters.html",
		"https://docs.crowdsec.net/docs/log_processor/scenarios/format/",
		"https://github.com/owasp-modsecurity/ModSecurity/wiki/Reference-Manual-(v2.x)",
		"https://nginx.org/en/docs/http/ngx_http_core_module.html",
	}
	b.Limitations = append(b.Limitations,
		"Copertura limitata a percorsi URI noti, anche con falsi positivi; codifiche alternative possono eludere le regole. Body e header non sono coperti. Accessi diretti all'origine aggirano un proxy non obbligatorio.",
		"Fail2ban/CrowdSec reagiscono dopo il log: la prima richiesta può raggiungere il servizio. Richiedono log con URI originale e IP client verificato; dietro proxy configurare la catena fidata.",
		"CrowdSec richiede parser nginx installato, acquisizione dei log, profilo di remediation e bouncer; una segnalazione non prova un blocco.",
		"Verificare percorsi log, collisioni ID ModSecurity e contesto Nginx server. Le regole non sono state validate con i motori nativi dal solo generatore.",
	)
	header := "# NextGen advisory template 2026-10-18.2 - " + entry.CVE + " - " + entry.Product + "\n# Strategy: " + entry.Strategy + "; review operational impact in the bundle.\n# Review and test in staging. Raw-URI coverage only; update the origin.\n"
	basename := "nextgen-" + strings.ToLower(entry.CVE)
	pattern := entry.Pattern
	ruleID := ""
	for i, item := range virtualPatchCatalog {
		if item.CVE == entry.CVE {
			ruleID = strconv.Itoa(1004100 + i)
		}
	}
	add := func(engine, filename, content string) {
		b.Files = append(b.Files, domain.VirtualPatchFile{Engine: engine, Filename: filename, Content: header + content})
	}
	// ConfigParser requires literal percent signs to be doubled.
	add("fail2ban", basename+".conf", "# Install under filter.d; nginx combined log with trusted client address.\n[Definition]\nfailregex = (?i)^<HOST> - \\S+ \\[[^\\]]*\\] \"[A-Z]+ [^ \"?]*"+strings.ReplaceAll(pattern, "%", "%%")+"[^ \"]* HTTP/[0-9.]+\" [0-9]{3} .*$\nignoreregex =\n")
	add("fail2ban", basename+".local", "# Install under jail.d; review logpath, ports and ban action before enabling.\n["+basename+"]\nenabled = false\nfilter = "+basename+"\nport = http,https\nlogpath = /var/log/nginx/access.log\nbackend = auto\nusedns = no\nmaxretry = 1\nfindtime = 60\nbantime = 600\n")
	add("crowdsec", basename+".yaml", "# Requires crowdsecurity/nginx-logs parser and original URI in http_path.\ntype: trigger\nname: nextgen/"+strings.ToLower(entry.CVE)+"\ndescription: Curated URI restriction; review operational impact and false positives\nfilter: \"evt.Meta.log_type == 'http_access-log' && evt.Meta.source_ip != '' && evt.Meta.http_path matches '(?i)"+pattern+"'\"\ngroupby: evt.Meta.source_ip\nblackhole: 1m\nlabels:\n  service: http\n  type: exploit\n  remediation: true\n")
	add("crowdsec", "nextgen-nginx-acquisition.yaml", "# Review path; do not duplicate an existing acquisition.\n# Install as an acquisition file only after nginx parser setup.\nfilenames:\n  - /var/log/nginx/access.log\nlabels:\n  type: nginx\n")
	add("modsecurity", basename+".rules", "# Review unique local rule ID; requires SecRuleEngine On for blocking.\nSecRule REQUEST_URI_RAW \"@rx (?i)"+pattern+"\" \"id:"+ruleID+",phase:1,deny,status:403,log,t:none,msg:'NextGen "+entry.CVE+" URI restriction'\"\n")
	add("nginx", basename+".nginx.conf", "# Include inside each protected server block, before proxy_pass processing.\n# Uses original request_uri, not normalized location matching.\nif ($request_uri ~* \""+pattern+"\") {\n    return 403;\n}\n")
	return b
}
