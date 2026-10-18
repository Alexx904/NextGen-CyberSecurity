package service

// Catalog entries are curated advisory indicators, not an exploit applicability
// engine. Vendor mitigation remains authoritative. Keep this list deterministic.
type virtualPatchEntry struct {
	CVE, Product, Strategy, Pattern, Source, Impact string
}

var virtualPatchCatalog = []virtualPatchEntry{
	{"CVE-2021-41773", "Apache HTTP Server", "uri_indicator", traversalIndicator, "https://httpd.apache.org/security/vulnerabilities_24.html", "Indicatori conservativi di traversal: possibili falsi positivi su percorsi codificati; verificare configurazione Alias/CGI e versione."},
	{"CVE-2021-42013", "Apache HTTP Server", "uri_indicator", traversalIndicator, "https://httpd.apache.org/security/vulnerabilities_24.html", "Indicatori conservativi di traversal: possibili bypass di codifica; verificare configurazione Alias/CGI e versione."},
	{"CVE-2019-19781", "Citrix ADC / Gateway", "endpoint_restriction", `/vpns/`, "https://support.citrix.com/external/article/CTX267679/mitigation-steps-for-cve201919781.html", "Restrizione conservativa del percorso /vpns/: può interrompere funzioni legittime. Non riproduce tutte le condizioni della responder policy Citrix; usare la mitigazione vendor sull'appliance e verificare tutte le vie di accesso."},
	{"CVE-2021-43798", "Grafana", "uri_indicator", `/public/plugins/[^ ?]*` + traversalIndicator, "https://grafana.com/blog/an-update-on-0day-cve-2021-43798-grafana-directory-traversal/", "Indicatori traversal sotto /public/plugins/: verificare reverse proxy e normalizzazione. Gli asset plugin senza indicatori restano fuori dal filtro; varianti non rappresentate possono sfuggire."},
	{"CVE-2021-21972", "VMware vCenter vROps plugin", "endpoint_restriction", `/ui/vropspluginui`, "https://knowledge.broadcom.com/external/article/317865/workaround-instructions-for-cve202121972.html", "Blocca tutte le richieste al plugin vROps, incluse quelle legittime. Protegge solo il traffico che attraversa il controllo: il workaround vendor imposta il plugin incompatible, non semplicemente disabilitato dalla UI."},
	{"CVE-2021-21973", "VMware vCenter vROps plugin", "endpoint_restriction", `/ui/vropspluginui`, "https://knowledge.broadcom.com/external/article/317865/workaround-instructions-for-cve202121972.html", "Blocca tutte le richieste al plugin vROps, incluse quelle legittime. Non sostituisce il workaround vendor che imposta il plugin incompatible né copre accessi diretti all'origine."},
	{"CVE-2023-22515", "Atlassian Confluence", "endpoint_restriction", `/setup/`, "https://confluence.atlassian.com/security/cve-2023-22515-broken-access-control-vulnerability-in-confluence-data-center-and-server-1295682276.html", "Impedisce accesso alle pagine di setup, anche legittimo. Restrizione temporanea indicata dal vendor; valutare contesto applicativo e accessi diretti prima dell'applicazione."},
	{"CVE-2023-22518", "Atlassian Confluence", "endpoint_restriction", `/json/setup-restore(-local|-progress)?[.]action`, "https://confluence.atlassian.com/security/cve-2023-22518-improper-authorization-vulnerability-in-confluence-data-center-and-confluence-server-1311473907.html", "Blocca i tre endpoint di restore indicati dal vendor, interrompendo anche restore legittimi. Misura temporanea: patch e analisi di eventuale compromissione restano necessarie."},
}

func findVirtualPatchEntry(cve string) (virtualPatchEntry, bool) {
	for _, entry := range virtualPatchCatalog {
		if entry.CVE == cve {
			return entry, true
		}
	}
	return virtualPatchEntry{}, false
}
