package report

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	fpdf "github.com/go-pdf/fpdf"
	"nextgen-cybersec/internal/domain"
)

type AuditMetadata struct {
	LegalEntity  string    `json:"legal_entity"`
	LegalDetails string    `json:"legal_details"`
	Signatory    string    `json:"signatory"`
	Period       string    `json:"period"`
	AuditDate    time.Time `json:"audit_date"`
}

// GenerateAuditPDF uses the existing report package and one frozen repository
// dataset for every section. It never fetches data while laying out pages.
func GenerateAuditPDF(s domain.ComplianceSnapshot, metadata AuditMetadata) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(18, 18, 18)
	pdf.SetAutoPageBreak(true, 19)
	translate := pdf.UnicodeTranslatorFromDescriptor("")
	line := func(text string) {
		pdf.SetFont("Arial", "", 9)
		pdf.MultiCell(0, 5, translate(text), "", "L", false)
		pdf.Ln(1)
	}
	heading := func(text string) {
		if pdf.GetY() > 250 {
			pdf.AddPage()
		}
		pdf.Ln(3)
		pdf.SetFont("Arial", "B", 13)
		pdf.MultiCell(0, 7, translate(text), "", "L", false)
		pdf.Ln(2)
	}
	pdf.SetFooterFunc(func() {
		pdf.SetY(-13)
		pdf.SetFont("Arial", "", 8)
		pdf.CellFormat(0, 5, fmt.Sprintf("Assessment dossier | %s | Page %d/{nb}", s.Cutoff.Format("2006-01-02"), pdf.PageNo()), "", 0, "C", false, 0, "")
	})
	pdf.AliasNbPages("")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 24)
	pdf.MultiCell(0, 12, "AUDIT EVIDENCE DOSSIER", "", "L", false)
	pdf.Ln(8)
	heading(metadata.LegalEntity)
	line(metadata.LegalDetails)
	line("Audit date: " + metadata.AuditDate.Format("2006-01-02"))
	line("Review period (operator declared): " + metadata.Period)
	if strings.TrimSpace(metadata.Signatory) == "" {
		line("Reviewer/signatory: UNSIGNED")
	} else {
		line("Reviewer/signatory (declared, no digital signature): " + metadata.Signatory)
	}
	line("Scope company: " + s.Scope.CompanyID.String())
	line("Department filter: " + s.Scope.DepartmentID.String())
	line("Data cutoff (repeatable-read): " + s.Cutoff.UTC().Format(time.RFC3339Nano))
	line("Rules: " + s.RuleVersion)
	line("Database snapshot: " + s.SnapshotID + " (MVCC visibility, not a historical replay token)")
	line("This dossier organizes technical and operator-supplied evidence for review. It is not a certificate, a legal applicability determination, or proof that all regulatory requirements are met.")
	heading("Scope, coverage and limitations")
	for _, g := range s.Groups {
		line(g.ID.String() + " | " + g.Name)
	}
	line(fmt.Sprintf("Latest target scans: %d | observations in latest scans: %d | persistent issues: %d", len(s.Scans), len(s.Findings), len(s.Issues)))
	for _, scan := range s.Scans {
		line(fmt.Sprintf("Scan %s | target %s | %s | %s", scan.ID, scan.TargetID, scan.Status, scan.At.Format(time.RFC3339)))
		line("Profile ID: " + scan.ProfileID.String())
		if scan.Authorization == nil {
			line("Historical authorization/parameter snapshot unavailable (legacy or synthetic observation).")
		} else {
			line(fmt.Sprintf("Authorized parameter snapshot: %s | timeout %ds | accepted %s", scan.Authorization.NmapArgs, scan.Authorization.TimeoutSeconds, scan.Authorization.AcceptedAt.Format(time.RFC3339)))
			for _, target := range scan.Authorization.Targets {
				line("Authorized target " + target.ID.String() + " | " + target.Value)
			}
		}
	}
	for _, limitation := range s.Limitations {
		line(limitation)
	}
	pdf.AddPage()
	heading("Control checklist and evidence relevance")
	for _, card := range s.Scorecards {
		if card.Framework == "ALL" {
			continue
		}
		score := "not applicable"
		if card.Score != nil {
			score = fmt.Sprintf("%.1f%%", *card.Score)
		}
		heading(fmt.Sprintf("%s | %s | %s", card.GroupName, card.Framework, score))
		line(fmt.Sprintf("Assessed %d / applicable %d | unknown %d | exclusions %d", card.Assessed, card.Applicable, card.Unknown, card.Excluded))
		for _, control := range card.Controls {
			line(fmt.Sprintf("%s | %s | %s", control.ID, control.Status, control.Title))
			line("Source: " + control.Source + " | curated rule " + s.RuleVersion)
			if control.EvidenceID != "" {
				line("Manual evidence: " + control.EvidenceID)
			}
			if len(control.FindingIDs) > 0 {
				line("Finding references: " + strings.Join(control.FindingIDs, ", "))
			}
			if len(control.IssueIDs) > 0 {
				line("Open issue references: " + strings.Join(control.IssueIDs, ", "))
			}
		}
	}
	if pdf.GetY() > 190 {
		pdf.AddPage()
	}
	heading("Evidence register and remediation plan")
	if len(s.Evidence) == 0 {
		line("No manual control evidence recorded.")
	}
	if len(s.Issues) == 0 {
		line("No persistent issues linked; remediation state is not assessed.")
	}
	for _, e := range s.Evidence {
		line(fmt.Sprintf("%s | group %s | %s | %s | actor %s | %s", e.ID, e.GroupID, e.ControlID, e.Status, e.Actor, e.At.Format(time.RFC3339)))
		line("Reason: " + e.Reason)
		line("Reference: " + e.Reference + " | source version: " + e.SourceVersion)
	}
	for _, issue := range s.Issues {
		heading(issue.AssetName + " | " + issue.Key)
		line(fmt.Sprintf("Issue %s | group %s | %s/%d | %s | owner %s | revision %d", issue.ID, issue.GroupID, issue.Protocol, issue.Port, issue.State, issue.Owner, issue.Version))
		line("First/last observed: " + issue.FirstSeen.Format(time.RFC3339) + " / " + issue.LastSeen.Format(time.RFC3339))
		if issue.Deadline != nil {
			line(fmt.Sprintf("Configured deadline: %s | active breach: %t", issue.Deadline.Format(time.RFC3339), issue.Breached))
		} else {
			line("No SLA policy configured; deadline unknown")
		}
		if issue.MTTRHours != nil {
			line(fmt.Sprintf("Current completed cycle: %.2f hours", *issue.MTTRHours))
		}
		for _, event := range issue.Events {
			line(fmt.Sprintf("Event %d | %s | actor %s | %s | %s", event.ExpectedVersion+1, event.Action, event.Actor, event.At.Format(time.RFC3339), event.Reason))
			if event.Action == "configure" {
				line(fmt.Sprintf("Policy %s v%s | %d hours | owner %s", event.PolicyID, event.PolicyVersion, event.SLAHours, event.Owner))
			}
			if event.Observation != nil {
				line(fmt.Sprintf("Source finding %s | scan %s | observed %s | %s v%s", event.Observation.FindingID, event.Observation.ScanID, event.Observation.At.Format(time.RFC3339), event.Observation.Provenance.Source, event.Observation.Provenance.Version))
				line(fmt.Sprintf("Source host %s | target %s | original group %s | confidence %s | method %s", event.Observation.SourceHost, event.Observation.SourceTargetID, event.Observation.SourceGroupID, event.Observation.Confidence, event.Observation.DetectionMethod))
			}
			if event.Verification != nil {
				line("Verification: " + event.Verification.Provenance.Reference + " | " + event.Verification.Provenance.Source + " v" + event.Verification.Provenance.Version)
			}
		}
	}
	if s.MeanMTTRHours != nil {
		line(fmt.Sprintf("Mean completed-cycle remediation time: %.2f hours across %d cycles; unresolved issues excluded.", *s.MeanMTTRHours, s.CompletedCycles))
	} else {
		line("MTTR not assessed: no verified completed cycle.")
	}
	heading("Source observation register")
	if len(s.Findings) == 0 {
		line("No findings in the latest scoped attempts. This is not proof of complete coverage or compliance.")
	}
	for _, f := range s.Findings {
		line(fmt.Sprintf("Finding %s | scan %s | target %s | group %s | %s %s/%d | %s | confidence %s | method %s", f.ID, f.ScanID, f.TargetID, f.GroupID, f.HostIP, f.Protocol, f.Port, f.Severity, f.Confidence, f.DetectionMethod))
	}
	var output bytes.Buffer
	if err := pdf.Output(&output); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}
