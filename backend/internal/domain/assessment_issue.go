package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// IssueIdentity requires a resolved stable asset, never an address-only identity.
type IssueIdentity struct {
	TenantID uuid.UUID `json:"tenant_id"`
	AssetID  uuid.UUID `json:"asset_id"`
	Protocol string    `json:"protocol"`
	Port     int       `json:"port"`
	Key      string    `json:"key"`
}

func NewIssueIdentity(tenant, asset uuid.UUID, protocol string, port int, key string) (IssueIdentity, error) {
	i := IssueIdentity{tenant, asset, strings.ToLower(strings.TrimSpace(protocol)), port, strings.TrimSpace(key)}
	if tenant == uuid.Nil || asset == uuid.Nil || (i.Protocol != "tcp" && i.Protocol != "udp") || port < 1 || port > 65535 || i.Key == "" {
		return IssueIdentity{}, fmt.Errorf("%w: unresolved issue identity or invalid endpoint/key", ErrValidation)
	}
	return i, nil
}

func (i IssueIdentity) valid() bool {
	canonical, err := NewIssueIdentity(i.TenantID, i.AssetID, i.Protocol, i.Port, i.Key)
	return err == nil && canonical == i
}

type IssueProvenance struct {
	Source    string `json:"source"`
	Version   string `json:"version"`
	Reference string `json:"reference"`
}

func (p IssueProvenance) valid() bool {
	return strings.TrimSpace(p.Source) != "" && strings.TrimSpace(p.Version) != "" && strings.TrimSpace(p.Reference) != ""
}

type IssueObservation struct {
	FindingID       uuid.UUID         `json:"finding_id"`
	ScanID          uuid.UUID         `json:"scan_id"`
	Identity        IssueIdentity     `json:"identity"`
	At              time.Time         `json:"at"`
	Provenance      IssueProvenance   `json:"provenance"`
	SourceHost      string            `json:"source_host"`
	SourceTargetID  uuid.UUID         `json:"source_target_id"`
	SourceGroupID   uuid.UUID         `json:"source_group_id"`
	Confidence      FindingConfidence `json:"confidence"`
	DetectionMethod DetectionMethod   `json:"detection_method"`
}

type IssueState string

const (
	IssueOpen               IssueState = "open"
	IssueMitigationApplied  IssueState = "mitigation_applied"
	IssueCorrectionDeclared IssueState = "correction_declared"
	IssueCorrectedVerified  IssueState = "corrected_verified"
)

type IssueSLAPolicy struct {
	ID       string
	Version  string
	Duration time.Duration
}

type IssueTransition struct {
	From          IssueState
	To            IssueState
	At            time.Time
	Actor         uuid.UUID
	Reason        string
	Evidence      *IssueVerification
	ObservationID uuid.UUID
}

// IssueVerification is issue-scoped evidence, not a general scan success flag.
type IssueVerification struct {
	Identity         IssueIdentity   `json:"identity"`
	At               time.Time       `json:"at"`
	Provenance       IssueProvenance `json:"provenance"`
	CompleteCoverage bool            `json:"complete_coverage"`
	NotPresent       bool            `json:"not_present"`
}

// AssessmentIssue is an in-memory aggregate. Persistence must provide atomicity
// and concurrency control; no application path ingests findings into it yet.
type AssessmentIssue struct {
	id                                         uuid.UUID
	identity                                   IssueIdentity
	state                                      IssueState
	owner                                      uuid.UUID
	policy                                     *IssueSLAPolicy
	firstSeen, lastSeen, cycleStart, lastEvent time.Time
	resolvedAt                                 *time.Time
	observations                               map[uuid.UUID]IssueObservation
	transitions                                []IssueTransition
}

func NewAssessmentIssue(id uuid.UUID, first IssueObservation, owner uuid.UUID, policy *IssueSLAPolicy) (*AssessmentIssue, error) {
	if id == uuid.Nil || !first.Identity.valid() || first.FindingID == uuid.Nil || first.ScanID == uuid.Nil || first.At.IsZero() || !first.Provenance.valid() {
		return nil, fmt.Errorf("%w: invalid issue observation", ErrValidation)
	}
	var policyCopy *IssueSLAPolicy
	if policy != nil {
		if strings.TrimSpace(policy.ID) == "" || strings.TrimSpace(policy.Version) == "" || policy.Duration <= 0 {
			return nil, fmt.Errorf("%w: invalid SLA policy", ErrValidation)
		}
		p := *policy
		policyCopy = &p
	}
	return &AssessmentIssue{id: id, identity: first.Identity, state: IssueOpen, owner: owner, policy: policyCopy,
		firstSeen: first.At, lastSeen: first.At, cycleStart: first.At, lastEvent: first.At,
		observations: map[uuid.UUID]IssueObservation{first.FindingID: first}}, nil
}

func (i *AssessmentIssue) ID() uuid.UUID           { return i.id }
func (i *AssessmentIssue) Identity() IssueIdentity { return i.identity }
func (i *AssessmentIssue) State() IssueState       { return i.state }
func (i *AssessmentIssue) Owner() uuid.UUID        { return i.owner }
func (i *AssessmentIssue) FirstSeen() time.Time    { return i.firstSeen }
func (i *AssessmentIssue) LastSeen() time.Time     { return i.lastSeen }
func (i *AssessmentIssue) ObservationCount() int   { return len(i.observations) }

