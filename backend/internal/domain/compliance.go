package domain

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const ComplianceRuleVersion = "2026-10-19.1"

type ComplianceScope struct {
	CompanyID    uuid.UUID `json:"company_id"`
	DepartmentID uuid.UUID `json:"department_id"`
}

type ComplianceGroup struct {
	ID       uuid.UUID  `json:"id"`
	Name     string     `json:"name"`
	ParentID *uuid.UUID `json:"parent_id"`
}

type AssessmentAsset struct {
	ID      uuid.UUID `json:"id"`
	GroupID uuid.UUID `json:"group_id"`
	Name    string    `json:"name"`
}

type ComplianceFinding struct {
	Finding
	TargetID      uuid.UUID  `json:"target_id"`
	GroupID       uuid.UUID  `json:"group_id"`
	ScanStatus    ScanStatus `json:"scan_status"`
	LinkedIssueID *uuid.UUID `json:"linked_issue_id,omitempty"`
}

type ComplianceScan struct {
	ProfileID     uuid.UUID          `json:"profile_id"`
	Authorization *ScanAuthorization `json:"authorization"`
	ID            uuid.UUID          `json:"id"`
	TargetID      uuid.UUID          `json:"target_id"`
	GroupID       uuid.UUID          `json:"group_id"`
	Status        ScanStatus         `json:"status"`
	At            time.Time          `json:"at"`
}

type IssueCommand struct {
	Action          string             `json:"action"`
	ExpectedVersion int                `json:"expected_version"`
	Reason          string             `json:"reason"`
	Actor           uuid.UUID          `json:"actor"`
	At              time.Time          `json:"at"`
	Observation     *IssueObservation  `json:"observation,omitempty"`
	Owner           uuid.UUID          `json:"owner"`
	PolicyID        string             `json:"policy_id"`
	PolicyVersion   string             `json:"policy_version"`
	SLAHours        int                `json:"sla_hours"`
	Verification    *IssueVerification `json:"verification,omitempty"`
}

type ComplianceIssue struct {
	ID                  uuid.UUID      `json:"id"`
	AssetID             uuid.UUID      `json:"asset_id"`
	GroupID             uuid.UUID      `json:"group_id"`
	AssetName           string         `json:"asset_name"`
	Key                 string         `json:"key"`
	Port                int            `json:"port"`
	Protocol            string         `json:"protocol"`
	Version             int            `json:"version"`
	State               IssueState     `json:"state"`
	Owner               uuid.UUID      `json:"owner"`
	FirstSeen           time.Time      `json:"first_seen"`
	LastSeen            time.Time      `json:"last_seen"`
	Deadline            *time.Time     `json:"deadline"`
	Breached            bool           `json:"breached"`
	MTTRHours           *float64       `json:"mttr_hours"`
	PolicyID            string         `json:"policy_id"`
	PolicyVersion       string         `json:"policy_version"`
	SLAHours            int            `json:"sla_hours"`
	CompletedCycleHours []float64      `json:"completed_cycle_hours"`
	Events              []IssueCommand `json:"events"`
}

