package domain

import (
	"encoding/json"
	"strings"
)

const RegulatoryContextVersion = "2026-10-19.2"

type RegulatoryReference struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Source string `json:"source"`
}
type RegulatoryImplication struct {
	Kind        string                `json:"kind"`
	Observation string                `json:"observation"`
	Meaning     string                `json:"meaning"`
	Action      string                `json:"action"`
	References  []RegulatoryReference `json:"references"`
}
type RegulatoryContext struct {
	RuleVersion     string                  `json:"rule_version"`
	Confidence      FindingConfidence       `json:"confidence"`
	DetectionMethod DetectionMethod         `json:"detection_method"`
	Implications    []RegulatoryImplication `json:"implications"`
	Limit           string                  `json:"limit"`
}

// BuildRegulatoryContext is read-only and deterministic. It never infers legal
// compliance, asset identity or verification of a correction from a scan result.
func BuildRegulatoryContext(f Finding) RegulatoryContext {
	r := RegulatoryContext{RuleVersion: RegulatoryContextVersion, Confidence: f.Confidence, DetectionMethod: f.DetectionMethod, Implications: []RegulatoryImplication{}, Limit: "Collegamenti tecnici indicativi, non una valutazione di conformità o di applicabilità legale. Nmap non verifica procedure, formazione, backup o gestione degli incidenti. Regole installate versionate, senza aggiornamento normativo automatico."}
	if f.Port <= 0 || strings.TrimSpace(f.HostIP) == "" || f.HostIP == "0.0.0.0" || (f.PortState != "" && f.PortState != "open" && f.PortState != "open|filtered") {
		return r
	}
	nis := "https://eur-lex.europa.eu/eli/dir/2022/2555/oj/eng"
	dora := "https://eur-lex.europa.eu/eli/reg/2022/2554/oj/eng"
	iso := "https://www.iso.org/standard/27001"
	if f.CVEID != nil && strings.TrimSpace(*f.CVEID) != "" {
		r.Implications = append(r.Implications, RegulatoryImplication{Kind: "vulnerability", Observation: "Il risultato di scansione contiene una correlazione CVE: " + *f.CVEID + ".", Meaning: "La vulnerabilità correlata è pertinente alla gestione delle vulnerabilità e alla protezione dei sistemi. Conserva la confidenza originale: non dimostra sfruttabilità né una violazione normativa.", Action: "Verificare prodotto, versione e prerequisiti nel bollettino del produttore; applicare la correzione supportata o una mitigazione appropriata e verificare con una nuova scansione autorizzata.", References: []RegulatoryReference{{"NIS2 Art. 21(2)(e)", "Gestione delle vulnerabilità", nis}, {"DORA Art. 9", "Protezione e prevenzione", dora}, {"ISO/IEC 27001:2022 A.8.8", "Gestione delle vulnerabilità tecniche", iso}}})
	}
	service := ""
	if f.ServiceName != nil {
		service = strings.ToLower(strings.TrimSpace(*f.ServiceName))
	}
	if service == "telnet" || service == "ftp" || service == "http" {
		r.Implications = append(r.Implications, RegulatoryImplication{Kind: "transport_review", Observation: "Nmap identifica il servizio " + service + ".", Meaning: "Questo servizio può usare comunicazioni non cifrate. Il nome non prova la trasmissione di dati in chiaro: TLS, tunnel e configurazione effettiva non sono stabiliti da questa regola. Il risultato è pertinente alla protezione delle comunicazioni.", Action: "Verificare cifratura e necessità dell'esposizione. Se il traffico è in chiaro, usare un'alternativa cifrata appropriata (SSH/SFTP o HTTPS), limitare gli accessi e disabilitare i servizi non necessari.", References: []RegulatoryReference{{"NIS2 Art. 21(2)(h)", "Uso della crittografia", nis}, {"DORA Art. 9(2)", "Protezione dei dati e dei sistemi", dora}, {"ISO/IEC 27001:2022 A.8.24", "Uso della crittografia", iso}}})
	}
	return r
}

// MarshalJSON aligns live and replayed SSE with REST without persisting derived text.
func (p ScanFindingPayload) MarshalJSON() ([]byte, error) {
	type contextualFinding struct {
		Finding
		RegulatoryContext RegulatoryContext `json:"regulatory_context"`
	}
	return json.Marshal(struct {
		ScanID  string            `json:"scan_id"`
		Finding contextualFinding `json:"finding"`
	}{p.ScanID.String(), contextualFinding{p.Finding, BuildRegulatoryContext(p.Finding)}})
}
