package report

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"

	"nextgen-cybersec/internal/domain"
)

// PDFReportData racchiude l'intero contesto aggregato per la generazione del report PDF
type PDFReportData struct {
	Tenant      *domain.Tenant
	Scan        *domain.Scan
	Target      *domain.Target
	Profile     *domain.ScanProfile
	Findings    []*domain.Finding
	Template    domain.ReportTemplate
	GeneratedAt time.Time
}

// PDFGenerator definisce il contratto per la generazione di file PDF di sicurezza
type PDFGenerator interface {
	Generate(data PDFReportData) ([]byte, error)
}

// DefaultPDFGenerator implementa PDFGenerator utilizzando gofpdf con supporto multi-template
type DefaultPDFGenerator struct{}

// NewPDFGenerator crea una nuova istanza di DefaultPDFGenerator
func NewPDFGenerator() PDFGenerator {
	return &DefaultPDFGenerator{}
}

// HexToRGB converte un colore esadecimale (#RGB o #RRGGBB) in componenti r, g, b (0-255)
func HexToRGB(hex string) (int, int, int) {
	clean := strings.TrimSpace(hex)
	clean = strings.TrimPrefix(clean, "#")
	if len(clean) == 3 {
		clean = string([]byte{clean[0], clean[0], clean[1], clean[1], clean[2], clean[2]})
	}
	if len(clean) != 6 {
		return 15, 23, 42 // Default primary (#0f172a)
	}
	r, err1 := strconv.ParseInt(clean[0:2], 16, 64)
	g, err2 := strconv.ParseInt(clean[2:4], 16, 64)
	b, err3 := strconv.ParseInt(clean[4:6], 16, 64)
	if err1 != nil || err2 != nil || err3 != nil {
		return 15, 23, 42
	}
	return int(r), int(g), int(b)
}

// SeverityRGB restituisce la terna cromatica standard per ciascun livello di severità
func SeverityRGB(sev domain.SeverityLevel) (int, int, int) {
	switch sev {
	case domain.SeverityCritical:
		return 220, 38, 38 // #dc2626
	case domain.SeverityHigh:
		return 234, 88, 12 // #ea580c
	case domain.SeverityMedium:
		return 217, 119, 6 // #d97706
	case domain.SeverityLow:
		return 37, 99, 235 // #2563eb
	case domain.SeverityInfo:
		fallthrough
	default:
		return 100, 116, 139 // #64748b
	}
}

// Generate assembla e compila il documento PDF multipagina selezionando il template richiesto
func (g *DefaultPDFGenerator) Generate(data PDFReportData) ([]byte, error) {
	// Bonifica a monte: esclusione categorica di finding nulli o con host vuoto/non valido (DoD TASK-35 / TASK-36)
	var validFindings []*domain.Finding
	for _, f := range data.Findings {
		if f != nil && strings.TrimSpace(f.HostIP) != "" && !strings.EqualFold(f.HostIP, "0.0.0.0") {
			validFindings = append(validFindings, f)
		}
	}
	data.Findings = validFindings

	if data.Template == domain.ReportTemplateExecutive {
		return g.generateExecutive(data)
	}
	return g.generateTechnical(data)
}

// =============================================================================
// MODULO 1: GENERAZIONE TEMPLATE EXECUTIVE (Sintetico per Management & Posture)
// =============================================================================

