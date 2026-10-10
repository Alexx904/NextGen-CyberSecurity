package report_test

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"

	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/pkg/report"
)

func createTinyPNG() string {
	img := image.NewRGBA(image.Rect(0, 0, 20, 20))
	for x := 0; x < 20; x++ {
		for y := 0; y < 20; y++ {
			img.Set(x, y, color.RGBA{R: 37, G: 99, B: 235, A: 255})
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
}

func TestUnicode(t *testing.T) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	tr := pdf.UnicodeTranslatorFromDescriptor("")
	got := tr("vulnerabilità")
	t.Logf("gofpdf tr: len %d, hex: %x, str: %s", len(got), []byte(got), got)
}

func TestHexToRGB(t *testing.T) {
	cases := []struct {
		hex   string
		expR  int
		expG  int
		expB  int
	}{
		{"#ffffff", 255, 255, 255},
		{"#000000", 0, 0, 0},
		{"#fff", 255, 255, 255},
		{"#0f172a", 15, 23, 42},
		{"#2563eb", 37, 99, 235},
		{"invalid", 15, 23, 42}, // Default fallback
		{"", 15, 23, 42},
		{"#12", 15, 23, 42},
	}

	for _, c := range cases {
		r, g, b := report.HexToRGB(c.hex)
		if r != c.expR || g != c.expG || b != c.expB {
			t.Errorf("HexToRGB(%s): atteso (%d,%d,%d), ottenuto (%d,%d,%d)", c.hex, c.expR, c.expG, c.expB, r, g, b)
		}
	}
}

func TestSeverityRGB(t *testing.T) {
	severities := []domain.SeverityLevel{
		domain.SeverityCritical,
		domain.SeverityHigh,
		domain.SeverityMedium,
		domain.SeverityLow,
		domain.SeverityInfo,
	}

	for _, s := range severities {
		r, g, b := report.SeverityRGB(s)
		if r < 0 || r > 255 || g < 0 || g > 255 || b < 0 || b > 255 {
			t.Errorf("colore fuori range per severità %s: (%d,%d,%d)", s, r, g, b)
		}
	}
}

func TestPDFGenerator_Generate_Success(t *testing.T) {
	gen := report.NewPDFGenerator()

	now := time.Now().UTC()
	startedAt := now.Add(-10 * time.Minute)
	completedAt := now

	logoData := createTinyPNG()
	tenant := &domain.Tenant{
		ID:      uuid.New(),
		Name:    "Acme Cyber Security Corp",
		Slug:    "acme-cyber",
		LogoURL: &logoData,
		BrandColors: domain.BrandColors{
			Primary:   "#0f172a",
			Secondary: "#2563eb",
			Accent:    "#38bdf8",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	target := &domain.Target{
		ID:        uuid.New(),
		TenantID:  tenant.ID,
		Type:      domain.TargetTypeIP,
		Value:     "192.168.1.50",
		CreatedAt: now,
		UpdatedAt: now,
	}

	profile := &domain.ScanProfile{
		ID:          uuid.New(),
		Name:        "Standard Service Scan",
		NmapArgs:    "-sV -T4",
		Description: func(s string) *string { return &s }("Scansione porte standard con rilevamento servizi"),
	}

	scan := &domain.Scan{
		ID:          uuid.New(),
		TenantID:    tenant.ID,
		TargetID:    target.ID,
		ProfileID:   profile.ID,
		TriggerType: domain.TriggerManual,
		Status:      domain.ScanStatusCompleted,
		StartedAt:   &startedAt,
		CompletedAt: &completedAt,
		CreatedAt:   startedAt,
		UpdatedAt:   completedAt,
	}

	cve1 := "CVE-2021-41773"
	cvss1 := 9.8
	svc1 := "http"
	ver1 := "Apache 2.4.49"
	banner1 := "Apache/2.4.49 (Unix) OpenSSL/1.1.1l"
	rem1 := "Aggiornare immediatamente alla versione più recente; verificare l'integrità e priorità di remediation per tutte le vulnerabilità critiche."

	cve2 := "CVE-2020-0796"
	cvss2 := 8.5
	svc2 := "microsoft-ds"
	ver2 := "Windows 10 SMBv3"
	rem2 := "Disabilitare la compressione SMBv3 e applicare la patch KB4551762 per massima affidabilità."

	findings := []*domain.Finding{
		{
			ID:             uuid.New(),
			ScanID:         scan.ID,
			TenantID:       tenant.ID,
			HostIP:         "192.168.1.50",
			Port:           80,
			Protocol:       "tcp",
			ServiceName:    &svc1,
			ServiceVersion: &ver1,
			CVEID:          &cve1,
			CVSSScore:      &cvss1,
			Severity:       domain.SeverityCritical,
			RawBanner:      &banner1,
			Remediation:    &rem1,
			CreatedAt:      now,
		},
		{
			ID:             uuid.New(),
			ScanID:         scan.ID,
			TenantID:       tenant.ID,
			HostIP:         "192.168.1.50",
			Port:           445,
			Protocol:       "tcp",
			ServiceName:    &svc2,
			ServiceVersion: &ver2,
			CVEID:          &cve2,
			CVSSScore:      &cvss2,
			Severity:       domain.SeverityHigh,
			Remediation:    &rem2,
			CreatedAt:      now,
		},
		{
			ID:       uuid.New(),
			ScanID:   scan.ID,
			TenantID: tenant.ID,
			HostIP:   "192.168.1.50",
			Port:     22,
			Protocol: "tcp",
			Severity: domain.SeverityInfo,
		},
	}

	reportData := report.PDFReportData{
		Tenant:      tenant,
		Scan:        scan,
		Target:      target,
		Profile:     profile,
		Findings:    findings,
		GeneratedAt: now,
	}

	pdfBytes, err := gen.Generate(reportData)
	if err != nil {
		t.Fatalf("generazione PDF fallita: %v", err)
	}

	if len(pdfBytes) < 1000 {
		t.Fatalf("dimensione PDF troppo piccola (%d byte)", len(pdfBytes))
	}

	// Verifica intestazione PDF conforme (%PDF-)
	if !bytes.HasPrefix(pdfBytes, []byte("%PDF-")) {
		t.Errorf("il file generato non ha l'intestazione standard PDF")
	}
}

func TestPDFGenerator_Generate_EmptyFindingsAndNoLogo(t *testing.T) {
	gen := report.NewPDFGenerator()
	now := time.Now().UTC()

	reportData := report.PDFReportData{
		Tenant: &domain.Tenant{
			ID:          uuid.New(),
			Name:        "Test Tenant Without Logo",
			BrandColors: domain.DefaultBrandColors(),
		},
		Scan: &domain.Scan{
			ID:        uuid.New(),
			Status:    domain.ScanStatusCompleted,
			CreatedAt: now,
		},
		Target: &domain.Target{
			Value: "10.0.0.1",
			Type:  domain.TargetTypeIP,
		},
		Profile: &domain.ScanProfile{
			Name:     "Fast Discovery",
			NmapArgs: "-F -T4",
		},
		Findings:    []*domain.Finding{},
		GeneratedAt: now,
	}

	pdfBytes, err := gen.Generate(reportData)
	if err != nil {
		t.Fatalf("generazione PDF fallita su report vuoto: %v", err)
	}

	if !bytes.HasPrefix(pdfBytes, []byte("%PDF-")) {
		t.Errorf("il file generato non ha l'intestazione standard PDF")
	}
}

func TestPDFGenerator_SanitizesEmptyFindings(t *testing.T) {
	gen := report.NewPDFGenerator()
	now := time.Now().UTC()

	validFinding := &domain.Finding{
		ID:        uuid.New(),
		ScanID:    uuid.New(),
		TenantID:  uuid.New(),
		HostIP:    "10.0.0.1",
		Port:      80,
		Protocol:  "tcp",
		Severity:  domain.SeverityMedium,
		CreatedAt: now,
	}

	reportData := report.PDFReportData{
		Tenant: &domain.Tenant{
			ID:          uuid.New(),
			Name:        "Test Tenant Sanitization",
			BrandColors: domain.DefaultBrandColors(),
		},
		Scan: &domain.Scan{
			ID:     uuid.New(),
			Status: domain.ScanStatusCompleted,
			SummaryStats: domain.ScanSummaryStats{
				TotalHosts: 10,
				UpHosts:    2,
				DownHosts:  8,
			},
			CreatedAt: now,
		},
		Target: &domain.Target{
			Value: "10.0.0.0/28",
			Type:  domain.TargetTypeRange,
		},
		Profile: &domain.ScanProfile{
			Name:     "Subnet Scan",
			NmapArgs: "-sV",
		},
		Findings: []*domain.Finding{
			nil,
			validFinding,
			{HostIP: ""},
			{HostIP: "0.0.0.0"},
			nil,
		},
		GeneratedAt: now,
	}

	pdfBytes, err := gen.Generate(reportData)
	if err != nil {
		t.Fatalf("generazione PDF fallita su findings da sanitizzare: %v", err)
	}

	if !bytes.HasPrefix(pdfBytes, []byte("%PDF-")) {
		t.Errorf("il file generato non ha l'intestazione standard PDF")
	}
}

func TestPDFGenerator_Generate_ExecutiveTemplate(t *testing.T) {
	gen := report.NewPDFGenerator()
	now := time.Now().UTC()
	startedAt := now.Add(-5 * time.Minute)
	completedAt := now

	logoData := createTinyPNG()
	tenant := &domain.Tenant{
		ID:      uuid.New(),
		Name:    "Global Executive Security Corp",
		Slug:    "global-exec",
		LogoURL: &logoData,
		BrandColors: domain.BrandColors{
			Primary:   "#1e293b",
			Secondary: "#0284c7",
			Accent:    "#38bdf8",
		},
	}

	target := &domain.Target{
		ID:       uuid.New(),
		TenantID: tenant.ID,
		Value:    "192.168.10.0/24",
		Type:     domain.TargetTypeCIDR,
	}

	profile := &domain.ScanProfile{
		ID:       uuid.New(),
		Name:     "Executive Baseline",
		NmapArgs: "-sV -T4",
	}

	scan := &domain.Scan{
		ID:          uuid.New(),
		TenantID:    tenant.ID,
		TargetID:    target.ID,
		ProfileID:   profile.ID,
		Status:      domain.ScanStatusCompleted,
		StartedAt:   &startedAt,
		CompletedAt: &completedAt,
		SummaryStats: domain.ScanSummaryStats{
			TotalHosts:    254,
			UpHosts:       12,
			DownHosts:     242,
			CriticalCount: 1,
			HighCount:     2,
			MediumCount:   4,
			LowCount:      5,
		},
	}

	cveCrit := "CVE-2023-38606"
	cvssCrit := 9.8
	svcHttp := "http"
	remCrit := "Applicare la patch di sicurezza vendor per prevenire l'esecuzione remota di codice non autenticato."

	findings := []*domain.Finding{
		{
			ID:          uuid.New(),
			ScanID:      scan.ID,
			TenantID:    tenant.ID,
			HostIP:      "192.168.10.5",
			Port:        80,
			Protocol:    "tcp",
			ServiceName: &svcHttp,
			CVEID:       &cveCrit,
			CVSSScore:   &cvssCrit,
			Severity:    domain.SeverityCritical,
			Remediation: &remCrit,
			CreatedAt:   now,
		},
	}

	reportData := report.PDFReportData{
		Tenant:      tenant,
		Scan:        scan,
		Target:      target,
		Profile:     profile,
		Findings:    findings,
		Template:    domain.ReportTemplateExecutive,
		GeneratedAt: now,
	}

	pdfBytes, err := gen.Generate(reportData)
	if err != nil {
		t.Fatalf("generazione PDF Executive fallita: %v", err)
	}

	if len(pdfBytes) < 1000 {
		t.Fatalf("dimensione PDF Executive troppo piccola (%d byte)", len(pdfBytes))
	}

	if !bytes.HasPrefix(pdfBytes, []byte("%PDF-")) {
		t.Errorf("intestazione PDF non conforme")
	}
}

func TestPDFGenerator_Generate_ExecutiveTemplate_NoFindings(t *testing.T) {
	gen := report.NewPDFGenerator()
	now := time.Now().UTC()

	reportData := report.PDFReportData{
		Tenant: &domain.Tenant{
			ID:          uuid.New(),
			Name:        "Clean Infrastructure Tenant",
			BrandColors: domain.DefaultBrandColors(),
		},
		Scan: &domain.Scan{
			ID:     uuid.New(),
			Status: domain.ScanStatusCompleted,
			SummaryStats: domain.ScanSummaryStats{
				TotalHosts: 1,
				UpHosts:    1,
			},
		},
		Target: &domain.Target{
			Value: "10.10.10.1",
			Type:  domain.TargetTypeIP,
		},
		Profile: &domain.ScanProfile{
			Name: "Quick Audit",
		},
		Findings:    []*domain.Finding{},
		Template:    domain.ReportTemplateExecutive,
		GeneratedAt: now,
	}

	pdfBytes, err := gen.Generate(reportData)
	if err != nil {
		t.Fatalf("generazione PDF Executive senza finding fallita: %v", err)
	}

	if !bytes.HasPrefix(pdfBytes, []byte("%PDF-")) {
		t.Errorf("intestazione PDF non conforme")
	}
}

func TestPDFGenerator_Generate_TechnicalTemplate(t *testing.T) {
	gen := report.NewPDFGenerator()
	now := time.Now().UTC()

	banner := "Apache/2.4.52 (Ubuntu)"
	reportData := report.PDFReportData{
		Tenant: &domain.Tenant{
			ID:          uuid.New(),
			Name:        "Technical SecOps Team",
			BrandColors: domain.DefaultBrandColors(),
		},
		Scan: &domain.Scan{
			ID:     uuid.New(),
			Status: domain.ScanStatusCompleted,
		},
		Target: &domain.Target{
			Value: "172.16.0.10",
			Type:  domain.TargetTypeIP,
		},
		Profile: &domain.ScanProfile{
			Name:     "Full Deep Scan",
			NmapArgs: "-sV -p- -T4",
		},
		Findings: []*domain.Finding{
			{
				ID:        uuid.New(),
				HostIP:    "172.16.0.10",
				Port:      8080,
				Protocol:  "tcp",
				Severity:  domain.SeverityMedium,
				RawBanner: &banner,
				CreatedAt: now,
			},
		},
		Template:    domain.ReportTemplateTechnical,
		GeneratedAt: now,
	}

	pdfBytes, err := gen.Generate(reportData)
	if err != nil {
		t.Fatalf("generazione PDF Technical fallita: %v", err)
	}

	if !bytes.HasPrefix(pdfBytes, []byte("%PDF-")) {
		t.Errorf("intestazione PDF non conforme")
	}
}

func TestCleanServiceName(t *testing.T) {
	svcDomain := "domain"
	verDomain := "2.89"
	bannerDnsmasq := "dnsmasq 2.89"

	f1 := &domain.Finding{
		ServiceName:    &svcDomain,
		ServiceVersion: &verDomain,
		RawBanner:      &bannerDnsmasq,
	}
	if got := report.CleanServiceName(f1); got != "dnsmasq 2.89" {
		t.Errorf("CleanServiceName(f1) = %q, atteso %q", got, "dnsmasq 2.89")
	}

	svcFtp := "ftp"
	verFtp := "2.0.8 or later"
	bannerVsftpd := "vsftpd 2.0.8 or later"
	f2 := &domain.Finding{
		ServiceName:    &svcFtp,
		ServiceVersion: &verFtp,
		RawBanner:      &bannerVsftpd,
	}
	if got := report.CleanServiceName(f2); got != "vsftpd 2.0.8 or later" {
		t.Errorf("CleanServiceName(f2) = %q, atteso %q", got, "vsftpd 2.0.8 or later")
	}

	prodApache := "Apache httpd"
	verApache := "2.4.49"
	f3 := &domain.Finding{
		ServiceName:    &prodApache,
		ServiceVersion: &verApache,
	}
	if got := report.CleanServiceName(f3); got != "Apache httpd 2.4.49" {
		t.Errorf("CleanServiceName(f3) = %q, atteso %q", got, "Apache httpd 2.4.49")
	}
}

func TestTruncateString(t *testing.T) {
	if got := report.TruncateString("Corto", 10); got != "Corto" {
		t.Errorf("TruncateString short = %q, atteso %q", got, "Corto")
	}
	if got := report.TruncateString("Audit Consolidato Perimetrale Aziendale", 20); got != "Audit Consolidato..." {
		t.Errorf("TruncateString long = %q, atteso %q", got, "Audit Consolidato...")
	}
}

func TestPDFGenerator_Executive_RealisticScenario(t *testing.T) {
	gen := report.NewPDFGenerator()
	now := time.Now().UTC()
	startedAt := now.Add(-30 * time.Minute)

	score75 := 7.5
	score50 := 5.0
	cve1 := "CVE-2023-50387"
	cve2 := "CVE-2023-28450"
	rem := "Aggiornare dnsmasq alla versione più recente supportata per mitigare la vulnerabilità CVE-2023-50387. Isolare o segmentare l'accesso alla porta se l'esposizione non è strettamente necessaria."

	banner := "dnsmasq 2.89"
	svc := "dnsmasq"
	ver := "2.89"

	var findings []*domain.Finding
	// 5 High findings
	for i := 0; i < 5; i++ {
		cve := cve1
		if i%2 == 1 {
			cve = cve2
		}
		findings = append(findings, &domain.Finding{
			ID:             uuid.New(),
			HostIP:         "192.168.1.1",
			Port:           53,
			Protocol:       "tcp",
			ServiceName:    &svc,
			ServiceVersion: &ver,
			RawBanner:      &banner,
			CVEID:          &cve,
			CVSSScore:      &score75,
			Severity:       domain.SeverityHigh,
			Remediation:    &rem,
			CreatedAt:      now,
		})
	}
	// 7 Medium findings
	for i := 0; i < 7; i++ {
		findings = append(findings, &domain.Finding{
			ID:          uuid.New(),
			HostIP:      "192.168.1.10",
			Port:        135,
			Protocol:    "tcp",
			ServiceName: &svc,
			CVSSScore:   &score50,
			Severity:    domain.SeverityMedium,
			Remediation: &rem,
			CreatedAt:   now,
		})
	}
	// 14 Info findings
	for i := 0; i < 14; i++ {
		findings = append(findings, &domain.Finding{
			ID:        uuid.New(),
			HostIP:    "192.168.1.10",
			Port:      80 + i,
			Protocol:  "tcp",
			Severity:  domain.SeverityInfo,
			CreatedAt: now,
		})
	}

	reportData := report.PDFReportData{
		Tenant: &domain.Tenant{
			ID:          uuid.New(),
			Name:        "Test Organization",
			BrandColors: domain.DefaultBrandColors(),
		},
		Scan: &domain.Scan{
			ID:          uuid.New(),
			Status:      domain.ScanStatusCompleted,
			TriggerType: domain.TriggerManual,
			StartedAt:   &startedAt,
			CompletedAt: &now,
			SummaryStats: domain.ScanSummaryStats{
				TotalHosts: 9,
				UpHosts:    9,
				HighCount:  5,
			},
		},
		Target: &domain.Target{
			Value: "Casa",
			Type:  domain.TargetType("COMPANY_SCOPE"),
		},
		Profile: &domain.ScanProfile{
			Name:     "Audit Consolidato Perimetrale Aziendale",
			NmapArgs: "Consolidato su 9 target dell'azienda",
		},
		Findings:    findings,
		Template:    domain.ReportTemplateExecutive,
		GeneratedAt: now,
	}

	pdfBytes, err := gen.Generate(reportData)
	if err != nil {
		t.Fatalf("generazione Executive PDF scenario realistico fallita: %v", err)
	}
	if !bytes.HasPrefix(pdfBytes, []byte("%PDF-")) {
		t.Errorf("intestazione PDF non conforme")
	}

	// Stesso per template Technical
	reportData.Template = domain.ReportTemplateTechnical
	pdfTechBytes, err := gen.Generate(reportData)
	if err != nil {
		t.Fatalf("generazione Technical PDF scenario realistico fallita: %v", err)
	}
	if !bytes.HasPrefix(pdfTechBytes, []byte("%PDF-")) {
		t.Errorf("intestazione Technical PDF non conforme")
	}
}



