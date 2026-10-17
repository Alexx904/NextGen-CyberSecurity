package report

import (
	"bytes"
	"fmt"
	fpdf "github.com/go-pdf/fpdf"
	"testing"
)

func TestDepartmentInventoryIncludesUnscannedTargetsAndPaginates(t *testing.T) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetCompression(false)
	ips := []string{}
	for i := 0; i < 120; i++ {
		ips = append(ips, fmt.Sprintf("10.0.0.%d", i))
	}
	appendDepartmentInventory(pdf, pdf.UnicodeTranslatorFromDescriptor(""), []DepartmentReportSection{{Name: "Finance", Targets: ips}, {Name: "IT", Targets: []string{"192.0.2.1 - non scansionato"}}})
	var out bytes.Buffer
	if err := pdf.Output(&out); err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{"Finance", "10.0.0.119", "IT", "192.0.2.1 - non scansionato", "continua"} {
		if !bytes.Contains(out.Bytes(), []byte(text)) {
			t.Fatalf("PDF omitted %s", text)
		}
	}
	if pdf.PageNo() < 3 {
		t.Fatalf("expected multiple pages, got %d", pdf.PageNo())
	}
}
