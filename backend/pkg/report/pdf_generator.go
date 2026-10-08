package report

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
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
	GeneratedAt time.Time
}

// PDFGenerator definisce il contratto per la generazione di file PDF di sicurezza
type PDFGenerator interface {
	Generate(data PDFReportData) ([]byte, error)
}

// DefaultPDFGenerator implementa PDFGenerator utilizzando gofpdf
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

// Generate assembla e compila il documento PDF multipagina
func (g *DefaultPDFGenerator) Generate(data PDFReportData) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AliasNbPages("{nb}")
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	// Palette colori del tenant
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

	// Registrazione Footer globale per le pagine successive alla copertina
	pdf.SetFooterFunc(func() {
		if pdf.PageNo() > 1 {
			pdf.SetY(-12)
			pdf.SetFont("Helvetica", "I", 8)
			pdf.SetTextColor(130, 140, 150)
			pdf.CellFormat(0, 8, tr(fmt.Sprintf("%s  |  Report Riservato  |  Pagina %d di {nb}", tenantName, pdf.PageNo())), "", 0, "C", false, 0, "")
		}
	})

	// =========================================================================
	// PAGINA 1: COPERTINA WHITE-LABEL (Cover Page)
	// =========================================================================
	pdf.AddPage()

	// Fascia superiore decorativa a colori primario e accento
	pdf.SetFillColor(prR, prG, prB)
	pdf.Rect(0, 0, 210, 28, "F")
	pdf.SetFillColor(accR, accG, accB)
	pdf.Rect(0, 28, 210, 3, "F")

	// Logo o Box Organizzazione in testata
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

	// Titolo del Report
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

	// Box Informazioni Scansione (Card con sfondo grigio chiaro)
	pdf.Ln(15)
	boxY := pdf.GetY()
	pdf.SetFillColor(248, 250, 252)
	pdf.SetDrawColor(226, 232, 240)
	pdf.RoundedRect(15, boxY, 180, 85, 3, "1234", "FD")

	pdf.SetXY(22, boxY+8)
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

	// Classificazione di sicurezza in basso alla copertina
	pdf.SetY(245)
	pdf.SetFillColor(254, 242, 242)
	pdf.SetDrawColor(254, 202, 202)
	pdf.RoundedRect(15, 245, 180, 15, 2, "1234", "FD")
	pdf.SetXY(15, 248)
	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetTextColor(185, 28, 28)
	pdf.CellFormat(180, 8, tr("CLASSIFICAZIONE: STRETTAMENTE CONFIDENZIALE / USO INTERNO"), "", 0, "C", false, 0, "")

	// =========================================================================
	// PAGINA 2: EXECUTIVE SUMMARY & GRAFICI DI RISCHIO
	// =========================================================================
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

	// Calcolo statistiche
	critCount, highCount, medCount, lowCount, infoCount, maxCVSS, avgCVSS := calculateStats(data.Findings)
	totalFindings := len(data.Findings)

	// 4 Box KPI sintetici
	kpiY := pdf.GetY()
	boxW := 42.0
	spacing := 4.0

	drawKPIBox(pdf, tr, 15+0*(boxW+spacing), kpiY, boxW, 20, "FINDINGS TOTALI", fmt.Sprintf("%d", totalFindings), prR, prG, prB)
	drawKPIBox(pdf, tr, 15+1*(boxW+spacing), kpiY, boxW, 20, "CRITICAL & HIGH", fmt.Sprintf("%d", critCount+highCount), 220, 38, 38)
	drawKPIBox(pdf, tr, 15+2*(boxW+spacing), kpiY, boxW, 20, "MAX CVSS SCORE", fmt.Sprintf("%.1f", maxCVSS), secR, secG, secB)
	drawKPIBox(pdf, tr, 15+3*(boxW+spacing), kpiY, boxW, 20, "MEDIA CVSS", fmt.Sprintf("%.1f", avgCVSS), 100, 116, 139)

	pdf.SetY(kpiY + 26)

	// Sezione Distribuzione Severità con barre grafiche
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

		// Barra orizzontale proporzionale
		barStartX := 60.0
		barMaxW := 120.0
		barW := (pct / 100.0) * barMaxW
		if sr.count > 0 && barW < 3 {
			barW = 3
		}

		currY := pdf.GetY()
		// Sfondo grigio barra
		pdf.SetFillColor(241, 245, 249)
		pdf.Rect(barStartX, currY+1.5, barMaxW, 3.5, "F")
		// Barra colorata
		if barW > 0 {
			pdf.SetFillColor(r, g, b)
			pdf.Rect(barStartX, currY+1.5, barW, 3.5, "F")
		}

		pdf.Ln(7)
	}

	// =========================================================================
	// SEZIONE 2: TARGET SCOPE & OPZIONI DI SCANSIONE
	// =========================================================================
	pdf.Ln(4)
	pdf.SetFont("Helvetica", "B", 14)
	pdf.SetTextColor(prR, prG, prB)
	pdf.CellFormat(0, 8, tr("2. Perimetro di Scansione & Specifiche Tecniche"), "", 1, "L", false, 0, "")

	pdf.SetFillColor(248, 250, 252)
	pdf.SetDrawColor(226, 232, 240)
	scopeY := pdf.GetY()
	pdf.RoundedRect(15, scopeY, 180, 36, 2, "1234", "FD")

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

	pdf.SetY(scopeY + 42)

	// =========================================================================
	// SEZIONE 3: SCHEDE TECNICHE DEI FINDINGS
	// =========================================================================
	pdf.SetFont("Helvetica", "B", 14)
	pdf.SetTextColor(prR, prG, prB)
	pdf.CellFormat(0, 8, tr("3. Schede Tecniche dei Findings"), "", 1, "L", false, 0, "")

	if len(data.Findings) == 0 {
		pdf.SetFont("Helvetica", "I", 10)
		pdf.SetTextColor(100, 116, 139)
		pdf.CellFormat(0, 10, tr("Nessuna vulnerabilità o servizio anomalo identificato durante la sessione di scansione."), "", 1, "L", false, 0, "")
	} else {
		for i, f := range data.Findings {
			// Controllo salto pagina per non spezzare le card
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
		return nil, fmt.Errorf("generazione output PDF fallita: %w", err)
	}

	return buf.Bytes(), nil
}

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
	pdf.SetLineWidth(0.2) // Ripristina spessore predefinito
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

	// Intestazione card
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

	// Linea separatrice inferiore sottile
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

	// Calcolo dimensioni mantenendo le proporzioni
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

