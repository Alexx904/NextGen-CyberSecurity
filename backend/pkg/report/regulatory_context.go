package report

import (
	"fmt"
	fpdf "github.com/go-pdf/fpdf"
	"nextgen-cybersec/internal/domain"
)

func appendRegulatoryContext(pdf *fpdf.Fpdf, tr func(string) string, data PDFReportData) {
	newPage := func() {
		pdf.AddPage()
		pdf.SetXY(15, 24)
		pdf.SetTextColor(15, 23, 42)
		pdf.SetFont("Helvetica", "B", 12)
		pdf.MultiCell(180, 6, tr("Implicazioni normative dei risultati"), "", "L", false)
		pdf.Ln(3)
	}
	text := func(value string, bold bool) {
		if pdf.GetY() > 245 {
			newPage()
		}
		style := ""
		if bold {
			style = "B"
		}
		pdf.SetFont("Helvetica", style, 9)
		pdf.SetTextColor(30, 41, 59)
		pdf.SetX(15)
		pdf.MultiCell(180, 4.5, tr(value), "", "L", false)
		pdf.Ln(2)
	}
	newPage()
	context := domain.BuildRegulatoryContext(domain.Finding{})
	text("Analisi automatica dei finding inclusi nel report. Regole: "+context.RuleVersion, false)
	text(context.Limit, false)
	count := 0
	for _, finding := range data.Findings {
		if finding == nil {
			continue
		}
		context := domain.BuildRegulatoryContext(*finding)
		if len(context.Implications) == 0 {
			continue
		}
		count++
		text(fmt.Sprintf("%s:%d/%s - finding %s - scansione %s", finding.HostIP, finding.Port, finding.Protocol, finding.ID, finding.ScanID), true)
		text(fmt.Sprintf("Confidenza: %s; metodo: %s", valueOrUnknown(string(context.Confidence)), valueOrUnknown(string(context.DetectionMethod))), false)
		for _, implication := range context.Implications {
			text("Dato rilevato: "+implication.Observation, false)
			text("Perché conta: "+implication.Meaning, false)
			text("Cosa fare: "+implication.Action, false)
			for _, reference := range implication.References {
				text(reference.ID+" - "+reference.Title+"\n"+reference.Source, false)
			}
		}
	}
	if count == 0 {
		text("Nessuna implicazione specifica nelle regole installate per i risultati inclusi. Questo non dimostra sicurezza o conformità. L'assenza di finding non prova copertura completa.", false)
	}
}

func valueOrUnknown(value string) string {
	if value == "" {
		return "non disponibile"
	}
	return value
}
