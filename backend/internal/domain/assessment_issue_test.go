package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func issueFixture(t *testing.T) (*AssessmentIssue, IssueObservation, uuid.UUID) {
	t.Helper()
	identity, err := NewIssueIdentity(uuid.New(), uuid.New(), " TCP ", 443, "CVE-2026-1234")
	if err != nil {
		t.Fatal(err)
	}
	o := IssueObservation{FindingID: uuid.New(), ScanID: uuid.New(), Identity: identity,
		At: time.Date(2026, 10, 19, 10, 0, 0, 0, time.UTC), Provenance: IssueProvenance{"scanner", "1", "finding/reference"}}
	actor := uuid.New()
	i, err := NewAssessmentIssue(uuid.New(), o, actor, &IssueSLAPolicy{"internal-critical", "1", 24 * time.Hour})
	if err != nil {
		t.Fatal(err)
	}
	return i, o, actor
}

func TestIssueIdentityStableAssetAndIsolation(t *testing.T) {
	i, o, _ := issueFixture(t)
	if i.Identity().Protocol != "tcp" {
		t.Fatal("protocol not canonical")
	}
	for _, change := range []func(*IssueIdentity){
		func(v *IssueIdentity) { v.TenantID = uuid.New() },
		func(v *IssueIdentity) { v.AssetID = uuid.New() },
		func(v *IssueIdentity) { v.Port = 8443 },
		func(v *IssueIdentity) { v.Protocol = "udp" },
		func(v *IssueIdentity) { v.Key = "another-control" },
	} {
		other := o
		other.FindingID = uuid.New()
		change(&other.Identity)
		if err := i.Observe(other); !errors.Is(err, ErrValidation) {
			t.Fatalf("cross-identity observation: %v", err)
		}
	}
	if _, err := NewIssueIdentity(o.Identity.TenantID, uuid.Nil, "tcp", 443, "key"); !errors.Is(err, ErrValidation) {
		t.Fatal("ambiguous asset accepted")
	}
	for _, protocol := range []string{"", "icmp", "tcp|udp"} {
		if _, err := NewIssueIdentity(o.Identity.TenantID, o.Identity.AssetID, protocol, 443, "key"); err == nil {
			t.Fatal("invalid protocol accepted")
		}
	}
	for _, port := range []int{0, -1, 65536} {
		if _, err := NewIssueIdentity(o.Identity.TenantID, o.Identity.AssetID, "tcp", port, "key"); err == nil {
			t.Fatal("invalid port accepted")
		}
	}
	if i.ObservationCount() != 1 || len(i.Transitions()) != 0 {
		t.Fatal("rejections mutated aggregate")
	}
}

func TestIssueReplayConflictAndChronology(t *testing.T) {
	i, original, actor := issueFixture(t)
	if err := i.Declare(IssueCorrectionDeclared, original.At.Add(time.Hour), actor, "patched"); err != nil {
		t.Fatal(err)
	}
	if err := i.Observe(original); err != nil {
		t.Fatal("replay must be idempotent", err)
	}
	conflict := original
	conflict.Provenance.Version = "2"
	if err := i.Observe(conflict); !errors.Is(err, ErrConflict) {
		t.Fatal("conflicting evidence accepted", err)
	}
	old := original
	old.FindingID = uuid.New()
	if err := i.Observe(old); !errors.Is(err, ErrValidation) {
		t.Fatal("out of order observation accepted", err)
	}
	if i.State() != IssueCorrectionDeclared || i.ObservationCount() != 1 {
		t.Fatal("invalid observation changed state")
	}
	newObservation := original
	newObservation.FindingID = uuid.New()
	newObservation.ScanID = uuid.New()
	newObservation.At = original.At.Add(2 * time.Hour)
	if err := i.Observe(newObservation); err != nil {
		t.Fatal(err)
	}
	if i.State() != IssueOpen || !i.FirstSeen().Equal(original.At) || !i.LastSeen().Equal(newObservation.At) {
		t.Fatal("positive evidence did not reopen issue")
	}
	if history := i.Transitions(); history[len(history)-1].ObservationID != newObservation.FindingID {
		t.Fatal("reopening lost its source observation")
	}
}

