package report

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	fpdf "github.com/go-pdf/fpdf"

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
	pdf := fpdf.New("P", "mm", "A4", "")
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

	pdf.Ln(12)
	boxY := pdf.GetY()
	pdf.SetFillColor(248, 250, 252)
	pdf.SetDrawColor(226, 232, 240)

	rawMetaFields := []struct {
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

	var metaFields []struct{ label, value string }
	for _, m := range rawMetaFields {
		if m.value != "" && m.value != "N/A" {
			metaFields = append(metaFields, m)
		}
	}

	boxH := float64(len(metaFields))*5.8 + 10.0
	pdf.RoundedRect(15, boxY, 180, boxH, 2.5, "1234", "FD")

	pdf.SetXY(22, boxY+6)
	pdf.SetFont("Helvetica", "B", 10.5)
	pdf.SetTextColor(prR, prG, prB)
	pdf.CellFormat(0, 5, tr("METADATI DELL'EXECUTIVE BRIEFING"), "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "", 8.5)
	for _, m := range metaFields {
		pdf.SetX(22)
		pdf.SetTextColor(100, 116, 139)
		pdf.CellFormat(45, 5.8, tr(m.label), "", 0, "L", false, 0, "")
		pdf.SetTextColor(15, 23, 42)
		pdf.CellFormat(0, 5.8, tr(m.value), "", 1, "L", false, 0, "")
	}

	classY := boxY + boxH + 22.0
	if classY < 235 {
		classY = 235
	}
	pdf.SetFillColor(254, 242, 242)
	pdf.SetDrawColor(254, 202, 202)
	pdf.RoundedRect(15, classY, 180, 13, 2, "1234", "FD")
	pdf.SetXY(15, classY+3)
	pdf.SetFont("Helvetica", "B", 8.5)
	pdf.SetTextColor(185, 28, 28)
	pdf.CellFormat(180, 7, tr("CLASSIFICAZIONE: STRETTAMENTE CONFIDENZIALE / EXECUTIVE BRIEFING"), "", 0, "C", false, 0, "")

	// -------------------------------------------------------------------------
	// PAGINA 2: EXECUTIVE SECURITY BRIEF (2 PAGINE TOTALI)
	// -------------------------------------------------------------------------
	pdf.AddPage()
	drawHeader(pdf, tr, tenantName, "Executive Security Summary", prR, prG, prB, accR, accG, accB)

	pdf.SetY(22)
	pdf.SetFont("Helvetica", "B", 13)
	pdf.SetTextColor(prR, prG, prB)
	pdf.CellFormat(0, 7, tr("1. Security Posture Index & Strategic Overview"), "", 1, "L", false, 0, "")

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

	// Card Posture Score & Grade
	postureCardY := pdf.GetY() + 1
	pdf.SetFillColor(248, 250, 252)
	pdf.SetDrawColor(226, 232, 240)
	pdf.RoundedRect(15, postureCardY, 180, 25, 2, "1234", "FD")

	// Punteggio numerico e grado a sinistra
	pdf.SetXY(20, postureCardY+2.5)
	pdf.SetFont("Helvetica", "B", 7.5)
	pdf.SetTextColor(100, 116, 139)
	pdf.CellFormat(45, 4, tr("SECURITY POSTURE SCORE"), "", 1, "L", false, 0, "")

	pdf.SetX(20)
	pdf.SetFont("Helvetica", "B", 16)
	scoreR, scoreG, scoreB := 22, 163, 74 // Verde standard
	if postureScore < 50 {
		scoreR, scoreG, scoreB = 220, 38, 38 // Rosso
	} else if postureScore < 75 {
		scoreR, scoreG, scoreB = 217, 119, 6 // Arancione
	} else if postureScore < 90 {
		scoreR, scoreG, scoreB = 37, 99, 235 // Blu
	}
	pdf.SetTextColor(scoreR, scoreG, scoreB)
	pdf.CellFormat(45, 7, tr(fmt.Sprintf("%d / 100", postureScore)), "", 1, "L", false, 0, "")

	pdf.SetX(20)
	pdf.SetFont("Helvetica", "B", 8.5)
	pdf.SetTextColor(prR, prG, prB)
	pdf.CellFormat(45, 4.5, tr(fmt.Sprintf("GRADE %s", postureGrade)), "", 0, "L", false, 0, "")

	// Separatore verticale
	pdf.SetDrawColor(226, 232, 240)
	pdf.Line(68, postureCardY+3, 68, postureCardY+22)

	// Valutazione qualitativa a destra (aderente al 100% ai dati reali)
	pdf.SetXY(73, postureCardY+2.5)
	pdf.SetFont("Helvetica", "B", 8.5)
	pdf.SetTextColor(prR, prG, prB)
	pdf.CellFormat(115, 4, tr("Valutazione del Profilo di Rischio Aziendale"), "", 1, "L", false, 0, "")

	pdf.SetXY(73, postureCardY+7)
	pdf.SetFont("Helvetica", "", 7.8)
	pdf.SetTextColor(71, 85, 105)
	var postureText string
	if critCount > 0 {
		postureText = fmt.Sprintf("Postura di sicurezza critica. Rilevate %d vulnerabilità critiche (CVSS >= 9.0) attivamente sfruttabili sul perimetro. Richiesto intervento d'emergenza immediato del team SecOps.", critCount)
	} else if highCount > 0 {
		postureText = fmt.Sprintf("Postura di sicurezza ad attenzione prioritaria. Nessuna vulnerabilità critica (CVSS >= 9.0) riscontrata; rilevate %d vulnerabilità ad alto impatto (CVSS 7.0-8.9) che richiedono bonifica e segmentazione dei servizi esposti.", highCount)
	} else if medCount > 0 {
		postureText = fmt.Sprintf("Postura di sicurezza con rischio moderato. Rilevate %d vulnerabilità a severità media. Si raccomanda la pianificazione ordinaria delle patch di sicurezza sui componenti identificati.", medCount)
	} else if lowCount > 0 {
		postureText = fmt.Sprintf("Postura di sicurezza soddisfacente. Rilevate solo %d anomalie minori a basso impatto, gestibili nei cicli standard di manutenzione.", lowCount)
	} else {
		postureText = "Postura di sicurezza ottimale. Il perimetro analizzato non presenta anomalie o vulnerabilità note nei servizi esposti."
	}
	pdf.MultiCell(118, 3.8, tr(postureText), "", "L", false)

	// 4 Box KPI sintetici con colori semantici coerenti
	kpiY := postureCardY + 28.0
	boxW := 42.0
	spacing := 4.0

	activeHostsStr := "1"
	if data.Scan != nil && data.Scan.SummaryStats.UpHosts > 0 {
		activeHostsStr = fmt.Sprintf("%d", data.Scan.SummaryStats.UpHosts)
	}

	// Colore semantico per MAX CVSS SCORE in base alla severità effettiva
	cvssSev := domain.SeverityFromCVSS(maxCVSS)
	cvssR, cvssG, cvssB := SeverityRGB(cvssSev)
	if maxCVSS == 0 {
		cvssR, cvssG, cvssB = 100, 116, 139
	}

	critHighR, critHighG, critHighB := 100, 116, 139
	if critCount+highCount > 0 {
		critHighR, critHighG, critHighB = 220, 38, 38
	}

	drawKPIBox(pdf, tr, 15+0*(boxW+spacing), kpiY, boxW, 14, "FINDINGS TOTALI", fmt.Sprintf("%d", totalFindings), prR, prG, prB)
	drawKPIBox(pdf, tr, 15+1*(boxW+spacing), kpiY, boxW, 14, "CRITICAL & HIGH", fmt.Sprintf("%d", critCount+highCount), critHighR, critHighG, critHighB)
	drawKPIBox(pdf, tr, 15+2*(boxW+spacing), kpiY, boxW, 14, "MAX CVSS SCORE", fmt.Sprintf("%.1f", maxCVSS), cvssR, cvssG, cvssB)
	drawKPIBox(pdf, tr, 15+3*(boxW+spacing), kpiY, boxW, 14, "HOST ATTIVI", activeHostsStr, 30, 41, 59)

	// Distribuzione Severità compatta
	distY := kpiY + 17.0
	pdf.SetY(distY)
	pdf.SetFont("Helvetica", "B", 9.5)
	pdf.SetTextColor(prR, prG, prB)
	pdf.CellFormat(0, 5, tr("DISTRIBUZIONE VULNERABILITÀ PER LIVELLO DI SEVERITÀ"), "", 1, "L", false, 0, "")
	pdf.Ln(1)

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

		pdf.SetFont("Helvetica", "B", 8)
		pdf.SetTextColor(r, g, b)
		pdf.CellFormat(24, 4.5, tr(string(sr.name)), "", 0, "L", false, 0, "")

		pdf.SetFont("Helvetica", "", 8)
		pdf.SetTextColor(15, 23, 42)
		pdf.CellFormat(18, 4.5, tr(fmt.Sprintf("%d (%.0f%%)", sr.count, pct)), "", 0, "L", false, 0, "")

		barStartX := 60.0
		barMaxW := 120.0
		barW := (pct / 100.0) * barMaxW
		if sr.count > 0 && barW < 3 {
			barW = 3
		}

		currY := pdf.GetY()
		pdf.SetFillColor(241, 245, 249)
		pdf.Rect(barStartX, currY+1.0, barMaxW, 2.8, "F")
		if barW > 0 {
			pdf.SetFillColor(r, g, b)
			pdf.Rect(barStartX, currY+1.0, barW, 2.8, "F")
		}

		pdf.Ln(4.8)
	}

	// -------------------------------------------------------------------------
	// SEZIONE 2: PRIORITÀ STRATEGICHE DI REMEDIATION (DI SEGUITO NELLA PAGINA 2)
	// -------------------------------------------------------------------------
	pdf.Ln(2)
	pdf.SetFont("Helvetica", "B", 11.5)
	pdf.SetTextColor(prR, prG, prB)
	pdf.CellFormat(0, 6, tr("2. Priorità Strategiche di Bonifica (Top Priorities)"), "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "", 7.8)
	pdf.SetTextColor(71, 85, 105)
	pdf.MultiCell(0, 3.8, tr("La seguente selezione isola le anomalie a maggiore magnitudo di rischio per gli asset censiti. Gli interventi prioritari consentono la massima riduzione dell'esposizione con il minimo overhead operativo."), "", "L", false)
	pdf.Ln(2)

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
		pdf.RoundedRect(15, pdf.GetY(), 180, 16, 2, "1234", "FD")
		pdf.SetXY(20, pdf.GetY()+3)
		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetTextColor(22, 101, 52)
		pdf.CellFormat(0, 4.5, tr("Nessuna vulnerabilità identificata"), "", 1, "L", false, 0, "")
		pdf.SetX(20)
		pdf.SetFont("Helvetica", "", 8)
		pdf.SetTextColor(21, 128, 61)
		pdf.CellFormat(0, 4.5, tr("Il perimetro analizzato non presenta servizi vulnerabili noti o falle di sicurezza registrate."), "", 1, "L", false, 0, "")
		pdf.Ln(8)
	} else {
		for i := 0; i < maxTop; i++ {
			f := sortedFindings[i]
			cardEstH := estimateExecutiveCardHeight(f)
			if pdf.GetY()+cardEstH > 268 {
				pdf.AddPage()
				drawHeader(pdf, tr, tenantName, "Executive Security Summary", prR, prG, prB, accR, accG, accB)
				pdf.SetY(24)
			}
			drawExecutivePriorityCard(pdf, tr, i+1, f, prR, prG, prB)
		}
	}

	// -------------------------------------------------------------------------
	// SEZIONE 3: GOVERNANCE & RACCOMANDAZIONI STRATEGICHE
	// -------------------------------------------------------------------------
	if pdf.GetY() > 228 {
		pdf.AddPage()
		drawHeader(pdf, tr, tenantName, "Executive Security Summary", prR, prG, prB, accR, accG, accB)
		pdf.SetY(24)
	} else {
		pdf.Ln(1)
	}

	pdf.SetFont("Helvetica", "B", 10.5)
	pdf.SetTextColor(prR, prG, prB)
	pdf.CellFormat(0, 5.5, tr("3. Governance & Raccomandazioni Strategiche"), "", 1, "L", false, 0, "")
	pdf.Ln(1)

	govBoxY := pdf.GetY()
	pdf.SetFillColor(248, 250, 252)
	pdf.SetDrawColor(226, 232, 240)
	pdf.RoundedRect(15, govBoxY, 180, 24, 2, "1234", "FD")

	recommendations := []string{
		"Patch Management SLA: Attivare finestre di manutenzione straordinaria per aggiornare tempestivamente i componenti esposti.",
		"Riduzione Attack Surface: Rivedere le policy di firewall restringendo l'accesso perimetrale ai soli host e subnet autorizzati.",
		"Continuous Security Validation: Mantenere pianificata la scansione periodica automatica per rilevare tempestivamente nuove CVE.",
	}

	pdf.SetY(govBoxY + 2.5)
	for idx, rec := range recommendations {
		pdf.SetX(20)
		pdf.SetFont("Helvetica", "B", 7.5)
		pdf.SetTextColor(prR, prG, prB)
		pdf.CellFormat(5, 4.2, tr(fmt.Sprintf("%d.", idx+1)), "", 0, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 7.5)
		pdf.SetTextColor(30, 41, 59)
		pdf.MultiCell(165, 4.2, tr(rec), "", "L", false)
		pdf.Ln(0.5)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("generazione output PDF Executive fallita: %w", err)
	}

	return buf.Bytes(), nil
}

