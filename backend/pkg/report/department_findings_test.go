package report

import (
	"bytes"
	"compress/zlib"
	"io"
	"strings"
	"testing"
	"time"

	fpdf "github.com/go-pdf/fpdf"
	"nextgen-cybersec/internal/domain"
)

// Inspect actual generated content streams, without claiming a visual render.
func pdfContent(t *testing.T, data []byte) string {
	t.Helper()
	var content strings.Builder
	for _, chunk := range bytes.Split(data, []byte("\nstream\n"))[1:] {
		end := bytes.Index(chunk, []byte("\nendstream"))
		if end < 0 {
			continue
		}
		r, err := zlib.NewReader(bytes.NewReader(chunk[:end]))
		if err != nil {
			continue
		}
		decoded, err := io.ReadAll(r)
		r.Close()
		if err != nil {
			t.Fatal(err)
		}
		content.Write(decoded)
	}
	return content.String()
}

func TestCompanyPDFFindingsAreGroupedInBothTemplates(t *testing.T) {
	cveA, cveB := "CVE-2026-10001", "CVE-2026-10002"
	a := &domain.Finding{HostIP: "10.0.0.1", Port: 443, Protocol: "tcp", Severity: domain.SeverityHigh, CVEID: &cveA}
	b := &domain.Finding{HostIP: "10.0.0.1", Port: 443, Protocol: "tcp", Severity: domain.SeverityHigh, CVEID: &cveB}
	for _, template := range []domain.ReportTemplate{domain.ReportTemplateTechnical, domain.ReportTemplateExecutive} {
		t.Run(string(template), func(t *testing.T) {
			data := PDFReportData{
				Tenant: &domain.Tenant{Name: "Test"}, Scan: &domain.Scan{}, Target: &domain.Target{Value: "Company"}, Profile: &domain.ScanProfile{}, GeneratedAt: time.Now(),
				Template: template, Findings: []*domain.Finding{a, b}, GroupFindingsByDepartment: true,
				Departments: []DepartmentReportSection{{Name: "Finance", Findings: []*domain.Finding{a, nil, {HostIP: "0.0.0.0"}}}, {Name: "IT", Findings: []*domain.Finding{b}}, {Name: "Unscanned"}},
			}
			pdf, err := NewPDFGenerator().Generate(data)
			if err != nil {
				t.Fatal(err)
			}
			content := pdfContent(t, pdf)
			finance, first, it, second := strings.Index(content, "Reparto: Finance"), strings.Index(content, cveA), strings.Index(content, "Reparto: IT"), strings.Index(content, cveB)
			if finance < 0 || first < finance || it < first || second < it {
				t.Fatalf("CVE cards are not within their department sections: Finance=%d CVE-A=%d IT=%d CVE-B=%d", finance, first, it, second)
			}
			if !strings.Contains(content, "Nessun finding disponibile") || strings.Contains(content, "0.0.0.0:0") {
				t.Fatal("empty department or invalid finding was rendered incorrectly")
			}
			data.GroupFindingsByDepartment = false
			data.Departments = nil
			plain, err := NewPDFGenerator().Generate(data)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(pdfContent(t, plain), "Reparto:") {
				t.Fatal("ungrouped reports changed layout")
			}
		})
	}
}

func TestDepartmentFindingPaginationRepeatsHeadingAndExecutiveLimitIsLocal(t *testing.T) {
	for _, executive := range []bool{false, true} {
		pdf := fpdf.New("P", "mm", "A4", "")
		pdf.SetMargins(15, 15, 15)
		pdf.SetCompression(false)
		newPage := func() { pdf.AddPage(); pdf.SetY(24) }
		newPage()
		sections := []DepartmentReportSection{{Name: "Finance"}, {Name: "IT"}}
		for i := range sections {
			for n := 0; n < 25; n++ {
				cve := sections[i].Name + "-CVE-" + string(rune('A'+n))
				sections[i].Findings = append(sections[i].Findings, &domain.Finding{HostIP: "10.0.0.1", CVEID: &cve, Severity: domain.SeverityHigh})
			}
		}
		drawDepartmentFindings(pdf, pdf.UnicodeTranslatorFromDescriptor(""), sections, executive, newPage, 15, 23, 42)
		var out bytes.Buffer
		if err := pdf.Output(&out); err != nil {
			t.Fatal(err)
		}
		content := out.String()
		for _, dept := range []string{"Finance", "IT"} {
			if !strings.Contains(content, dept+"-CVE-A") || !strings.Contains(content, dept+"-CVE-E") {
				t.Fatal("one department lost its priority cards")
			}
			if executive == strings.Contains(content, dept+"-CVE-Y") {
				t.Fatal("wrong per-department card limit")
			}
		}
		if !executive && (!strings.Contains(content, "continua") || pdf.PageNo() < 4) {
			t.Fatal("long technical sections did not paginate with department headings")
		}
	}
}