func TestIssueVerificationAndRecurrence(t *testing.T) {
	i, first, actor := issueFixture(t)
	v := IssueVerification{first.Identity, first.At.Add(4 * time.Hour), IssueProvenance{"retest", "1", "job/evidence"}, true, true}
	if err := i.Verify(v, actor, "retested"); err == nil {
		t.Fatal("undeclared correction verified")
	}
	if err := i.Declare(IssueMitigationApplied, first.At.Add(time.Hour), actor, "mitigated"); err != nil {
		t.Fatal(err)
	}
	if _, complete := i.MTTR(); complete {
		t.Fatal("mitigation counted as correction")
	}
	if err := i.Declare(IssueCorrectionDeclared, first.At.Add(2*time.Hour), actor, "patched"); err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*IssueVerification){
		func(v *IssueVerification) { v.CompleteCoverage = false },
		func(v *IssueVerification) { v.NotPresent = false },
		func(v *IssueVerification) { v.Identity.AssetID = uuid.New() },
		func(v *IssueVerification) { v.Identity.TenantID = uuid.New() },
		func(v *IssueVerification) { v.Provenance.Version = "" },
		func(v *IssueVerification) { v.At = first.At },
	} {
		invalid := v
		change(&invalid)
		if err := i.Verify(invalid, actor, "retested"); !errors.Is(err, ErrValidation) {
			t.Fatalf("unsafe verification accepted: %v", err)
		}
	}
	if err := i.Verify(v, uuid.Nil, "retested"); err == nil {
		t.Fatal("missing actor accepted")
	}
	if err := i.Verify(v, actor, ""); err == nil {
		t.Fatal("missing reason accepted")
	}
	if err := i.Verify(v, actor, "retested"); err != nil {
		t.Fatal(err)
	}
	if elapsed, complete := i.MTTR(); !complete || elapsed != 4*time.Hour {
		t.Fatal("incorrect MTTR", elapsed, complete)
	}
	if i.Breached(first.At.Add(48 * time.Hour)) {
		t.Fatal("closed issue marked as active breach")
	}
	if err := i.Observe(first); err != nil || i.State() != IssueCorrectedVerified {
		t.Fatal("replay reopened issue")
	}
	recurrence := first
	recurrence.FindingID = uuid.New()
	recurrence.ScanID = uuid.New()
	recurrence.At = first.At.Add(72 * time.Hour)
	if err := i.Observe(recurrence); err != nil {
		t.Fatal(err)
	}
	if _, complete := i.MTTR(); complete {
		t.Fatal("old MTTR attributed to reopened issue")
	}
	deadline, configured := i.Deadline()
	if !configured || !deadline.Equal(recurrence.At.Add(24*time.Hour)) {
		t.Fatal("recurrence deadline not reset")
	}
	if !i.FirstSeen().Equal(first.At) || len(i.Transitions()) != 4 {
		t.Fatal("history lost")
	}
	if err := i.Declare(IssueCorrectionDeclared, recurrence.At.Add(time.Hour), actor, "fixed again"); err != nil {
		t.Fatal(err)
	}
	v.At = recurrence.At.Add(2 * time.Hour)
	if err := i.Verify(v, actor, "verified again"); err != nil {
		t.Fatal(err)
	}
	if elapsed, complete := i.MTTR(); !complete || elapsed != 2*time.Hour {
		t.Fatal("MTTR did not use new cycle")
	}
}

func TestIssuePolicyAndHistoryAreDefensiveCopies(t *testing.T) {
	_, first, actor := issueFixture(t)
	policy := &IssueSLAPolicy{"custom", "1", time.Hour}
	i, err := NewAssessmentIssue(uuid.New(), first, uuid.Nil, policy)
	if err != nil {
		t.Fatal(err)
	}
	policy.Duration = 0
	deadline := first.At.Add(time.Hour)
	if i.Breached(deadline) || !i.Breached(deadline.Add(time.Nanosecond)) {
		t.Fatal("deadline boundary or policy copy incorrect")
	}
	if err := i.Declare(IssueCorrectionDeclared, deadline, actor, "fixed"); err != nil {
		t.Fatal(err)
	}
	v := IssueVerification{first.Identity, deadline, IssueProvenance{"test", "1", "evidence"}, true, true}
	if err := i.Verify(v, actor, "verified"); err != nil {
		t.Fatal(err)
	}
	history := i.Transitions()
	history[1].Evidence.Provenance.Reference = "tampered"
	history[0].Reason = "tampered"
	if i.Transitions()[1].Evidence.Provenance.Reference != "evidence" || i.Transitions()[0].Reason != "fixed" {
		t.Fatal("history mutable through accessor")
	}
	withoutPolicy, err := NewAssessmentIssue(uuid.New(), first, uuid.Nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, configured := withoutPolicy.Deadline(); configured || withoutPolicy.Breached(deadline.Add(time.Hour)) {
		t.Fatal("implicit policy created")
	}
}

func TestIssueInvalidInputs(t *testing.T) {
	_, first, actor := issueFixture(t)
	for _, change := range []func(*IssueObservation){
		func(o *IssueObservation) { o.FindingID = uuid.Nil },
		func(o *IssueObservation) { o.ScanID = uuid.Nil },
		func(o *IssueObservation) { o.At = time.Time{} },
		func(o *IssueObservation) { o.Identity.Protocol = "TCP" },
		func(o *IssueObservation) { o.Provenance.Source = " " },
		func(o *IssueObservation) { o.Provenance.Reference = "" },
	} {
		o := first
		change(&o)
		if _, err := NewAssessmentIssue(uuid.New(), o, actor, nil); !errors.Is(err, ErrValidation) {
			t.Fatal("invalid observation accepted", err)
		}
		i, _, _ := issueFixture(t)
		o.Identity = i.Identity()
		if o.FindingID == first.FindingID {
			o.FindingID = uuid.New()
		}
		// Identity canonicalization is tested by the constructor above; ingestion
		// must also reject missing source metadata and invalid observation IDs.
		if o.ScanID == uuid.Nil || o.FindingID == uuid.Nil || o.At.IsZero() || !o.Provenance.valid() {
			if err := i.Observe(o); !errors.Is(err, ErrValidation) {
				t.Fatal("invalid ingestion accepted", err)
			}
		}
	}
	for _, p := range []*IssueSLAPolicy{{"", "1", time.Hour}, {"id", "", time.Hour}, {"id", "1", 0}, {"id", "1", -1}} {
		if _, err := NewAssessmentIssue(uuid.New(), first, actor, p); !errors.Is(err, ErrValidation) {
			t.Fatal("invalid policy accepted", err)
		}
	}
	i, _, _ := issueFixture(t)
	for _, next := range []IssueState{IssueOpen, IssueCorrectedVerified, "unknown"} {
		if err := i.Declare(next, first.At, actor, "reason"); err == nil {
			t.Fatal("invalid transition accepted")
		}
	}
	if err := i.Declare(IssueCorrectionDeclared, first.At, uuid.Nil, "reason"); err == nil {
		t.Fatal("missing actor accepted")
	}
	if err := i.Declare(IssueCorrectionDeclared, first.At, actor, " "); err == nil {
		t.Fatal("missing reason accepted")
	}
}
