package domain

import (
	"github.com/google/uuid"
	"testing"
	"time"
)

func findComplianceCard(t *testing.T, s *ComplianceSnapshot, group uuid.UUID, framework string) ComplianceScorecard {
	t.Helper()
	for _, c := range s.Scorecards {
		if c.GroupID == group.String() && c.Framework == framework {
			return c
		}
	}
	t.Fatal("missing scorecard")
	return ComplianceScorecard{}
}
func TestComplianceUnknownDenominatorsExclusionsAndGaps(t *testing.T) {
	g := uuid.New()
	s := &ComplianceSnapshot{Groups: []ComplianceGroup{{ID: g}}, TargetCounts: map[string]int{g.String(): 1}, Scans: []ComplianceScan{{GroupID: g, Status: ScanStatusCompleted}}}
	BuildComplianceScorecards(s)
	card := findComplianceCard(t, s, g, "ISO27001")
	if card.Applicable != 3 || card.Unknown != 3 || card.Score == nil || *card.Score != 0 {
		t.Fatal("absence misreported as compliance", card)
	}
	s.Evidence = []ControlEvidence{{ID: uuid.New(), GroupID: g, ControlID: "ISO-A.8.8", Status: "satisfied"}, {ID: uuid.New(), GroupID: g, ControlID: "ISO-A.8.20", Status: "partial"}, {ID: uuid.New(), GroupID: g, ControlID: "ISO-A.8.24", Status: "not_applicable"}}
	BuildComplianceScorecards(s)
	card = findComplianceCard(t, s, g, "ISO27001")
	if card.Excluded != 1 || card.Applicable != 2 || card.Assessed != 2 || *card.Score != 75 {
		t.Fatal("wrong denominator/partial credit", card)
	}
	cve := "CVE-2026-1234"
	s.Findings = []ComplianceFinding{{Finding: Finding{ID: uuid.New(), CVEID: &cve, PortState: "open"}, GroupID: g}}
	BuildComplianceScorecards(s)
	card = findComplianceCard(t, s, g, "ISO27001")
	if card.Controls[0].Status != "not_satisfied" || card.Controls[1].Status != "not_satisfied" || *card.Score != 0 {
		t.Fatal("manual evidence hid technical gap")
	}
	s.Findings = nil
	s.Scans[0].Status = ScanStatusFailed
	BuildComplianceScorecards(s)
	card = findComplianceCard(t, s, g, "ISO27001")
	if card.Unknown != 2 || *card.Score != 0 {
		t.Fatal("incomplete coverage claimed satisfied")
	}
}
func TestComplianceCompanyRollupAndRetainedIssue(t *testing.T) {
	root, child := uuid.New(), uuid.New()
	s := &ComplianceSnapshot{Scope: ComplianceScope{CompanyID: root}, Groups: []ComplianceGroup{{ID: root}, {ID: child, ParentID: &root}}, Issues: []ComplianceIssue{{ID: uuid.New(), GroupID: child, Key: "service:telnet", State: IssueMitigationApplied}}}
	BuildComplianceScorecards(s)
	for _, g := range []uuid.UUID{root, child} {
		card := findComplianceCard(t, s, g, "ISO27001")
		if card.Controls[2].Status != "not_satisfied" || len(card.Controls[2].IssueIDs) != 1 {
			t.Fatal("rollup lost unresolved historical gap", card)
		}
	}
	s.Issues[0].State = IssueCorrectedVerified
	BuildComplianceScorecards(s)
	if card := findComplianceCard(t, s, root, "ISO27001"); card.Unknown != 3 {
		t.Fatal("verified issue implied full control compliance")
	}
	if len(ComplianceControls()) != 28 {
		t.Fatal("unexpected catalogue denominator")
	}
	s.TargetCounts = map[string]int{child.String(): 1}
	s.Scans = []ComplianceScan{{GroupID: child, Status: ScanStatusCompleted}}
	for _, key := range []string{"ISO-A.8.8", "ISO-A.8.20", "ISO-A.8.24"} {
		s.Evidence = append(s.Evidence, ControlEvidence{GroupID: child, ControlID: key, Status: "satisfied"})
	}
	BuildComplianceScorecards(s)
	if card := findComplianceCard(t, s, root, "ISO27001"); card.Score == nil || *card.Score != 100 {
		t.Fatal("empty company container counted as unscanned asset", card)
	}
}
func TestComplianceReplayConfigurationAndCycleMean(t *testing.T) {
	i, o, actor := issueFixture(t)
	dto := &ComplianceIssue{ID: i.ID(), AssetID: o.Identity.AssetID, Key: o.Identity.Key, Port: o.Identity.Port, Protocol: o.Identity.Protocol}
	dto.Events = []IssueCommand{{Action: "observe", Actor: actor, At: o.At, ExpectedVersion: 0, Observation: &o}, {Action: "configure", Actor: actor, At: o.At.Add(time.Hour), ExpectedVersion: 1, Reason: "Explicit policy", SLAHours: 24, PolicyID: "internal", PolicyVersion: "2", Owner: actor}, {Action: "correction_declared", Actor: actor, At: o.At.Add(2 * time.Hour), ExpectedVersion: 2, Reason: "patched"}, {Action: "verify", Actor: actor, At: o.At.Add(4 * time.Hour), ExpectedVersion: 3, Reason: "verified", Verification: &IssueVerification{Identity: o.Identity, At: o.At.Add(4 * time.Hour), CompleteCoverage: true, NotPresent: true, Provenance: IssueProvenance{"retest", "1", "test/evidence"}}}}
	if _, err := ReplayIssue(o.Identity.TenantID, dto, o.At.Add(48*time.Hour)); err != nil {
		t.Fatal(err)
	}
	if dto.MTTRHours == nil || *dto.MTTRHours != 4 || dto.Breached || dto.Version != 4 || dto.Owner != actor {
		t.Fatal("replay changed policy/time semantics", dto)
	}
	s := &ComplianceSnapshot{Issues: []ComplianceIssue{*dto}}
	BuildComplianceScorecards(s)
	if s.MeanMTTRHours == nil || *s.MeanMTTRHours != 4 || s.CompletedCycles != 1 {
		t.Fatal("incorrect completed-cycle mean")
	}
	dto.Events[1].ExpectedVersion = 8
	if _, err := ReplayIssue(o.Identity.TenantID, dto, o.At); err == nil {
		t.Fatal("out-of-order event accepted")
	}
}