// estimateExecutiveCardHeight calcola l'altezza prevista per la card Executive
func estimateExecutiveCardHeight(f *domain.Finding) float64 {
	remText := "Verificare l'integrità del componente ed applicare le patch del fornitore."
	if f.Remediation != nil && strings.TrimSpace(*f.Remediation) != "" {
		remText = *f.Remediation
	}
	remLines := int(math.Ceil(float64(len(remText)) / 85.0))
	if remLines < 1 {
		remLines = 1
	}
	return 12.0 + float64(remLines)*4.0 + 2.5
}

// drawExecutivePriorityCard disegna una card compatta focalizzata su CVE, severità e remediation
func drawExecutivePriorityCard(pdf *fpdf.Fpdf, tr func(string) string, index int, f *domain.Finding, prR, prG, prB int) {
	cardY := pdf.GetY()
	cardW := 180.0
	sevR, sevG, sevB := SeverityRGB(f.Severity)

	scoreStr := "N/A"
	if f.CVSSScore != nil {
		scoreStr = fmt.Sprintf("%.1f", *f.CVSSScore)
	}
	badgeStr := fmt.Sprintf("#%d [%s - CVSS %s]", index, f.Severity, scoreStr)
	badgeW := pdf.GetStringWidth(badgeStr) + 3.0

	svcName := CleanServiceName(f)
	targetStr := fmt.Sprintf("%s:%d/%s  -  %s", f.HostIP, f.Port, f.Protocol, svcName)

	cveText := "N/A"
	if f.CVEID != nil && *f.CVEID != "" {
		cveText = *f.CVEID
	}

	remText := "Verificare l'integrità del componente ed applicare le patch del fornitore."
	if f.Remediation != nil && *f.Remediation != "" {
		remText = *f.Remediation
	}

	remLines := int(math.Ceil(float64(len(remText)) / 85.0))
	if remLines < 1 {
		remLines = 1
	}
	cardH := 12.0 + float64(remLines)*4.0 + 2.0

	// Sfondo card e bordo
	pdf.SetFillColor(248, 250, 252)
	pdf.SetDrawColor(226, 232, 240)
	pdf.RoundedRect(15, cardY, cardW, cardH, 1.5, "1234", "FD")

	// Barra colorata laterale sinistra a tutta altezza
	pdf.SetFillColor(sevR, sevG, sevB)
	pdf.Rect(15, cardY, 3.5, cardH, "F")

	// Riga 1: Badge + Target
	pdf.SetXY(22, cardY+2)
	pdf.SetFont("Helvetica", "B", 8.5)
	pdf.SetTextColor(sevR, sevG, sevB)
	pdf.CellFormat(badgeW, 4.5, tr(badgeStr), "", 0, "L", false, 0, "")

	pdf.SetXY(22+badgeW+2.5, cardY+2)
	pdf.SetFont("Helvetica", "B", 8.5)
	pdf.SetTextColor(prR, prG, prB)
	pdf.CellFormat(0, 4.5, tr(targetStr), "", 1, "L", false, 0, "")

	// Riga 2: CVE ID su riga dedicata
	pdf.SetXY(22, cardY+7)
	pdf.SetFont("Helvetica", "B", 8)
	pdf.SetTextColor(100, 116, 139)
	pdf.CellFormat(16, 4.0, tr("CVE ID:"), "", 0, "L", false, 0, "")

	pdf.SetFont("Helvetica", "B", 8)
	pdf.SetTextColor(220, 38, 38)
	pdf.CellFormat(0, 4.0, tr(cveText), "", 1, "L", false, 0, "")

	// Riga 3: Azione con wrapping MultiCell
	pdf.SetXY(22, cardY+11.5)
	pdf.SetFont("Helvetica", "B", 8)
	pdf.SetTextColor(100, 116, 139)
	pdf.CellFormat(16, 4.0, tr("Azione:"), "", 0, "L", false, 0, "")

	pdf.SetXY(38, cardY+11.5)
	pdf.SetFont("Helvetica", "", 8)
	pdf.SetTextColor(30, 41, 59)
	pdf.MultiCell(154, 4.0, tr(remText), "", "L", false)

	pdf.SetY(cardY + cardH + 2.0)
}