func (g *DefaultPDFGenerator) generateExecutive(data PDFReportData) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AliasNbPages("{nb}")
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	primaryHex := "#0f172a"
	secondaryHex := "#2563eb"
	accentHex := "#38bdf8"
	tenantName := "Security Organization"

	if data.Tenant != nil {
		if data.Tenant.Name != "" {
			tenantName = data.Tenant.Name
		}
		if data.Tenant.BrandColors.Primary != "" {
			primaryHex = data.Tenant.BrandColors.Primary
		}
		if data.Tenant.BrandColors.Secondary != "" {
			secondaryHex = data.Tenant.BrandColors.Secondary
		}
		if data.Tenant.BrandColors.Accent != "" {
			accentHex = data.Tenant.BrandColors.Accent
		}
	}

	prR, prG, prB := HexToRGB(primaryHex)
	secR, secG, secB := HexToRGB(secondaryHex)
	accR, accG, accB := HexToRGB(accentHex)

	pdf.SetFooterFunc(func() {
		if pdf.PageNo() > 1 {
			pdf.SetY(-12)
			pdf.SetFont("Helvetica", "I", 8)
			pdf.SetTextColor(130, 140, 150)
			pdf.CellFormat(0, 8, tr(fmt.Sprintf("%s  |  Executive Security Brief  |  Pagina %d di {nb}", tenantName, pdf.PageNo())), "", 0, "C", false, 0, "")
		}
	})

	// -------------------------------------------------------------------------
	// PAGINA 1: COPERTINA EXECUTIVE WHITE-LABEL
	// -------------------------------------------------------------------------
	pdf.AddPage()

	pdf.SetFillColor(prR, prG, prB)
	pdf.Rect(0, 0, 210, 28, "F")
	pdf.SetFillColor(accR, accG, accB)
	pdf.Rect(0, 28, 210, 3, "F")

	hasLogo := false
	if data.Tenant != nil && data.Tenant.LogoURL != nil {
		hasLogo = registerAndDrawLogo(pdf, *data.Tenant.LogoURL, 15, 8, 40, 15)
	}
	if !hasLogo {
		pdf.SetXY(15, 9)
		pdf.SetFont("Helvetica", "B", 14)
		pdf.SetTextColor(255, 255, 255)
		pdf.CellFormat(0, 10, tr(strings.ToUpper(tenantName)), "", 0, "L", false, 0, "")
	}

	pdf.SetY(60)
	pdf.SetFont("Helvetica", "B", 26)
	pdf.SetTextColor(prR, prG, prB)
	pdf.MultiCell(0, 11, tr("EXECUTIVE SECURITY\nSUMMARY"), "", "L", false)

	pdf.Ln(3)
	pdf.SetFont("Helvetica", "B", 13)
	pdf.SetTextColor(secR, secG, secB)
	pdf.CellFormat(0, 8, tr("Strategic Risk & Security Posture Assessment"), "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(100, 110, 120)
	pdf.CellFormat(0, 6, tr("NextGen CyberSecurity Platform - Automated Black-Box Engine"), "", 1, "L", false, 0, "")

	pdf.Ln(15)
	boxY := pdf.GetY()
	pdf.SetFillColor(248, 250, 252)
	pdf.SetDrawColor(226, 232, 240)
	pdf.RoundedRect(15, boxY, 180, 88, 3, "1234", "FD")

	pdf.SetXY(22, boxY+7)
	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetTextColor(prR, prG, prB)
	pdf.CellFormat(0, 6, tr("METADATI DELL'EXECUTIVE BRIEFING"), "", 1, "L", false, 0, "")

	metaFields := []struct {
		label string
		value string
	}{
		{"Organizzazione:", tenantName},
		{"Target Scope:", getTargetValue(data.Target)},
		{"Tipologia:", getTargetType(data.Target)},
		{"Profilo Scansione:", getProfileName(data.Profile)},
		{"Scan ID:", getScanID(data.Scan)},
		{"Stato Scansione:", getScanStatus(data.Scan)},
		{"Host Monitorati:", getHostScopeSummary(data.Scan)},
		{"Data Esecuzione:", formatTimestamp(getCompletedAt(data.Scan))},
		{"Data Generazione:", data.GeneratedAt.UTC().Format("2006-01-02 15:04:05 UTC")},
	}

	pdf.SetFont("Helvetica", "", 9)
	for _, m := range metaFields {
		pdf.SetX(22)
		pdf.SetTextColor(100, 116, 139)
		pdf.CellFormat(45, 6, tr(m.label), "", 0, "L", false, 0, "")
		pdf.SetTextColor(15, 23, 42)
		pdf.CellFormat(0, 6, tr(m.value), "", 1, "L", false, 0, "")
	}

	pdf.SetY(245)
	pdf.SetFillColor(254, 242, 242)
	pdf.SetDrawColor(254, 202, 202)
	pdf.RoundedRect(15, 245, 180, 15, 2, "1234", "FD")
	pdf.SetXY(15, 248)
	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetTextColor(185, 28, 28)
	pdf.CellFormat(180, 8, tr("CLASSIFICAZIONE: STRETTAMENTE CONFIDENZIALE / EXECUTIVE BRIEFING"), "", 0, "C", false, 0, "")

	// -------------------------------------------------------------------------
	// PAGINA 2: SECURITY POSTURE INDEX & KEY RISK INDICATORS
	// -------------------------------------------------------------------------
	pdf.AddPage()
	drawHeader(pdf, tr, tenantName, prR, prG, prB, accR, accG, accB)

	pdf.SetY(26)
	pdf.SetFont("Helvetica", "B", 14)
	pdf.SetTextColor(prR, prG, prB)
	pdf.CellFormat(0, 8, tr("1. Security Posture Index & Strategic Overview"), "", 1, "L", false, 0, "")

	critCount, highCount, medCount, lowCount, infoCount, maxCVSS, _ := calculateStats(data.Findings)
	totalFindings := len(data.Findings)

	// Calcolo Security Posture Score
	stats := domain.ScanSummaryStats{
		CriticalCount: critCount,
		HighCount:     highCount,
		MediumCount:   medCount,
		LowCount:      lowCount,
		InfoCount:     infoCount,
	}
	if data.Scan != nil && data.Scan.SummaryStats.TotalHosts > 0 {
		stats.TotalHosts = data.Scan.SummaryStats.TotalHosts
		stats.UpHosts = data.Scan.SummaryStats.UpHosts
		stats.DownHosts = data.Scan.SummaryStats.DownHosts
	}
	postureScore, postureGrade := domain.CalculatePostureScore(stats)

	// Card Ingrandita Posture Score & Grade
	postureCardY := pdf.GetY() + 1
	pdf.SetFillColor(248, 250, 252)
	pdf.SetDrawColor(226, 232, 240)
	pdf.RoundedRect(15, postureCardY, 180, 32, 2.5, "1234", "FD")

	// Punteggio numerico e grado a sinistra
	pdf.SetXY(20, postureCardY+3)
	pdf.SetFont("Helvetica", "B", 8)
	pdf.SetTextColor(100, 116, 139)
	pdf.CellFormat(40, 5, tr("SECURITY POSTURE SCORE"), "", 1, "L", false, 0, "")

	pdf.SetX(20)
	pdf.SetFont("Helvetica", "B", 18)
	scoreR, scoreG, scoreB := 22, 163, 74 // Verde standard
	if postureScore < 50 {
		scoreR, scoreG, scoreB = 220, 38, 38 // Rosso
	} else if postureScore < 75 {
		scoreR, scoreG, scoreB = 217, 119, 6 // Arancione
	} else if postureScore < 90 {
		scoreR, scoreG, scoreB = 37, 99, 235 // Blu
	}
	pdf.SetTextColor(scoreR, scoreG, scoreB)
	pdf.CellFormat(40, 8, tr(fmt.Sprintf("%d / 100", postureScore)), "", 1, "L", false, 0, "")

	pdf.SetX(20)
	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetTextColor(prR, prG, prB)
	pdf.CellFormat(40, 5, tr(fmt.Sprintf("GRADE %s", postureGrade)), "", 0, "L", false, 0, "")

	// Separatore verticale
	pdf.SetDrawColor(226, 232, 240)
	pdf.Line(68, postureCardY+4, 68, postureCardY+28)

	// Valutazione qualitativa a destra
	pdf.SetXY(73, postureCardY+4)
	pdf.SetFont("Helvetica", "B", 9.5)
	pdf.SetTextColor(prR, prG, prB)
	pdf.CellFormat(115, 5, tr("Valutazione del Profilo di Rischio Aziendale"), "", 1, "L", false, 0, "")

	pdf.SetX(73)
	pdf.SetFont("Helvetica", "", 8.5)
	pdf.SetTextColor(71, 85, 105)
	var postureText string
	switch postureGrade {
	case "A":
		postureText = "Postura di sicurezza ottimale. Il perimetro analizzato non presenta vulnerabilità critiche o ad elevato impatto. Si raccomanda il mantenimento della cadenza di scansione ordinaria."
	case "B":
		postureText = "Postura di sicurezza soddisfacente con anomalie marginali. Sono state identificate vulnerabilità a severità media/bassa che richiedono una revisione programmata."
	case "C":
		postureText = "Postura di sicurezza con rischio moderato. La presenza di vulnerabilità significative impone l'adozione di misure correttive prima che possano costituire vettori di exploit."
	case "D":
		postureText = "Postura di sicurezza degradata con elevata esposizione. Rilevate molteplici criticità che richiedono l'avvio immediato del piano di remediation per evitare violazioni."
	default:
		postureText = "Postura di sicurezza critica. Il perimetro target presenta vulnerabilità severe (CVSS >= 9.0) attivamente sfruttabili. Richiesto intervento d'emergenza immediato del team SecOps."
	}
	pdf.MultiCell(118, 4.5, tr(postureText), "", "L", false)

	pdf.SetY(postureCardY + 37)

	// 4 Box KPI sintetici
	kpiY := pdf.GetY()
	boxW := 42.0
	spacing := 4.0

	activeHostsStr := "1"
	if data.Scan != nil && data.Scan.SummaryStats.UpHosts > 0 {
		activeHostsStr = fmt.Sprintf("%d", data.Scan.SummaryStats.UpHosts)
	}

	drawKPIBox(pdf, tr, 15+0*(boxW+spacing), kpiY, boxW, 20, "FINDINGS TOTALI", fmt.Sprintf("%d", totalFindings), prR, prG, prB)
	drawKPIBox(pdf, tr, 15+1*(boxW+spacing), kpiY, boxW, 20, "CRITICAL & HIGH", fmt.Sprintf("%d", critCount+highCount), 220, 38, 38)
	drawKPIBox(pdf, tr, 15+2*(boxW+spacing), kpiY, boxW, 20, "MAX CVSS SCORE", fmt.Sprintf("%.1f", maxCVSS), secR, secG, secB)
	drawKPIBox(pdf, tr, 15+3*(boxW+spacing), kpiY, boxW, 20, "HOST ATTIVI", activeHostsStr, 16, 185, 129)

	pdf.SetY(kpiY + 26)

	// Distribuzione Severità
	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetTextColor(prR, prG, prB)
	pdf.CellFormat(0, 6, tr("DISTRIBUZIONE VULNERABILITÀ PER LIVELLO DI SEVERITÀ"), "", 1, "L", false, 0, "")
	pdf.Ln(2)

	sevRows := []struct {
		name  domain.SeverityLevel
		count int
	}{
		{domain.SeverityCritical, critCount},
		{domain.SeverityHigh, highCount},
		{domain.SeverityMedium, medCount},
		{domain.SeverityLow, lowCount},
		{domain.SeverityInfo, infoCount},
	}

	for _, sr := range sevRows {
		r, g, b := SeverityRGB(sr.name)
		pct := 0.0
		if totalFindings > 0 {
			pct = (float64(sr.count) / float64(totalFindings)) * 100.0
		}

		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetTextColor(r, g, b)
		pdf.CellFormat(25, 6, tr(string(sr.name)), "", 0, "L", false, 0, "")

		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(15, 23, 42)
		pdf.CellFormat(18, 6, tr(fmt.Sprintf("%d (%.0f%%)", sr.count, pct)), "", 0, "L", false, 0, "")

		barStartX := 60.0
		barMaxW := 120.0
		barW := (pct / 100.0) * barMaxW
		if sr.count > 0 && barW < 3 {
			barW = 3
		}

		currY := pdf.GetY()
		pdf.SetFillColor(241, 245, 249)
		pdf.Rect(barStartX, currY+1.5, barMaxW, 3.5, "F")
		if barW > 0 {
			pdf.SetFillColor(r, g, b)
			pdf.Rect(barStartX, currY+1.5, barW, 3.5, "F")
		}

		pdf.Ln(7)
	}

	// -------------------------------------------------------------------------
	// PAGINA 3: PRIORITÀ STRATEGICHE DI REMEDIATION & GOVERNANCE ROADMAP
	// -------------------------------------------------------------------------
	pdf.AddPage()
	drawHeader(pdf, tr, tenantName, prR, prG, prB, accR, accG, accB)

	pdf.SetY(26)
	pdf.SetFont("Helvetica", "B", 14)
	pdf.SetTextColor(prR, prG, prB)
	pdf.CellFormat(0, 8, tr("2. Priorità Strategiche di Bonifica (Top Priorities)"), "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(71, 85, 105)
	pdf.MultiCell(0, 5, tr("La seguente selezione isola le anomalie a maggiore magnitudo di rischio per gli asset censiti. Gli interventi elencati consentono la massima riduzione dell'esposizione con il minimo overhead operativo."), "", "L", false)
	pdf.Ln(4)

	// Selezione dei Top Findings ordinati per severità e punteggio CVSS decrescente
	sortedFindings := make([]*domain.Finding, len(data.Findings))
	copy(sortedFindings, data.Findings)
	sort.SliceStable(sortedFindings, func(i, j int) bool {
		sevRank := func(s domain.SeverityLevel) int {
			switch s {
			case domain.SeverityCritical:
				return 4
			case domain.SeverityHigh:
				return 3
			case domain.SeverityMedium:
				return 2
			case domain.SeverityLow:
				return 1
			default:
				return 0
			}
		}
		rI, rJ := sevRank(sortedFindings[i].Severity), sevRank(sortedFindings[j].Severity)
		if rI != rJ {
			return rI > rJ
		}
		scoreI, scoreJ := 0.0, 0.0
		if sortedFindings[i].CVSSScore != nil {
			scoreI = *sortedFindings[i].CVSSScore
		}
		if sortedFindings[j].CVSSScore != nil {
			scoreJ = *sortedFindings[j].CVSSScore
		}
		return scoreI > scoreJ
	})

	maxTop := 5
	if len(sortedFindings) < maxTop {
		maxTop = len(sortedFindings)
	}

	if maxTop == 0 {
		pdf.SetFillColor(240, 253, 244)
		pdf.SetDrawColor(187, 247, 208)
		pdf.RoundedRect(15, pdf.GetY(), 180, 20, 2, "1234", "FD")
		pdf.SetXY(20, pdf.GetY()+5)
		pdf.SetFont("Helvetica", "B", 10)
		pdf.SetTextColor(22, 101, 52)
		pdf.CellFormat(0, 5, tr("Nessuna vulnerabilità identificata"), "", 1, "L", false, 0, "")
		pdf.SetX(20)
		pdf.SetFont("Helvetica", "", 8.5)
		pdf.SetTextColor(21, 128, 61)
		pdf.CellFormat(0, 5, tr("Il perimetro analizzato non presenta servizi vulnerabili noti o falle di sicurezza registrate."), "", 1, "L", false, 0, "")
		pdf.Ln(12)
	} else {
		for i := 0; i < maxTop; i++ {
			f := sortedFindings[i]
			drawExecutivePriorityCard(pdf, tr, i+1, f, prR, prG, prB)
			pdf.Ln(3)
		}
	}

	// Governance & Raccomandazioni Strategiche in chiusura
	if pdf.GetY() > 215 {
		pdf.AddPage()
		drawHeader(pdf, tr, tenantName, prR, prG, prB, accR, accG, accB)
		pdf.SetY(26)
	} else {
		pdf.Ln(4)
	}

	pdf.SetFont("Helvetica", "B", 13)
	pdf.SetTextColor(prR, prG, prB)
	pdf.CellFormat(0, 7, tr("3. Governance & Raccomandazioni Strategiche"), "", 1, "L", false, 0, "")
	pdf.Ln(1)

	govBoxY := pdf.GetY()
	pdf.SetFillColor(248, 250, 252)
	pdf.SetDrawColor(226, 232, 240)
	pdf.RoundedRect(15, govBoxY, 180, 36, 2, "1234", "FD")

	recommendations := []string{
		"Patch Management SLA: Attivare finestre di manutenzione straordinaria per aggiornare i componenti affetti da CVE critiche.",
		"Riduzione Attack Surface: Rivedere le policy di firewall (iptables/nftables) restringendo l'accesso ai soli host fidati.",
		"Continuous Security Validation: Mantenere pianificata la scansione periodica settimanale per intercettare nuove CVE.",
	}

	pdf.SetY(govBoxY + 3)
	pdf.SetFont("Helvetica", "", 8)
	pdf.SetTextColor(30, 41, 59)
	for idx, rec := range recommendations {
		pdf.SetX(20)
		pdf.SetFont("Helvetica", "B", 8)
		pdf.CellFormat(5, 5, tr(fmt.Sprintf("%d.", idx+1)), "", 0, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 8)
		pdf.MultiCell(165, 4.5, tr(rec), "", "L", false)
		pdf.Ln(1)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("generazione output PDF Executive fallita: %w", err)
	}

	return buf.Bytes(), nil
}

// drawExecutivePriorityCard disegna una card compatta focalizzata su CVE, severità e remediation
func drawExecutivePriorityCard(pdf *gofpdf.Fpdf, tr func(string) string, index int, f *domain.Finding, prR, prG, prB int) {
	cardY := pdf.GetY()
	cardW := 180.0
	sevR, sevG, sevB := SeverityRGB(f.Severity)

	pdf.SetFillColor(248, 250, 252)
	pdf.SetDrawColor(226, 232, 240)
	pdf.RoundedRect(15, cardY, cardW, 23, 1.5, "1234", "FD")

	// Barra colorata sinistra
	pdf.SetFillColor(sevR, sevG, sevB)
	pdf.Rect(15, cardY, 3, 23, "F")

	// Titolo card
	pdf.SetXY(21, cardY+2)
	pdf.SetFont("Helvetica", "B", 8.5)
	pdf.SetTextColor(sevR, sevG, sevB)
	scoreStr := "N/A"
	if f.CVSSScore != nil {
		scoreStr = fmt.Sprintf("%.1f", *f.CVSSScore)
	}
	pdf.CellFormat(40, 5, tr(fmt.Sprintf("#%d [%s - CVSS %s]", index, f.Severity, scoreStr)), "", 0, "L", false, 0, "")

	pdf.SetFont("Helvetica", "B", 8.5)
	pdf.SetTextColor(prR, prG, prB)
	svcName := "Unknown Service"
	if f.ServiceName != nil && *f.ServiceName != "" {
		svcName = *f.ServiceName
		if f.ServiceVersion != nil && *f.ServiceVersion != "" {
			svcName += " " + *f.ServiceVersion
		}
	}
	pdf.CellFormat(0, 5, tr(fmt.Sprintf("%s:%d/%s - %s", f.HostIP, f.Port, f.Protocol, svcName)), "", 1, "L", false, 0, "")

	// Dettaglio CVE e Azione di Remediation
	pdf.SetX(21)
	pdf.SetFont("Helvetica", "B", 7.5)
	pdf.SetTextColor(100, 116, 139)
	cveText := "N/A"
	if f.CVEID != nil && *f.CVEID != "" {
		cveText = *f.CVEID
	}
	pdf.CellFormat(20, 4.5, tr(fmt.Sprintf("CVE: %s", cveText)), "", 0, "L", false, 0, "")

	pdf.SetFont("Helvetica", "B", 7.5)
	pdf.SetTextColor(100, 116, 139)
	pdf.CellFormat(20, 4.5, tr("Azione:"), "", 0, "L", false, 0, "")

	pdf.SetFont("Helvetica", "", 7.5)
	pdf.SetTextColor(30, 41, 59)
	remText := "Verificare l'integrità del componente ed applicare le patch del fornitore."
	if f.Remediation != nil && *f.Remediation != "" {
		remText = *f.Remediation
		if len(remText) > 130 {
			remText = remText[:127] + "..."
		}
	}
	pdf.CellFormat(0, 4.5, tr(remText), "", 1, "L", false, 0, "")
	pdf.SetY(cardY + 23)
}

// =============================================================================
// MODULO 2: GENERAZIONE TEMPLATE TECNICO (Dettaglio Tecnico Esteso & Deep-Dive)
// =============================================================================

func (g *DefaultPDFGenerator) generateTechnical(data PDFReportData) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AliasNbPages("{nb}")
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	primaryHex := "#0f172a"
	secondaryHex := "#2563eb"
	accentHex := "#38bdf8"
	tenantName := "Security Organization"

	if data.Tenant != nil {
		if data.Tenant.Name != "" {
			tenantName = data.Tenant.Name
		}
		if data.Tenant.BrandColors.Primary != "" {
			primaryHex = data.Tenant.BrandColors.Primary
		}
		if data.Tenant.BrandColors.Secondary != "" {
			secondaryHex = data.Tenant.BrandColors.Secondary
		}
		if data.Tenant.BrandColors.Accent != "" {
			accentHex = data.Tenant.BrandColors.Accent
		}
	}

	prR, prG, prB := HexToRGB(primaryHex)
	secR, secG, secB := HexToRGB(secondaryHex)
	accR, accG, accB := HexToRGB(accentHex)

	pdf.SetFooterFunc(func() {
		if pdf.PageNo() > 1 {
			pdf.SetY(-12)
			pdf.SetFont("Helvetica", "I", 8)
			pdf.SetTextColor(130, 140, 150)
			pdf.CellFormat(0, 8, tr(fmt.Sprintf("%s  |  Technical Assessment Report  |  Pagina %d di {nb}", tenantName, pdf.PageNo())), "", 0, "C", false, 0, "")
		}
	})

	// -------------------------------------------------------------------------
	// PAGINA 1: COPERTINA WHITE-LABEL TECNICA
	// -------------------------------------------------------------------------
	pdf.AddPage()

	pdf.SetFillColor(prR, prG, prB)
	pdf.Rect(0, 0, 210, 28, "F")
	pdf.SetFillColor(accR, accG, accB)
	pdf.Rect(0, 28, 210, 3, "F")

	hasLogo := false
	if data.Tenant != nil && data.Tenant.LogoURL != nil {
		hasLogo = registerAndDrawLogo(pdf, *data.Tenant.LogoURL, 15, 8, 40, 15)
	}
	if !hasLogo {
		pdf.SetXY(15, 9)
		pdf.SetFont("Helvetica", "B", 14)
		pdf.SetTextColor(255, 255, 255)
		pdf.CellFormat(0, 10, tr(strings.ToUpper(tenantName)), "", 0, "L", false, 0, "")
	}

	pdf.SetY(60)
	pdf.SetFont("Helvetica", "B", 26)
	pdf.SetTextColor(prR, prG, prB)
	pdf.MultiCell(0, 11, tr("SECURITY ASSESSMENT\nREPORT"), "", "L", false)

	pdf.Ln(3)
	pdf.SetFont("Helvetica", "B", 13)
	pdf.SetTextColor(secR, secG, secB)
	pdf.CellFormat(0, 8, tr("Vulnerability & Network Exposure Assessment"), "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(100, 110, 120)
	pdf.CellFormat(0, 6, tr("NextGen CyberSecurity Platform - Automated Black-Box Engine"), "", 1, "L", false, 0, "")

	pdf.Ln(15)
	boxY := pdf.GetY()
	pdf.SetFillColor(248, 250, 252)
	pdf.SetDrawColor(226, 232, 240)
	pdf.RoundedRect(15, boxY, 180, 88, 3, "1234", "FD")

	pdf.SetXY(22, boxY+7)
	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetTextColor(prR, prG, prB)
	pdf.CellFormat(0, 6, tr("METADATI DEL REPORT & SCANSIONE"), "", 1, "L", false, 0, "")

	metaFields := []struct {
		label string
		value string
	}{
		{"Organizzazione:", tenantName},
		{"Target:", getTargetValue(data.Target)},
		{"Profilo:", getProfileName(data.Profile)},
		{"Scan ID:", getScanID(data.Scan)},
		{"Stato Scansione:", getScanStatus(data.Scan)},
		{"Host Monitorati:", getHostScopeSummary(data.Scan)},
		{"Data Avvio:", formatTimestamp(getStartedAt(data.Scan))},
		{"Data Conclusione:", formatTimestamp(getCompletedAt(data.Scan))},
		{"Data Generazione:", data.GeneratedAt.UTC().Format("2006-01-02 15:04:05 UTC")},
	}

	pdf.SetFont("Helvetica", "", 9)
	for _, m := range metaFields {
		pdf.SetX(22)
		pdf.SetTextColor(100, 116, 139)
		pdf.CellFormat(45, 6, tr(m.label), "", 0, "L", false, 0, "")
		pdf.SetTextColor(15, 23, 42)
		pdf.CellFormat(0, 6, tr(m.value), "", 1, "L", false, 0, "")
	}

	pdf.SetY(245)
	pdf.SetFillColor(254, 242, 242)
	pdf.SetDrawColor(254, 202, 202)
	pdf.RoundedRect(15, 245, 180, 15, 2, "1234", "FD")
	pdf.SetXY(15, 248)
	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetTextColor(185, 28, 28)
	pdf.CellFormat(180, 8, tr("CLASSIFICAZIONE: STRETTAMENTE CONFIDENZIALE / USO INTERNO"), "", 0, "C", false, 0, "")

	// -------------------------------------------------------------------------
	// PAGINA 2: EXECUTIVE SUMMARY & METRICHE DI RISCHIO
	// -------------------------------------------------------------------------
	pdf.AddPage()
	drawHeader(pdf, tr, tenantName, prR, prG, prB, accR, accG, accB)

	pdf.SetY(26)
	pdf.SetFont("Helvetica", "B", 14)
	pdf.SetTextColor(prR, prG, prB)
	pdf.CellFormat(0, 8, tr("1. Executive Summary & Metriche di Rischio"), "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "", 9.5)
	pdf.SetTextColor(51, 65, 85)
	introText := "Il presente rapporto riassume i risultati della scansione di sicurezza e dell'analisi delle vulnerabilità condotta sul perimetro target specificato. Le vulnerabilità rilevate sono state correlate con il database locale CVE/CVSS e classificate secondo il Common Vulnerability Scoring System (CVSS v3.1)."
	pdf.MultiCell(0, 5, tr(introText), "", "L", false)
	pdf.Ln(6)

	critCount, highCount, medCount, lowCount, infoCount, maxCVSS, avgCVSS := calculateStats(data.Findings)
	totalFindings := len(data.Findings)

	kpiY := pdf.GetY()
	boxW := 42.0
	spacing := 4.0

	drawKPIBox(pdf, tr, 15+0*(boxW+spacing), kpiY, boxW, 20, "FINDINGS TOTALI", fmt.Sprintf("%d", totalFindings), prR, prG, prB)
	drawKPIBox(pdf, tr, 15+1*(boxW+spacing), kpiY, boxW, 20, "CRITICAL & HIGH", fmt.Sprintf("%d", critCount+highCount), 220, 38, 38)
	drawKPIBox(pdf, tr, 15+2*(boxW+spacing), kpiY, boxW, 20, "MAX CVSS SCORE", fmt.Sprintf("%.1f", maxCVSS), secR, secG, secB)
	drawKPIBox(pdf, tr, 15+3*(boxW+spacing), kpiY, boxW, 20, "MEDIA CVSS", fmt.Sprintf("%.1f", avgCVSS), 100, 116, 139)

	pdf.SetY(kpiY + 26)

	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetTextColor(prR, prG, prB)
	pdf.CellFormat(0, 6, tr("DISTRIBUZIONE VULNERABILITÀ PER LIVELLO DI SEVERITÀ"), "", 1, "L", false, 0, "")
	pdf.Ln(2)

	sevRows := []struct {
		name  domain.SeverityLevel
		count int
	}{
		{domain.SeverityCritical, critCount},
		{domain.SeverityHigh, highCount},
		{domain.SeverityMedium, medCount},
		{domain.SeverityLow, lowCount},
		{domain.SeverityInfo, infoCount},
	}

	for _, sr := range sevRows {
		r, g, b := SeverityRGB(sr.name)
		pct := 0.0
		if totalFindings > 0 {
			pct = (float64(sr.count) / float64(totalFindings)) * 100.0
		}

		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetTextColor(r, g, b)
		pdf.CellFormat(25, 6, tr(string(sr.name)), "", 0, "L", false, 0, "")

		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(15, 23, 42)
		pdf.CellFormat(18, 6, tr(fmt.Sprintf("%d (%.0f%%)", sr.count, pct)), "", 0, "L", false, 0, "")

		barStartX := 60.0
		barMaxW := 120.0
		barW := (pct / 100.0) * barMaxW
		if sr.count > 0 && barW < 3 {
			barW = 3
		}

		currY := pdf.GetY()
		pdf.SetFillColor(241, 245, 249)
		pdf.Rect(barStartX, currY+1.5, barMaxW, 3.5, "F")
		if barW > 0 {
			pdf.SetFillColor(r, g, b)
			pdf.Rect(barStartX, currY+1.5, barW, 3.5, "F")
		}

		pdf.Ln(7)
	}

	// -------------------------------------------------------------------------
	// SEZIONE 2: TARGET SCOPE & OPZIONI DI SCANSIONE
	// -------------------------------------------------------------------------
	pdf.Ln(4)
	pdf.SetFont("Helvetica", "B", 14)
	pdf.SetTextColor(prR, prG, prB)
	pdf.CellFormat(0, 8, tr("2. Perimetro di Scansione & Specifiche Tecniche"), "", 1, "L", false, 0, "")

	pdf.SetFillColor(248, 250, 252)
	pdf.SetDrawColor(226, 232, 240)
	scopeY := pdf.GetY()
	pdf.RoundedRect(15, scopeY, 180, 42, 2, "1234", "FD")

	pdf.SetXY(20, scopeY+4)
	pdf.SetFont("Helvetica", "B", 8.5)
	pdf.SetTextColor(100, 116, 139)
	pdf.CellFormat(35, 5.5, tr("Bersaglio / Valore:"), "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 8.5)
	pdf.SetTextColor(15, 23, 42)
	pdf.CellFormat(55, 5.5, tr(getTargetValue(data.Target)), "", 0, "L", false, 0, "")

	pdf.SetFont("Helvetica", "B", 8.5)
	pdf.SetTextColor(100, 116, 139)
	pdf.CellFormat(35, 5.5, tr("Tipologia Target:"), "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 8.5)
	pdf.SetTextColor(15, 23, 42)
	pdf.CellFormat(0, 5.5, tr(getTargetType(data.Target)), "", 1, "L", false, 0, "")

	pdf.SetX(20)
	pdf.SetFont("Helvetica", "B", 8.5)
	pdf.SetTextColor(100, 116, 139)
	pdf.CellFormat(35, 5.5, tr("Profilo di Scansione:"), "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 8.5)
	pdf.SetTextColor(15, 23, 42)
	pdf.CellFormat(55, 5.5, tr(getProfileName(data.Profile)), "", 0, "L", false, 0, "")

	pdf.SetFont("Helvetica", "B", 8.5)
	pdf.SetTextColor(100, 116, 139)
	pdf.CellFormat(35, 5.5, tr("Argomenti Nmap:"), "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "I", 8.5)
	pdf.SetTextColor(15, 23, 42)
	pdf.CellFormat(0, 5.5, tr(getProfileArgs(data.Profile)), "", 1, "L", false, 0, "")

	pdf.SetX(20)
	pdf.SetFont("Helvetica", "B", 8.5)
	pdf.SetTextColor(100, 116, 139)
	pdf.CellFormat(35, 5.5, tr("Modalità Trigger:"), "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 8.5)
	pdf.SetTextColor(15, 23, 42)
	pdf.CellFormat(55, 5.5, tr(getTriggerType(data.Scan)), "", 0, "L", false, 0, "")

	pdf.SetFont("Helvetica", "B", 8.5)
	pdf.SetTextColor(100, 116, 139)
	pdf.CellFormat(35, 5.5, tr("Durata Totale:"), "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 8.5)
	pdf.SetTextColor(15, 23, 42)
	pdf.CellFormat(0, 5.5, tr(calculateDuration(data.Scan)), "", 1, "L", false, 0, "")

	pdf.SetX(20)
	pdf.SetFont("Helvetica", "B", 8.5)
	pdf.SetTextColor(100, 116, 139)
	pdf.CellFormat(35, 5.5, tr("Host Censiti:"), "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 8.5)
	pdf.SetTextColor(15, 23, 42)
	pdf.CellFormat(0, 5.5, tr(getHostScopeSummary(data.Scan)), "", 1, "L", false, 0, "")

	pdf.SetY(scopeY + 47)

	// -------------------------------------------------------------------------
	// SEZIONE 3: SCHEDE TECNICHE DEI FINDINGS (DEEP-DIVE COMPLETO)
	// -------------------------------------------------------------------------
	pdf.SetFont("Helvetica", "B", 14)
	pdf.SetTextColor(prR, prG, prB)
	pdf.CellFormat(0, 8, tr("3. Schede Tecniche dei Findings"), "", 1, "L", false, 0, "")

	if len(data.Findings) == 0 {
		pdf.SetFont("Helvetica", "I", 10)
		pdf.SetTextColor(100, 116, 139)
		pdf.CellFormat(0, 10, tr("Nessuna vulnerabilità o servizio anomalo identificato durante la sessione di scansione."), "", 1, "L", false, 0, "")
	} else {
		for i, f := range data.Findings {
			if pdf.GetY() > 225 {
				pdf.AddPage()
				drawHeader(pdf, tr, tenantName, prR, prG, prB, accR, accG, accB)
				pdf.SetY(25)
			}

			drawFindingCard(pdf, tr, i+1, f, prR, prG, prB)
			pdf.Ln(4)
		}
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("generazione output PDF Tecnico fallita: %w", err)
	}

	return buf.Bytes(), nil
}

// =============================================================================
// FUNZIONI AUSILIARIE GRAFICHE
// =============================================================================

// drawHeader disegna l'intestazione standard per le pagine interne
func drawHeader(pdf *gofpdf.Fpdf, tr func(string) string, tenantName string, prR, prG, prB, accR, accG, accB int) {
	pdf.SetY(10)
	pdf.SetFont("Helvetica", "B", 8)
	pdf.SetTextColor(prR, prG, prB)
	pdf.CellFormat(90, 5, tr(strings.ToUpper(tenantName)), "", 0, "L", false, 0, "")

	pdf.SetFont("Helvetica", "", 8)
	pdf.SetTextColor(100, 116, 139)
	pdf.CellFormat(90, 5, tr("Vulnerability Assessment Report"), "", 1, "R", false, 0, "")

	pdf.SetDrawColor(accR, accG, accB)
	pdf.SetLineWidth(0.4)
	pdf.Line(15, 16, 195, 16)
	pdf.SetLineWidth(0.2)
}

// drawKPIBox disegna un singolo riquadro KPI sintetico
func drawKPIBox(pdf *gofpdf.Fpdf, tr func(string) string, x, y, w, h float64, label, value string, r, g, b int) {
	pdf.SetFillColor(248, 250, 252)
	pdf.SetDrawColor(226, 232, 240)
	pdf.RoundedRect(x, y, w, h, 2, "1234", "FD")

	pdf.SetXY(x, y+3)
	pdf.SetFont("Helvetica", "B", 7)
	pdf.SetTextColor(100, 116, 139)
	pdf.CellFormat(w, 4, tr(label), "", 2, "C", false, 0, "")

	pdf.SetFont("Helvetica", "B", 13)
	pdf.SetTextColor(r, g, b)
	pdf.CellFormat(w, 7, tr(value), "", 0, "C", false, 0, "")
}

// drawFindingCard disegna una scheda tecnica completa per un finding
func drawFindingCard(pdf *gofpdf.Fpdf, tr func(string) string, index int, f *domain.Finding, prR, prG, prB int) {
	cardY := pdf.GetY()
	cardW := 180.0
	sevR, sevG, sevB := SeverityRGB(f.Severity)

	pdf.SetFillColor(248, 250, 252)
	pdf.SetDrawColor(226, 232, 240)
	pdf.RoundedRect(15, cardY, cardW, 8, 1.5, "12", "FD")

	// Barra colorata verticale sul lato sinistro
	pdf.SetFillColor(sevR, sevG, sevB)
	pdf.Rect(15, cardY, 3, 8, "F")

	// Badge Severità e CVSS
	pdf.SetXY(21, cardY+1.5)
	pdf.SetFont("Helvetica", "B", 8.5)
	pdf.SetTextColor(sevR, sevG, sevB)
	scoreStr := "N/A"
	if f.CVSSScore != nil {
		scoreStr = fmt.Sprintf("%.1f", *f.CVSSScore)
	}
	pdf.CellFormat(35, 5, tr(fmt.Sprintf("#%d [%s - CVSS %s]", index, f.Severity, scoreStr)), "", 0, "L", false, 0, "")

	// Host e Porta
	pdf.SetFont("Helvetica", "B", 8.5)
	pdf.SetTextColor(prR, prG, prB)
	svcName := "Unknown Service"
	if f.ServiceName != nil && *f.ServiceName != "" {
		svcName = *f.ServiceName
		if f.ServiceVersion != nil && *f.ServiceVersion != "" {
			svcName += " " + *f.ServiceVersion
		}
	}
	targetStr := fmt.Sprintf("%s:%d/%s  -  %s", f.HostIP, f.Port, f.Protocol, svcName)
	pdf.CellFormat(0, 5, tr(targetStr), "", 1, "L", false, 0, "")

	// Corpo card
	pdf.SetX(20)
	pdf.SetFont("Helvetica", "", 8)

	if f.CVEID != nil && *f.CVEID != "" {
		pdf.SetTextColor(100, 116, 139)
		pdf.CellFormat(20, 5, tr("CVE ID:"), "", 0, "L", false, 0, "")
		pdf.SetFont("Helvetica", "B", 8)
		pdf.SetTextColor(220, 38, 38)
		pdf.CellFormat(0, 5, tr(*f.CVEID), "", 1, "L", false, 0, "")
	}

	if f.RawBanner != nil && *f.RawBanner != "" {
		pdf.SetX(20)
		pdf.SetFont("Helvetica", "", 8)
		pdf.SetTextColor(100, 116, 139)
		pdf.CellFormat(20, 5, tr("Banner:"), "", 0, "L", false, 0, "")
		pdf.SetFont("Helvetica", "I", 8)
		pdf.SetTextColor(15, 23, 42)
		bannerText := *f.RawBanner
		if len(bannerText) > 120 {
			bannerText = bannerText[:117] + "..."
		}
		pdf.CellFormat(0, 5, tr(bannerText), "", 1, "L", false, 0, "")
	}

	if f.Remediation != nil && *f.Remediation != "" {
		pdf.SetX(20)
		pdf.SetFont("Helvetica", "B", 8)
		pdf.SetTextColor(100, 116, 139)
		pdf.CellFormat(20, 5, tr("Remediation:"), "", 0, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 8)
		pdf.SetTextColor(30, 41, 59)
		pdf.MultiCell(150, 4.5, tr(*f.Remediation), "", "L", false)
	}

	pdf.Ln(2)
	pdf.SetDrawColor(241, 245, 249)
	pdf.Line(15, pdf.GetY(), 195, pdf.GetY())
}

// registerAndDrawLogo gestisce la decodifica base64 del logo e il posizionamento nel PDF
func registerAndDrawLogo(pdf *gofpdf.Fpdf, logoStr string, x, y, maxW, maxH float64) bool {
	if logoStr == "" {
		return false
	}

	var rawBase64 string
	if idx := strings.Index(logoStr, "base64,"); idx != -1 {
		rawBase64 = logoStr[idx+7:]
	} else {
		rawBase64 = logoStr
	}

	data, err := base64.StdEncoding.DecodeString(strings.TrimSpace(rawBase64))
	if err != nil {
		return false
	}

	cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return false
	}

	imageType := "PNG"
	if strings.EqualFold(format, "jpeg") || strings.EqualFold(format, "jpg") {
		imageType = "JPG"
	}

	imageName := "logo_tenant"
	opt := gofpdf.ImageOptions{
		ImageType: imageType,
		ReadDpi:   true,
	}
	pdf.RegisterImageOptionsReader(imageName, opt, bytes.NewReader(data))

	ratio := float64(cfg.Width) / float64(cfg.Height)
	w := maxW
	h := w / ratio
	if h > maxH {
		h = maxH
		w = h * ratio
	}

	pdf.ImageOptions(imageName, x, y, w, h, false, opt, 0, "")
	return true
}

func calculateStats(findings []*domain.Finding) (crit, high, med, low, info int, maxCVSS, avgCVSS float64) {
	totalCVSS := 0.0
	cvssCount := 0

	for _, f := range findings {
		switch f.Severity {
		case domain.SeverityCritical:
			crit++
		case domain.SeverityHigh:
			high++
		case domain.SeverityMedium:
			med++
		case domain.SeverityLow:
			low++
		case domain.SeverityInfo:
			info++
		}

		if f.CVSSScore != nil {
			score := *f.CVSSScore
			if score > maxCVSS {
				maxCVSS = score
			}
			totalCVSS += score
			cvssCount++
		}
	}

	if cvssCount > 0 {
		avgCVSS = totalCVSS / float64(cvssCount)
	}
	return
}

func getTargetValue(t *domain.Target) string {
	if t != nil && t.Value != "" {
		return t.Value
	}
	return "N/A"
}

func getTargetType(t *domain.Target) string {
	if t != nil && t.Type != "" {
		return string(t.Type)
	}
	return "N/A"
}

func getProfileName(p *domain.ScanProfile) string {
	if p != nil && p.Name != "" {
		return p.Name
	}
	return "Standard Profile"
}

func getProfileArgs(p *domain.ScanProfile) string {
	if p != nil && p.NmapArgs != "" {
		return p.NmapArgs
	}
	return "-sV -T4"
}

func getScanID(s *domain.Scan) string {
	if s != nil {
		return s.ID.String()
	}
	return "N/A"
}

func getScanStatus(s *domain.Scan) string {
	if s != nil {
		return string(s.Status)
	}
	return "UNKNOWN"
}

func getStartedAt(s *domain.Scan) *time.Time {
	if s != nil {
		return s.StartedAt
	}
	return nil
}

func getCompletedAt(s *domain.Scan) *time.Time {
	if s != nil {
		return s.CompletedAt
	}
	return nil
}

func getTriggerType(s *domain.Scan) string {
	if s != nil {
		return string(s.TriggerType)
	}
	return "MANUAL"
}

func formatTimestamp(t *time.Time) string {
	if t == nil || t.IsZero() {
		return "N/A"
	}
	return t.UTC().Format("2006-01-02 15:04:05 UTC")
}

func calculateDuration(s *domain.Scan) string {
	if s == nil || s.StartedAt == nil || s.CompletedAt == nil {
		return "N/A"
	}
	dur := s.CompletedAt.Sub(*s.StartedAt).Round(time.Second)
	return dur.String()
}

func getHostScopeSummary(s *domain.Scan) string {
	if s == nil || s.SummaryStats.TotalHosts == 0 {
		return "N/A"
	}
	if s.SummaryStats.DownHosts > 0 {
		return fmt.Sprintf("%d Totali (%d UP, %d Down)", s.SummaryStats.TotalHosts, s.SummaryStats.UpHosts, s.SummaryStats.DownHosts)
	}
	return fmt.Sprintf("%d Totali (%d UP)", s.SummaryStats.TotalHosts, s.SummaryStats.UpHosts)
}