// Observe accepts positive evidence only; missing results never close an issue.
func (i *AssessmentIssue) Observe(o IssueObservation) error {
	if previous, exists := i.observations[o.FindingID]; exists {
		if previous.Identity == o.Identity && previous.ScanID == o.ScanID && previous.At.Equal(o.At) && previous.Provenance == o.Provenance && previous.SourceHost == o.SourceHost && previous.SourceTargetID == o.SourceTargetID && previous.SourceGroupID == o.SourceGroupID && previous.Confidence == o.Confidence && previous.DetectionMethod == o.DetectionMethod {
			return nil
		}
		return fmt.Errorf("%w: observation ID reused with different evidence", ErrConflict)
	}
	if o.Identity != i.identity || o.FindingID == uuid.Nil || o.ScanID == uuid.Nil || !o.Provenance.valid() || o.At.IsZero() || o.At.Before(i.lastEvent) {
		return fmt.Errorf("%w: mismatched or out-of-order observation", ErrValidation)
	}
	if i.state != IssueOpen {
		if i.state == IssueCorrectedVerified {
			i.cycleStart = o.At
			i.resolvedAt = nil
		}
		i.transitions = append(i.transitions, IssueTransition{From: i.state, To: IssueOpen, At: o.At, Reason: "positive_observation", ObservationID: o.FindingID})
		i.state = IssueOpen
	}
	i.observations[o.FindingID] = o
	i.lastSeen, i.lastEvent = o.At, o.At
	return nil
}

// Declare records operator intent, which is distinct from verification.
func (i *AssessmentIssue) Declare(next IssueState, at time.Time, actor uuid.UUID, reason string) error {
	allowed := (next == IssueMitigationApplied && i.state == IssueOpen) ||
		(next == IssueCorrectionDeclared && (i.state == IssueOpen || i.state == IssueMitigationApplied))
	if !allowed || actor == uuid.Nil || strings.TrimSpace(reason) == "" || at.IsZero() || at.Before(i.lastEvent) {
		return fmt.Errorf("%w: invalid operator transition", ErrValidation)
	}
	i.transitions = append(i.transitions, IssueTransition{From: i.state, To: next, At: at, Actor: actor, Reason: reason})
	i.state, i.lastEvent = next, at
	return nil
}

func (i *AssessmentIssue) Verify(v IssueVerification, actor uuid.UUID, reason string) error {
	if i.state != IssueCorrectionDeclared || v.Identity != i.identity || !v.CompleteCoverage || !v.NotPresent || !v.Provenance.valid() || actor == uuid.Nil || strings.TrimSpace(reason) == "" || v.At.IsZero() || v.At.Before(i.lastEvent) {
		return fmt.Errorf("%w: correction requires complete issue-scoped verification", ErrValidation)
	}
	copy := v
	i.transitions = append(i.transitions, IssueTransition{From: i.state, To: IssueCorrectedVerified, At: v.At, Actor: actor, Reason: reason, Evidence: &copy})
	i.state, i.lastEvent = IssueCorrectedVerified, v.At
	at := v.At
	i.resolvedAt = &at
	return nil
}

func (i *AssessmentIssue) Transitions() []IssueTransition {
	result := append([]IssueTransition(nil), i.transitions...)
	for n := range result {
		if result[n].Evidence != nil {
			v := *result[n].Evidence
			result[n].Evidence = &v
		}
	}
	return result
}

func (i *AssessmentIssue) MTTR() (time.Duration, bool) {
	if i.resolvedAt == nil {
		return 0, false
	}
	return i.resolvedAt.Sub(i.cycleStart), true
}

func (i *AssessmentIssue) Deadline() (time.Time, bool) {
	if i.policy == nil {
		return time.Time{}, false
	}
	return i.cycleStart.Add(i.policy.Duration), true
}

func (i *AssessmentIssue) Breached(at time.Time) bool {
	deadline, configured := i.Deadline()
	return configured && i.resolvedAt == nil && at.After(deadline)
}

// Configure changes only explicit operational policy; the event store retains
// the previous configuration and actor alongside the reason.
func (i *AssessmentIssue) Configure(owner uuid.UUID, policy *IssueSLAPolicy, at time.Time, actor uuid.UUID, reason string) error {
	if actor == uuid.Nil || strings.TrimSpace(reason) == "" || at.IsZero() || at.Before(i.lastEvent) {
		return fmt.Errorf("%w: invalid configuration event", ErrValidation)
	}
	var next *IssueSLAPolicy
	if policy != nil {
		if strings.TrimSpace(policy.ID) == "" || strings.TrimSpace(policy.Version) == "" || policy.Duration <= 0 {
			return fmt.Errorf("%w: invalid SLA policy", ErrValidation)
		}
		p := *policy
		next = &p
	}
	i.owner, i.policy, i.lastEvent = owner, next, at
	i.transitions = append(i.transitions, IssueTransition{From: i.state, To: i.state, At: at, Actor: actor, Reason: reason})
	return nil
}
