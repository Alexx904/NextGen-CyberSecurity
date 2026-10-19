package domain

import "strings"

type ComplianceControl struct {
	ID        string `json:"id"`
	Framework string `json:"framework"`
	Title     string `json:"title"`
	Source    string `json:"source"`
	Mapping   string `json:"mapping"`
}

// Curated relevance rules, not a statement of legal applicability. Organizational
// controls deliberately have no positive inference from network scans.
func ComplianceControls() []ComplianceControl {
	var controls []ComplianceControl
	add := func(id, framework, title, source, mapping string) {
		controls = append(controls, ComplianceControl{id, framework, title, source, mapping})
	}
	nis := "https://eur-lex.europa.eu/eli/dir/2022/2555/oj/eng"
	for index, title := range []string{"Risk analysis and information security policy", "Incident handling", "Business continuity and crisis management", "Supply chain security", "Secure acquisition, development and vulnerability handling", "Effectiveness assessment", "Cyber hygiene and training", "Cryptography policy", "Access control and asset management", "Multi-factor authentication and secure communications"} {
		mapping := "manual"
		if index == 4 {
			mapping = "vulnerability"
		}
		if index == 7 {
			mapping = "cleartext"
		}
		if index == 8 {
			mapping = "network"
		}
		add("NIS2-21.2"+string(rune('a'+index)), "NIS2", title, nis, mapping)
	}
	dora := "https://eur-lex.europa.eu/eli/reg/2022/2554/oj/eng"
	for _, c := range [][3]string{{"5", "Governance and organization", "manual"}, {"6", "ICT risk management framework", "manual"}, {"7", "ICT systems and tools", "manual"}, {"8", "Identification and inventory", "network"}, {"9", "Protection and prevention", "vulnerability"}, {"10", "Detection", "manual"}, {"11", "Response and recovery", "manual"}, {"12", "Backup and restoration", "manual"}, {"13", "Learning and evolving", "manual"}, {"14", "Communication", "manual"}, {"16", "Simplified ICT risk framework applicability", "manual"}, {"24", "Digital operational resilience testing programme", "manual"}, {"25", "Testing ICT tools and systems", "vulnerability"}, {"26", "Advanced testing applicability", "manual"}, {"27", "Tester requirements", "manual"}} {
		add("DORA-"+c[0], "DORA", c[1], dora, c[2])
	}
	iso := "https://www.iso.org/standard/27001"
	add("ISO-A.8.8", "ISO27001", "Technical vulnerability management", iso, "vulnerability")
	add("ISO-A.8.20", "ISO27001", "Network security", iso, "network")
	add("ISO-A.8.24", "ISO27001", "Use of cryptography", iso, "cleartext")
	return controls
}

type ComplianceControlResult struct {
	ComplianceControl
	Status     string   `json:"status"`
	FindingIDs []string `json:"finding_ids"`
	IssueIDs   []string `json:"issue_ids"`
	EvidenceID string   `json:"evidence_id,omitempty"`
}

type ComplianceScorecard struct {
	GroupID    string                    `json:"group_id"`
	GroupName  string                    `json:"group_name"`
	Framework  string                    `json:"framework"`
	Score      *float64                  `json:"score"`
	Assessed   int                       `json:"assessed"`
	Applicable int                       `json:"applicable"`
	Unknown    int                       `json:"unknown"`
	Excluded   int                       `json:"excluded"`
	Controls   []ComplianceControlResult `json:"controls"`
}

func findingMatchesControl(f Finding, mapping string) bool {
	if f.PortState != "" && f.PortState != "open" && f.PortState != "open|filtered" {
		return false
	}
	if mapping == "vulnerability" {
		return f.CVEID != nil && *f.CVEID != ""
	}
	service := ""
	if f.ServiceName != nil {
		service = strings.ToLower(*f.ServiceName)
	}
	if mapping == "cleartext" {
		return service == "telnet" || service == "ftp" || service == "http"
	}
	return mapping == "network" && (f.CVEID != nil || service == "telnet" || service == "ftp")
}

