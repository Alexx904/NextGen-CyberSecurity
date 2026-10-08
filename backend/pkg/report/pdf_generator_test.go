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