// ReplayIssue applies the same domain rules on every write and read. Database
// version ordering is authoritative; caller-provided state is never trusted.
func ReplayIssue(tenant uuid.UUID, issue *ComplianceIssue, cutoff time.Time) (*AssessmentIssue, error) {
	var aggregate *AssessmentIssue
	issue.CompletedCycleHours = []float64{}
	for index, e := range issue.Events {
		if e.At.IsZero() || e.Actor == uuid.Nil || e.ExpectedVersion != index {
			return nil, fmt.Errorf("%w: invalid issue event envelope", ErrValidation)
		}
		var err error
		switch e.Action {
		case "observe":
			if e.Observation == nil || e.Observation.Identity.TenantID != tenant || e.Observation.Identity.AssetID != issue.AssetID || e.Observation.Identity.Key != issue.Key || e.Observation.Identity.Port != issue.Port || e.Observation.Identity.Protocol != issue.Protocol {
				return nil, fmt.Errorf("%w: observation identity mismatch", ErrValidation)
			}
			if aggregate == nil {
				aggregate, err = NewAssessmentIssue(issue.ID, *e.Observation, uuid.Nil, nil)
			} else {
				err = aggregate.Observe(*e.Observation)
			}
		case "configure":
			if aggregate == nil {
				return nil, ErrValidation
			}
			var policy *IssueSLAPolicy
			if e.SLAHours != 0 {
				if e.SLAHours < 1 || e.SLAHours > 87600 {
					return nil, ErrValidation
				}
				policy = &IssueSLAPolicy{e.PolicyID, e.PolicyVersion, time.Duration(e.SLAHours) * time.Hour}
			}
			err = aggregate.Configure(e.Owner, policy, e.At, e.Actor, e.Reason)
			if err == nil {
				issue.PolicyID, issue.PolicyVersion, issue.SLAHours = e.PolicyID, e.PolicyVersion, e.SLAHours
			}
		case "mitigation_applied", "correction_declared":
			if aggregate == nil {
				return nil, ErrValidation
			}
			err = aggregate.Declare(IssueState(e.Action), e.At, e.Actor, e.Reason)
		case "verify":
			if aggregate == nil || e.Verification == nil || !e.Verification.At.Equal(e.At) {
				return nil, ErrValidation
			}
			err = aggregate.Verify(*e.Verification, e.Actor, e.Reason)
			if err == nil {
				elapsed, _ := aggregate.MTTR()
				issue.CompletedCycleHours = append(issue.CompletedCycleHours, elapsed.Hours())
			}
		default:
			return nil, ErrValidation
		}
		if err != nil {
			return nil, err
		}
	}
	if aggregate == nil {
		return nil, ErrValidation
	}
	issue.Version, issue.State, issue.Owner = len(issue.Events), aggregate.State(), aggregate.Owner()
	issue.FirstSeen, issue.LastSeen, issue.Breached = aggregate.FirstSeen(), aggregate.LastSeen(), aggregate.Breached(cutoff)
	issue.Deadline, issue.MTTRHours = nil, nil
	if deadline, ok := aggregate.Deadline(); ok {
		issue.Deadline = &deadline
	}
	if elapsed, ok := aggregate.MTTR(); ok {
		hours := elapsed.Hours()
		issue.MTTRHours = &hours
	}
	return aggregate, nil
}

type ControlEvidence struct {
	ID            uuid.UUID `json:"id"`
	GroupID       uuid.UUID `json:"group_id"`
	ControlID     string    `json:"control_id"`
	Status        string    `json:"status"`
	Reason        string    `json:"reason"`
	Reference     string    `json:"reference"`
	SourceVersion string    `json:"source_version"`
	Actor         uuid.UUID `json:"actor"`
	At            time.Time `json:"at"`
}

func (e ControlEvidence) Valid() bool {
	validStatus := e.Status == "satisfied" || e.Status == "partial" || e.Status == "not_satisfied" || e.Status == "not_applicable" || e.Status == "unknown"
	known := false
	for _, c := range ComplianceControls() {
		if c.ID == e.ControlID {
			known = true
		}
	}
	return known && validStatus && e.GroupID != uuid.Nil && e.Actor != uuid.Nil && strings.TrimSpace(e.Reason) != "" && strings.TrimSpace(e.Reference) != "" && strings.TrimSpace(e.SourceVersion) != "" && len(e.Reason) <= 4000 && len(e.Reference) <= 2000 && len(e.SourceVersion) <= 100
}

type ComplianceSnapshot struct {
	Owners          []ComplianceOwner     `json:"owners"`
	SnapshotID      string                `json:"snapshot_id"`
	TargetCounts    map[string]int        `json:"target_counts"`
	Cutoff          time.Time             `json:"cutoff"`
	RuleVersion     string                `json:"rule_version"`
	Scope           ComplianceScope       `json:"scope"`
	Groups          []ComplianceGroup     `json:"groups"`
	Assets          []AssessmentAsset     `json:"assets"`
	Findings        []ComplianceFinding   `json:"findings"`
	Scans           []ComplianceScan      `json:"scans"`
	Issues          []ComplianceIssue     `json:"issues"`
	Evidence        []ControlEvidence     `json:"evidence"`
	Scorecards      []ComplianceScorecard `json:"scorecards"`
	MeanMTTRHours   *float64              `json:"mean_mttr_hours"`
	CompletedCycles int                   `json:"completed_cycles"`
	Limitations     []string              `json:"limitations"`
}

type ComplianceOwner struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type ComplianceRepository interface {
	Snapshot(context.Context, uuid.UUID, ComplianceScope) (*ComplianceSnapshot, error)
	CreateAsset(context.Context, uuid.UUID, AssessmentAsset) (AssessmentAsset, error)
	ObserveFinding(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) (uuid.UUID, error)
	AppendCommand(context.Context, uuid.UUID, uuid.UUID, IssueCommand) error
	AppendEvidence(context.Context, uuid.UUID, ControlEvidence) error
}