func BuildComplianceScorecards(s *ComplianceSnapshot) {
	s.RuleVersion = ComplianceRuleVersion
	s.Scorecards = []ComplianceScorecard{}
	s.Limitations = []string{"Curated technical evidence mapping, not legal applicability, certification or a full regulatory audit.", "Unknown controls remain in the denominator; evidenced exclusions are shown separately. Partial controls receive half credit.", "Assets require explicit operator identity resolution. Unlinked observations have no remediation MTTR.", "Missing scan results never close an issue. Manual verification is an operator attestation, not automated exploitation or retest.", "Company selection filters tenant-authorized data; tenant roles retain tenant-wide read access.", "DORA catalogue covers ICT risk/testing articles, not every financial or third-party obligation. ISO catalogue covers A.8.8/A.8.20/A.8.24 only."}
	var totalHours float64
	s.CompletedCycles = 0
	s.Limitations = append(s.Limitations, "COMPLETED means the configured scan profile finished, not complete host, port or control coverage. Manual assessments remain operator attestations; source parameter snapshots are shown when available.")
	s.MeanMTTRHours = nil
	for _, issue := range s.Issues {
		for _, hours := range issue.CompletedCycleHours {
			totalHours += hours
			s.CompletedCycles++
		}
	}
	if s.CompletedCycles > 0 {
		mean := totalHours / float64(s.CompletedCycles)
		s.MeanMTTRHours = &mean
	}
	for _, group := range s.Groups {
		for _, framework := range []string{"NIS2", "DORA", "ISO27001", "ALL"} {
			card := ComplianceScorecard{GroupID: group.ID.String(), GroupName: group.Name, Framework: framework, Controls: []ComplianceControlResult{}}
			for _, control := range ComplianceControls() {
				if framework != "ALL" && control.Framework != framework {
					continue
				}
				result := ComplianceControlResult{ComplianceControl: control, Status: "unknown", FindingIDs: []string{}, IssueIDs: []string{}}
				for _, evidence := range s.Evidence {
					if evidence.GroupID == group.ID && evidence.ControlID == control.ID {
						result.Status = evidence.Status
						result.EvidenceID = evidence.ID.String()
					}
				}
				for _, f := range s.Findings {
					if f.GroupID != group.ID || !findingMatchesControl(f.Finding, control.Mapping) {
						continue
					}
					active := true
					if f.LinkedIssueID != nil {
						for _, issue := range s.Issues {
							if issue.ID == *f.LinkedIssueID && issue.State == IssueCorrectedVerified {
								active = false
							}
						}
					}
					if active {
						result.FindingIDs = append(result.FindingIDs, f.ID.String())
					}
				}
				for _, issue := range s.Issues {
					service := strings.TrimPrefix(issue.Key, "service:")
					finding := Finding{ServiceName: &service}
					if strings.HasPrefix(issue.Key, "CVE-") {
						key := issue.Key
						finding.CVEID = &key
					}
					if issue.GroupID == group.ID && issue.State != IssueCorrectedVerified && findingMatchesControl(finding, control.Mapping) {
						result.IssueIDs = append(result.IssueIDs, issue.ID.String())
					}
				}
				if len(result.FindingIDs)+len(result.IssueIDs) > 0 {
					result.Status = "not_satisfied"
				} else if control.Mapping != "manual" && (result.Status == "satisfied" || result.Status == "partial") {
					completed := 0
					for _, scan := range s.Scans {
						if scan.GroupID == group.ID && scan.Status == ScanStatusCompleted {
							completed++
						}
					}
					if s.TargetCounts[group.ID.String()] == 0 || completed != s.TargetCounts[group.ID.String()] {
						result.Status = "unknown"
					}
				}
				card.Controls = append(card.Controls, result)
				if result.Status == "not_applicable" {
					card.Excluded++
					continue
				}
				card.Applicable++
				if result.Status == "unknown" {
					card.Unknown++
				} else {
					card.Assessed++
				}
			}
			if card.Applicable > 0 {
				var credit float64
				for _, c := range card.Controls {
					if c.Status == "satisfied" {
						credit++
					}
					if c.Status == "partial" {
						credit += .5
					}
				}
				score := credit * 100 / float64(card.Applicable)
				card.Score = &score
			}
			s.Scorecards = append(s.Scorecards, card)
		}
	}
	// Company cards aggregate their direct scope and every selected department.
	// A known gap dominates; missing evidence cannot be averaged into compliance.
	for rootIndex := range s.Scorecards {
		root := &s.Scorecards[rootIndex]
		if root.GroupID != s.Scope.CompanyID.String() {
			continue
		}
		merged := append([]ComplianceControlResult(nil), root.Controls...)
		for index := range merged {
			// An empty company container is not an additional unscanned asset.
			if merged[index].Mapping != "manual" && s.TargetCounts[root.GroupID] == 0 && len(s.Groups) > 1 && len(merged[index].IssueIDs) == 0 && len(merged[index].FindingIDs) == 0 {
				merged[index].Status = "not_applicable"
			}
			merged[index].FindingIDs = append([]string{}, merged[index].FindingIDs...)
			merged[index].IssueIDs = append([]string{}, merged[index].IssueIDs...)
			for _, child := range s.Scorecards {
				if child.Framework != root.Framework || child.GroupID == root.GroupID {
					continue
				}
				other := child.Controls[index]
				merged[index].Status = mergeControlStatus(merged[index].Status, other.Status)
				merged[index].FindingIDs = append(merged[index].FindingIDs, other.FindingIDs...)
				merged[index].IssueIDs = append(merged[index].IssueIDs, other.IssueIDs...)
			}
		}
		root.Controls = merged
		root.Applicable, root.Assessed, root.Unknown, root.Excluded = 0, 0, 0, 0
		var credit float64
		for _, control := range merged {
			if control.Status == "not_applicable" {
				root.Excluded++
				continue
			}
			root.Applicable++
			if control.Status == "unknown" {
				root.Unknown++
			} else {
				root.Assessed++
			}
			if control.Status == "satisfied" {
				credit++
			}
			if control.Status == "partial" {
				credit += .5
			}
		}
		root.Score = nil
		if root.Applicable > 0 {
			score := 100 * credit / float64(root.Applicable)
			root.Score = &score
		}
	}
}

func mergeControlStatus(a, b string) string {
	if a == "not_applicable" {
		return b
	}
	if b == "not_applicable" {
		return a
	}
	for _, status := range []string{"not_satisfied", "unknown", "partial", "satisfied"} {
		if a == status || b == status {
			return status
		}
	}
	return "unknown"
}