// =============================================================================
// MODULO 2: GENERAZIONE TEMPLATE TECNICO (Dettaglio Tecnico Esteso & Deep-Dive)
// =============================================================================

func (g *DefaultPDFGenerator) generateTechnical(data PDFReportData) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
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

	pdf.Ln(12)
	boxY := pdf.GetY()
	pdf.SetFillColor(248, 250, 252)
	pdf.SetDrawColor(226, 232, 240)

	rawMetaFields := []struct {
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

	var metaFields []struct{ label, value string }
	for _, m := range rawMetaFields {
		if m.value != "" && m.value != "N/A" {
			metaFields = append(metaFields, m)
		}
	}

	boxH := float64(len(metaFields))*5.8 + 10.0
	pdf.RoundedRect(15, boxY, 180, boxH, 2.5, "1234", "FD")

	pdf.SetXY(22, boxY+6)
	pdf.SetFont("Helvetica", "B", 10.5)
	pdf.SetTextColor(prR, prG, prB)
	pdf.CellFormat(0, 5, tr("METADATI DEL REPORT & SCANSIONE"), "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "", 8.5)
	for _, m := range metaFields {
		pdf.SetX(22)
		pdf.SetTextColor(100, 116, 139)
		pdf.CellFormat(45, 5.8, tr(m.label), "", 0, "L", false, 0, "")
		pdf.SetTextColor(15, 23, 42)
		pdf.CellFormat(0, 5.8, tr(m.value), "", 1, "L", false, 0, "")
	}

	classY := boxY + boxH + 22.0
	if classY < 235 {
		classY = 235
	}
	pdf.SetFillColor(254, 242, 242)
	pdf.SetDrawColor(254, 202, 202)
	pdf.RoundedRect(15, classY, 180, 13, 2, "1234", "FD")
	pdf.SetXY(15, classY+3)
	pdf.SetFont("Helvetica", "B", 8.5)
	pdf.SetTextColor(185, 28, 28)
	pdf.CellFormat(180, 7, tr("CLASSIFICAZIONE: STRETTAMENTE CONFIDENZIALE / USO INTERNO"), "", 0, "C", false, 0, "")

	// -------------------------------------------------------------------------
	// PAGINA 2: EXECUTIVE SUMMARY & METRICHE DI RISCHIO
	// -------------------------------------------------------------------------
	pdf.AddPage()
	drawHeader(pdf, tr, tenantName, "Technical Assessment Report", prR, prG, prB, accR, accG, accB)

	pdf.SetY(22)
	pdf.SetFont("Helvetica", "B", 13)
	pdf.SetTextColor(prR, prG, prB)
	pdf.CellFormat(0, 7, tr("1. Executive Summary & Metriche di Rischio"), "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "", 8.5)
	pdf.SetTextColor(51, 65, 85)
	introText := "Il presente rapporto riassume i risultati della scansione di sicurezza e dell'analisi delle vulnerabilità condotta sul perimetro target specificato. Le vulnerabilità rilevate sono state correlate con il database locale CVE/CVSS e classificate secondo il Common Vulnerability Scoring System (CVSS v3.1)."
	pdf.MultiCell(0, 4.5, tr(introText), "", "L", false)
	pdf.Ln(4)

	critCount, highCount, medCount, lowCount, infoCount, maxCVSS, avgCVSS := calculateStats(data.Findings)
	totalFindings := len(data.Findings)

	kpiY := pdf.GetY()
	boxW := 42.0
	spacing := 4.0

	// Colore semantico rigoroso: CVSS Score secondo la gravità effettiva (Arancione se High, Rosso se Critico)
	cvssSev := domain.SeverityFromCVSS(maxCVSS)
	cvssR, cvssG, cvssB := SeverityRGB(cvssSev)
	if maxCVSS == 0 {
		cvssR, cvssG, cvssB = 100, 116, 139
	}

	critHighR, critHighG, critHighB := 100, 116, 139
	if critCount+highCount > 0 {
		critHighR, critHighG, critHighB = 220, 38, 38
	}

	drawKPIBox(pdf, tr, 15+0*(boxW+spacing), kpiY, boxW, 14, "FINDINGS TOTALI", fmt.Sprintf("%d", totalFindings), prR, prG, prB)
	drawKPIBox(pdf, tr, 15+1*(boxW+spacing), kpiY, boxW, 14, "CRITICAL & HIGH", fmt.Sprintf("%d", critCount+highCount), critHighR, critHighG, critHighB)
	drawKPIBox(pdf, tr, 15+2*(boxW+spacing), kpiY, boxW, 14, "MAX CVSS SCORE", fmt.Sprintf("%.1f", maxCVSS), cvssR, cvssG, cvssB)
	drawKPIBox(pdf, tr, 15+3*(boxW+spacing), kpiY, boxW, 14, "MEDIA CVSS", fmt.Sprintf("%.1f", avgCVSS), 100, 116, 139)

	distY := kpiY + 17.0
	pdf.SetY(distY)
	pdf.SetFont("Helvetica", "B", 9.5)
	pdf.SetTextColor(prR, prG, prB)
	pdf.CellFormat(0, 5, tr("DISTRIBUZIONE VULNERABILITÀ PER LIVELLO DI SEVERITÀ"), "", 1, "L", false, 0, "")
	pdf.Ln(1)

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

		pdf.SetFont("Helvetica", "B", 8)
		pdf.SetTextColor(r, g, b)
		pdf.CellFormat(24, 4.5, tr(string(sr.name)), "", 0, "L", false, 0, "")

		pdf.SetFont("Helvetica", "", 8)
		pdf.SetTextColor(15, 23, 42)
		pdf.CellFormat(18, 4.5, tr(fmt.Sprintf("%d (%.0f%%)", sr.count, pct)), "", 0, "L", false, 0, "")

		barStartX := 60.0
		barMaxW := 120.0
		barW := (pct / 100.0) * barMaxW
		if sr.count > 0 && barW < 3 {
			barW = 3
		}

		currY := pdf.GetY()
		pdf.SetFillColor(241, 245, 249)
		pdf.Rect(barStartX, currY+1.0, barMaxW, 2.8, "F")
		if barW > 0 {
			pdf.SetFillColor(r, g, b)
			pdf.Rect(barStartX, currY+1.0, barW, 2.8, "F")
		}

		pdf.Ln(4.8)
	}

	// -------------------------------------------------------------------------
	// SEZIONE 2: TARGET SCOPE & OPZIONI DI SCANSIONE (ALTEZZA ADERENTE, ZERO GHOST GAP)
	// -------------------------------------------------------------------------
	pdf.Ln(3)
	pdf.SetFont("Helvetica", "B", 12)
	pdf.SetTextColor(prR, prG, prB)
	pdf.CellFormat(0, 6, tr("2. Perimetro di Scansione & Specifiche Tecniche"), "", 1, "L", false, 0, "")

	pdf.SetFillColor(248, 250, 252)
	pdf.SetDrawColor(226, 232, 240)
	scopeY := pdf.GetY()
	pdf.RoundedRect(15, scopeY, 180, 26, 2, "1234", "FD")

	col1X := 20.0
	col2X := 110.0
	col1LabelW := 33.0
	col1ValW := 51.0
	col2LabelW := 30.0
	col2ValW := 52.0
	rowH := 5.2

	// Riga 1
	pdf.SetXY(col1X, scopeY+2.5)
	pdf.SetFont("Helvetica", "B", 8)
	pdf.SetTextColor(100, 116, 139)
	pdf.CellFormat(col1LabelW, rowH, tr("Bersaglio / Valore:"), "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 8)
	pdf.SetTextColor(15, 23, 42)
	pdf.CellFormat(col1ValW, rowH, tr(TruncateString(getTargetValue(data.Target), 28)), "", 0, "L", false, 0, "")

	pdf.SetXY(col2X, scopeY+2.5)
	pdf.SetFont("Helvetica", "B", 8)
	pdf.SetTextColor(100, 116, 139)
	pdf.CellFormat(col2LabelW, rowH, tr("Tipologia Target:"), "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 8)
	pdf.SetTextColor(15, 23, 42)
	pdf.CellFormat(col2ValW, rowH, tr(getTargetType(data.Target)), "", 1, "L", false, 0, "")

	// Riga 2
	pdf.SetXY(col1X, scopeY+2.5+rowH)
	pdf.SetFont("Helvetica", "B", 8)
	pdf.SetTextColor(100, 116, 139)
	pdf.CellFormat(col1LabelW, rowH, tr("Profilo di Scansione:"), "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 8)
	pdf.SetTextColor(15, 23, 42)
	pdf.CellFormat(col1ValW, rowH, tr(TruncateString(getProfileName(data.Profile), 26)), "", 0, "L", false, 0, "")

	pdf.SetXY(col2X, scopeY+2.5+rowH)
	pdf.SetFont("Helvetica", "B", 8)
	pdf.SetTextColor(100, 116, 139)
	pdf.CellFormat(col2LabelW, rowH, tr("Argomenti Nmap:"), "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "I", 7.8)
	pdf.SetTextColor(15, 23, 42)
	pdf.CellFormat(col2ValW, rowH, tr(TruncateString(getProfileArgs(data.Profile), 30)), "", 1, "L", false, 0, "")

	// Riga 3
	pdf.SetXY(col1X, scopeY+2.5+rowH*2)
	pdf.SetFont("Helvetica", "B", 8)
	pdf.SetTextColor(100, 116, 139)
	pdf.CellFormat(col1LabelW, rowH, tr("Modalità Trigger:"), "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 8)
	pdf.SetTextColor(15, 23, 42)
	pdf.CellFormat(col1ValW, rowH, tr(getTriggerType(data.Scan)), "", 0, "L", false, 0, "")

	pdf.SetXY(col2X, scopeY+2.5+rowH*2)
	pdf.SetFont("Helvetica", "B", 8)
	pdf.SetTextColor(100, 116, 139)
	pdf.CellFormat(col2LabelW, rowH, tr("Durata Totale:"), "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 8)
	pdf.SetTextColor(15, 23, 42)
	pdf.CellFormat(col2ValW, rowH, tr(calculateDuration(data.Scan)), "", 1, "L", false, 0, "")

	// Riga 4
	pdf.SetXY(col1X, scopeY+2.5+rowH*3)
	pdf.SetFont("Helvetica", "B", 8)
	pdf.SetTextColor(100, 116, 139)
	pdf.CellFormat(col1LabelW, rowH, tr("Host Censiti:"), "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 8)
	pdf.SetTextColor(15, 23, 42)
	pdf.CellFormat(col1ValW, rowH, tr(getHostScopeSummary(data.Scan)), "", 1, "L", false, 0, "")

	pdf.SetY(scopeY + 29)

	// -------------------------------------------------------------------------
	// SEZIONE 3: SCHEDE TECNICHE DEI FINDINGS (DISTRIBUZIONE OMOGENEA E NATURALE)
	// -------------------------------------------------------------------------
	pdf.SetFont("Helvetica", "B", 12)
	pdf.SetTextColor(prR, prG, prB)
	pdf.CellFormat(0, 6, tr("3. Schede Tecniche dei Findings"), "", 1, "L", false, 0, "")

	if len(data.Findings) == 0 {
		pdf.SetFont("Helvetica", "I", 9.5)
		pdf.SetTextColor(100, 116, 139)
		pdf.CellFormat(0, 8, tr("Nessuna vulnerabilità o servizio anomalo identificato durante la sessione di scansione."), "", 1, "L", false, 0, "")
	} else {
		for i, f := range data.Findings {
			cardH := EstimateFindingCardHeight(f)
			if pdf.GetY()+cardH > 270 {
				pdf.AddPage()
				drawHeader(pdf, tr, tenantName, "Technical Assessment Report", prR, prG, prB, accR, accG, accB)
				pdf.SetY(24)
			}

			drawFindingCard(pdf, tr, i+1, f, prR, prG, prB)
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

// drawHeader disegna l'intestazione standard per le pagine interne con titolo documento dedicato
func drawHeader(pdf *fpdf.Fpdf, tr func(string) string, tenantName, reportTitle string, prR, prG, prB, accR, accG, accB int) {
	pdf.SetY(10)
	pdf.SetFont("Helvetica", "B", 8)
	pdf.SetTextColor(prR, prG, prB)
	pdf.CellFormat(90, 5, tr(strings.ToUpper(tenantName)), "", 0, "L", false, 0, "")

	pdf.SetFont("Helvetica", "", 8)
	pdf.SetTextColor(100, 116, 139)
	pdf.CellFormat(90, 5, tr(reportTitle), "", 1, "R", false, 0, "")

	pdf.SetDrawColor(accR, accG, accB)
	pdf.SetLineWidth(0.4)
	pdf.Line(15, 16, 195, 16)
	pdf.SetLineWidth(0.2)
}

// drawKPIBox disegna un singolo riquadro KPI sintetico
func drawKPIBox(pdf *fpdf.Fpdf, tr func(string) string, x, y, w, h float64, label, value string, r, g, b int) {
	pdf.SetFillColor(248, 250, 252)
	pdf.SetDrawColor(226, 232, 240)
	pdf.RoundedRect(x, y, w, h, 2, "1234", "FD")

	pdf.SetXY(x, y+2.5)
	pdf.SetFont("Helvetica", "B", 6.8)
	pdf.SetTextColor(100, 116, 139)
	pdf.CellFormat(w, 3.5, tr(label), "", 2, "C", false, 0, "")

	pdf.SetFont("Helvetica", "B", 12.5)
	pdf.SetTextColor(r, g, b)
	pdf.CellFormat(w, 6.5, tr(value), "", 0, "C", false, 0, "")
}

// drawFindingCard disegna una scheda tecnica completa per un finding con allineamento millimetrico uniforme
func drawFindingCard(pdf *fpdf.Fpdf, tr func(string) string, index int, f *domain.Finding, prR, prG, prB int) {
	cardY := pdf.GetY()
	cardW := 180.0
	sevR, sevG, sevB := SeverityRGB(f.Severity)

	scoreStr := "N/A"
	if f.CVSSScore != nil {
		scoreStr = fmt.Sprintf("%.1f", *f.CVSSScore)
	}
	badgeStr := fmt.Sprintf("#%d [%s - CVSS %s]", index, f.Severity, scoreStr)
	badgeW := pdf.GetStringWidth(badgeStr) + 3.0

	svcName := CleanServiceName(f)
	targetStr := fmt.Sprintf("%s:%d/%s  -  %s", f.HostIP, f.Port, f.Protocol, svcName)

	hasCVE := f.CVEID != nil && strings.TrimSpace(*f.CVEID) != ""
	hasBanner := f.RawBanner != nil && strings.TrimSpace(*f.RawBanner) != ""
	hasRem := f.Remediation != nil && strings.TrimSpace(*f.Remediation) != ""

	remLines := 0
	remText := ""
	if hasRem {
		remText = *f.Remediation
		remLines = int(math.Ceil(float64(len(remText)) / 85.0))
		if remLines < 1 {
			remLines = 1
		}
	}

	// Calcolo dinamico altezza scheda
	cardH := 8.0
	if hasCVE {
		cardH += 4.5
	}
	if hasBanner {
		cardH += 4.5
	}
	if hasRem {
		cardH += float64(remLines)*4.0 + 1.0
	}
	cardH += 2.0 // padding inferiore

	// Riquadro esterno completo
	pdf.SetFillColor(248, 250, 252)
	pdf.SetDrawColor(226, 232, 240)
	pdf.RoundedRect(15, cardY, cardW, cardH, 1.5, "1234", "FD")

	// Barra colorata laterale sinistra a tutta altezza (senza sporgenze)
	pdf.SetFillColor(sevR, sevG, sevB)
	pdf.Rect(15, cardY, 3.5, cardH, "F")

	// Riga di testata (Badge + Host/Porta/Servizio distanziato)
	pdf.SetXY(22, cardY+2)
	pdf.SetFont("Helvetica", "B", 8.5)
	pdf.SetTextColor(sevR, sevG, sevB)
	pdf.CellFormat(badgeW, 4.5, tr(badgeStr), "", 0, "L", false, 0, "")

	pdf.SetXY(22+badgeW+2.5, cardY+2)
	pdf.SetFont("Helvetica", "B", 8.5)
	pdf.SetTextColor(prR, prG, prB)
	pdf.CellFormat(0, 4.5, tr(targetStr), "", 1, "L", false, 0, "")

	currY := cardY + 7.2

	if hasCVE {
		pdf.SetXY(22, currY)
		pdf.SetFont("Helvetica", "B", 8)
		pdf.SetTextColor(100, 116, 139)
		pdf.CellFormat(20, 4.2, tr("CVE ID:"), "", 0, "L", false, 0, "")
		pdf.SetFont("Helvetica", "B", 8)
		pdf.SetTextColor(220, 38, 38)
		pdf.CellFormat(0, 4.2, tr(*f.CVEID), "", 1, "L", false, 0, "")
		currY += 4.5
	}

	if hasBanner {
		bannerText := *f.RawBanner
		if len(bannerText) > 110 {
			bannerText = bannerText[:107] + "..."
		}
		pdf.SetXY(22, currY)
		pdf.SetFont("Helvetica", "B", 8)
		pdf.SetTextColor(100, 116, 139)
		pdf.CellFormat(20, 4.2, tr("Banner:"), "", 0, "L", false, 0, "")
		pdf.SetFont("Helvetica", "I", 8)
		pdf.SetTextColor(15, 23, 42)
		pdf.CellFormat(0, 4.2, tr(bannerText), "", 1, "L", false, 0, "")
		currY += 4.5
	}

	if hasRem {
		pdf.SetXY(22, currY)
		pdf.SetFont("Helvetica", "B", 8)
		pdf.SetTextColor(100, 116, 139)
		pdf.CellFormat(20, 4.2, tr("Remediation:"), "", 0, "L", false, 0, "")
		pdf.SetXY(42, currY)
		pdf.SetFont("Helvetica", "", 8)
		pdf.SetTextColor(30, 41, 59)
		pdf.MultiCell(150, 4.0, tr(remText), "", "L", false)
	}

	pdf.SetY(cardY + cardH + 2.5)
}

// registerAndDrawLogo gestisce la decodifica base64 del logo e il posizionamento nel PDF
func registerAndDrawLogo(pdf *fpdf.Fpdf, logoStr string, x, y, maxW, maxH float64) bool {
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
	opt := fpdf.ImageOptions{
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
		return ""
	}
	return t.UTC().Format("2006-01-02 15:04:05 UTC")
}

func calculateDuration(s *domain.Scan) string {
	if s == nil {
		return "Consolidata"
	}
	if s.SummaryStats.DurationSeconds > 0 {
		dur := time.Duration(s.SummaryStats.DurationSeconds) * time.Second
		if dur >= time.Minute {
			mins := int(dur.Minutes())
			secs := int(dur.Seconds()) % 60
			return fmt.Sprintf("%dm %ds", mins, secs)
		}
		return fmt.Sprintf("%ds", int(dur.Seconds()))
	}
	if s.StartedAt != nil && s.CompletedAt != nil {
		dur := s.CompletedAt.Sub(*s.StartedAt).Round(time.Second)
		if dur >= time.Minute {
			mins := int(dur.Minutes())
			secs := int(dur.Seconds()) % 60
			return fmt.Sprintf("%dm %ds", mins, secs)
		}
		return dur.String()
	}
	return "Consolidata"
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

// CleanServiceName restituisce una denominazione pulita e leggibile per il software/servizio
func CleanServiceName(f *domain.Finding) string {
	if f == nil {
		return "Unknown Service"
	}

	name := ""
	if f.ServiceName != nil && *f.ServiceName != "" {
		name = strings.TrimSpace(*f.ServiceName)
	}
	ver := ""
	if f.ServiceVersion != nil && *f.ServiceVersion != "" {
		ver = strings.TrimSpace(*f.ServiceVersion)
	}

	// Se il nome è un servizio generico di porta nmap (domain, ftp, ssh, http, etc.)
	// ma disponiamo del RawBanner arricchito con il prodotto reale, usiamo il banner
	if f.RawBanner != nil && strings.TrimSpace(*f.RawBanner) != "" {
		banner := strings.TrimSpace(*f.RawBanner)
		if IsGenericService(name) || name == "" {
			return banner
		}
	}

	if name != "" && ver != "" {
		return fmt.Sprintf("%s %s", name, ver)
	}
	if name != "" {
		return name
	}
	if f.RawBanner != nil && strings.TrimSpace(*f.RawBanner) != "" {
		return strings.TrimSpace(*f.RawBanner)
	}
	return "Unknown Service"
}

func IsGenericService(name string) bool {
	lower := strings.ToLower(name)
	switch lower {
	case "domain", "ftp", "ssh", "http", "https", "msrpc", "netbios-ssn", "microsoft-ds", "telnet":
		return true
	default:
		return false
	}
}

func TruncateString(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	return s
}

func EstimateFindingCardHeight(f *domain.Finding) float64 {
	h := 9.0 // padding base + title bar
	if f.CVEID != nil && strings.TrimSpace(*f.CVEID) != "" {
		h += 4.5
	}
	if f.RawBanner != nil && strings.TrimSpace(*f.RawBanner) != "" {
		h += 4.5
	}
	if f.Remediation != nil && strings.TrimSpace(*f.Remediation) != "" {
		remLines := int(math.Ceil(float64(len(*f.Remediation)) / 85.0))
		if remLines < 1 {
			remLines = 1
		}
		h += float64(remLines)*4.0 + 1.0
	}
	return h
}
