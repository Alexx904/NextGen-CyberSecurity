package report

import (
	"bytes"
	"compress/zlib"
	"github.com/google/uuid"
	"io"
	"nextgen-cybersec/internal/domain"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAutomaticRegulatoryPDF(t *testing.T) {
	service, cve := "http", "CVE-2021-41773"
	data := PDFReportData{Scan: &domain.Scan{ID: uuid.New(), Status: domain.ScanStatusCompleted}, Target: &domain.Target{Value: "192.0.2.1"}, GeneratedAt: time.Now(), Findings: []*domain.Finding{{ID: uuid.New(), ScanID: uuid.New(), HostIP: "192.0.2.1", Port: 80, Protocol: "tcp", PortState: "open", ServiceName: &service, CVEID: &cve, Severity: domain.SeverityHigh, Confidence: domain.ConfidenceProbable}}}
	for _, template := range []domain.ReportTemplate{domain.ReportTemplateExecutive, domain.ReportTemplateTechnical} {
		data.Template = template
		pdf, err := NewPDFGenerator().Generate(data)
		if err != nil {
			t.Fatal(err)
		}
		var streams []byte
		remaining := pdf
		for {
			start := bytes.Index(remaining, []byte("stream\n"))
			if start < 0 {
				break
			}
			remaining = remaining[start+7:]
			end := bytes.Index(remaining, []byte("endstream"))
			if end < 0 {
				break
			}
			reader, err := zlib.NewReader(bytes.NewReader(remaining[:end]))
			if err == nil {
				content, _ := io.ReadAll(reader)
				reader.Close()
				streams = append(streams, content...)
			}
			remaining = remaining[end+9:]
		}
		for _, expected := range []string{"Implicazioni normative", "2026-10-19.2", "CVE-2021-41773", "PROBABLE", "eur-lex.europa.eu", "iso.org/standard/27001"} {
			if !bytes.Contains(streams, []byte(expected)) {
				t.Fatalf("%s missing %q", template, expected)
			}
		}
		if directory := os.Getenv("REGULATORY_PDF_DIR"); directory != "" {
			if err := os.WriteFile(filepath.Join(directory, string(template)+".pdf"), pdf, 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
}
