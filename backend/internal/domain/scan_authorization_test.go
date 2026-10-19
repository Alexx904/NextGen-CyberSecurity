package domain

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestScanAuthorizationBinding(t *testing.T) {
	tenant, user := uuid.New(), uuid.New()
	target := &Target{ID: uuid.New(), TenantID: tenant, Value: "192.168.1.20"}
	profile := &ScanProfile{ID: uuid.New(), NmapArgs: "-sS -T3", TimeoutSeconds: 60}
	request := ScanAuthorizationRequest{Accepted: true, Statement: ScanAuthorizationStatement, Targets: []AuthorizedTarget{{ID: target.ID, Value: target.Value}}, NmapArgs: profile.NmapArgs, TimeoutSeconds: 60}
	auth := ScanAuthorizationFromContext(WithScanAuthorization(context.Background(), tenant, user, &request))
	if err := auth.Validate(tenant, []*Target{target}, profile, ""); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name   string
		mutate func(*ScanAuthorization)
	}{
		{"missing consent", func(a *ScanAuthorization) { a.Accepted = false }},
		{"old statement", func(a *ScanAuthorization) { a.Statement = "old" }},
		{"other tenant", func(a *ScanAuthorization) { a.TenantID = uuid.New() }},
		{"missing actor", func(a *ScanAuthorization) { a.UserID = uuid.Nil }},
		{"changed target value", func(a *ScanAuthorization) { a.Targets[0].Value = "192.168.1.21" }},
		{"changed target ID", func(a *ScanAuthorization) { a.Targets[0].ID = uuid.New() }},
		{"changed options", func(a *ScanAuthorization) { a.NmapArgs = "-sS -T5" }},
		{"changed timeout", func(a *ScanAuthorization) { a.TimeoutSeconds++ }},
		{"changed recurrence", func(a *ScanAuthorization) { a.CronExpression = "0 2 * * *" }},
		{"extra target", func(a *ScanAuthorization) { a.Targets = append(a.Targets, a.Targets[0]) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			copy := *auth
			copy.Targets = append([]AuthorizedTarget(nil), auth.Targets...)
			tc.mutate(&copy)
			if err := copy.Validate(tenant, []*Target{target}, profile, ""); err != ErrScanAuthorizationRequired {
				t.Fatalf("got %v", err)
			}
		})
	}
	if err := (*ScanAuthorization)(nil).Validate(tenant, []*Target{target}, profile, ""); err != ErrScanAuthorizationRequired {
		t.Fatal(err)
	}
	other := uuid.New()
	profile.TenantID = &other
	if err := auth.Validate(tenant, []*Target{target}, profile, ""); err != ErrScanAuthorizationRequired {
		t.Fatal("cross-tenant profile accepted")
	}
}

func TestScanAuthorizationCopiesRequestAndNarrowsScope(t *testing.T) {
	tenant := uuid.New()
	first := &Target{ID: uuid.New(), TenantID: tenant, Value: "192.168.1.20"}
	second := &Target{ID: uuid.New(), TenantID: tenant, Value: "192.168.1.21"}
	profile := &ScanProfile{NmapArgs: "-sS", TimeoutSeconds: 60}
	request := ScanAuthorizationRequest{Accepted: true, Statement: ScanAuthorizationStatement, Targets: []AuthorizedTarget{{ID: first.ID, Value: first.Value}, {ID: second.ID, Value: second.Value}}, NmapArgs: "-sS", TimeoutSeconds: 60}
	auth := ScanAuthorizationFromContext(WithScanAuthorization(context.Background(), tenant, uuid.New(), &request))
	request.Targets[0].Value = "edited"
	if err := auth.Validate(tenant, []*Target{first, second}, profile, ""); err != nil {
		t.Fatal(err)
	}
	narrowed := auth.ForTarget(first)
	if err := narrowed.Validate(tenant, []*Target{first}, profile, ""); err != nil {
		t.Fatal(err)
	}
	if len(auth.Targets) != 2 || narrowed.UserID != auth.UserID || !narrowed.AcceptedAt.Equal(auth.AcceptedAt) {
		t.Fatal("evidence identity or source mutated")
	}
	auth.Targets[1] = auth.Targets[0]
	if err := auth.Validate(tenant, []*Target{first, second}, profile, ""); err != ErrScanAuthorizationRequired {
		t.Fatal("duplicate scope accepted")
	}
}
