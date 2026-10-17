package service_test

import (
	"context"
	"encoding/csv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/service"
	"nextgen-cybersec/pkg/report"
)

func TestCompanyReportsIncludeDepartmentsAndKeepDuplicateIPsSeparate(t *testing.T) {
	tenant, root, a, b := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	groups := []*domain.TargetGroup{{ID: root, TenantID: tenant, Name: "Company"}, {ID: a, TenantID: tenant, ParentID: &root, Name: "Finance"}, {ID: b, TenantID: tenant, ParentID: &root, Name: "IT"}}
	t1, t2, t3 := uuid.New(), uuid.New(), uuid.New()
	s1, s2 := uuid.New(), uuid.New()
	targets := map[uuid.UUID][]*domain.Target{
		a: {{ID: t1, TenantID: tenant, GroupID: &a, Value: "10.0.0.1"}, {ID: t3, TenantID: tenant, GroupID: &a, Value: "10.0.0.99"}},
		b: {{ID: t2, TenantID: tenant, GroupID: &b, Value: "10.0.0.1"}},
	}
	scans := map[uuid.UUID][]*domain.ReportItem{
		a: {{ScanID: s1, TargetID: t1, GroupID: &a, Status: domain.ScanStatusCompleted}},
		b: {{ScanID: s2, TargetID: t2, GroupID: &b, Status: domain.ScanStatusCompleted}},
	}
	groupRepo := &reportMockGroupRepo{getByIDFunc: func(_ context.Context, _ uuid.UUID, id uuid.UUID) (*domain.TargetGroup, error) {
		for _, g := range groups {
			if g.ID == id {
				return g, nil
			}
		}
		return nil, domain.ErrNotFound
	}, listFunc: func(context.Context, uuid.UUID) ([]*domain.TargetGroup, error) { return groups, nil }}
	targetRepo := &reportMockTargetRepoWithList{listFunc: func(_ context.Context, _ uuid.UUID, f domain.TargetFilter) ([]*domain.Target, int, error) {
		v := targets[*f.GroupID]
		return v, len(v), nil
	}}
	reportRepo := &reportMockReportRepo{listFunc: func(_ context.Context, _ uuid.UUID, f domain.ReportFilter) ([]*domain.ReportItem, int, error) {
		v := scans[*f.GroupID]
		return v, len(v), nil
	}}
	findingRepo := &reportMockFindingRepo{listFunc: func(_ context.Context, _ uuid.UUID, f domain.FindingFilter) ([]*domain.Finding, int, error) {
		return []*domain.Finding{{ScanID: *f.ScanID, HostIP: "10.0.0.1", Port: 443, Protocol: "tcp", Severity: domain.SeverityInfo}}, 1, nil
	}}
	tenantRepo := &reportMockTenantRepo{getByIDFunc: func(context.Context, uuid.UUID) (*domain.Tenant, error) {
		return &domain.Tenant{ID: tenant, Name: "Test"}, nil
	}}
	var captured report.PDFReportData
	pdf := &reportMockPDFGenerator{generateFunc: func(d report.PDFReportData) ([]byte, error) { captured = d; return []byte("PDF"), nil }}
	svc := service.NewReportService(reportRepo, &reportMockScanRepo{}, &reportMockProfileRepo{}, targetRepo, findingRepo, tenantRepo, pdf, groupRepo)
	if _, _, err := svc.GenerateCompanyPDFReport(context.Background(), tenant, root, domain.ReportTemplateExecutive); err != nil {
		t.Fatal(err)
	}
	if len(captured.Findings) != 2 || len(captured.Departments) != 2 {
		t.Fatalf("company omitted departments: findings=%d sections=%+v", len(captured.Findings), captured.Departments)
	}
	if captured.Departments[0].Name != "Finance" || len(captured.Departments[0].Targets) != 2 {
		t.Fatalf("target without findings disappeared: %+v", captured.Departments)
	}
	if !captured.GroupFindingsByDepartment || len(captured.Departments[0].Findings) != 1 || len(captured.Departments[1].Findings) != 1 || captured.Departments[0].Findings[0].ScanID != s1 || captured.Departments[1].Findings[0].ScanID != s2 {
		t.Fatalf("findings mixed across departments sharing an IP: %+v", captured.Departments)
	}
	if _, _, err := svc.GenerateCompanyPDFReport(context.Background(), tenant, a, domain.ReportTemplateTechnical); err != nil {
		t.Fatal(err)
	}
	if len(captured.Findings) != 1 || len(captured.Departments) != 1 || captured.Departments[0].Name != "Finance" {
		t.Fatalf("department leaked another scope: %+v", captured.Departments)
	}
	if captured.GroupFindingsByDepartment {
		t.Fatal("single-department PDF must keep its existing layout")
	}
	data, _, err := svc.GenerateCompanyCSVReport(context.Background(), tenant, root)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := csv.NewReader(strings.NewReader(string(data))).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3 || rows[0][len(rows[0])-1] != "reparto" || rows[1][len(rows[1])-1] != "Finance" || rows[2][len(rows[2])-1] != "IT" {
		t.Fatalf("wrong CSV departments: %+v", rows)
	}

	// Direct-company findings must not merge into a real department with a fallback name.
	directTarget, directScan, namedTarget, namedScan, namedDepartment, emptyDepartment := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	groups = append(groups, &domain.TargetGroup{ID: namedDepartment, TenantID: tenant, ParentID: &root, Name: "Senza reparto"}, &domain.TargetGroup{ID: emptyDepartment, TenantID: tenant, ParentID: &root, Name: "Unscanned"})
	targets[root] = []*domain.Target{{ID: directTarget, TenantID: tenant, GroupID: &root, Value: "10.0.0.1"}}
	targets[namedDepartment] = []*domain.Target{{ID: namedTarget, TenantID: tenant, GroupID: &namedDepartment, Value: "10.0.0.1"}}
	scans[root] = []*domain.ReportItem{{ScanID: directScan, TargetID: directTarget, GroupID: &root, Status: domain.ScanStatusCompleted}}
	scans[namedDepartment] = []*domain.ReportItem{{ScanID: namedScan, TargetID: namedTarget, GroupID: &namedDepartment, Status: domain.ScanStatusCompleted}}
	if _, _, err := svc.GenerateCompanyPDFReport(context.Background(), tenant, root, domain.ReportTemplateTechnical); err != nil {
		t.Fatal(err)
	}
	if len(captured.Departments) != 5 {
		t.Fatalf("department identities merged or empty department missing: %+v", captured.Departments)
	}
	for _, section := range captured.Departments {
		if section.Name == "Senza reparto" && (len(section.Findings) != 1 || section.Findings[0].ScanID != directScan) {
			t.Fatalf("direct-company finding misattributed: %+v", section)
		}
	}
	groups = groups[:1] // Same company, no departments: keep the original flat layout.
	if _, _, err := svc.GenerateCompanyPDFReport(context.Background(), tenant, root, domain.ReportTemplateTechnical); err != nil {
		t.Fatal(err)
	}
	if captured.GroupFindingsByDepartment || len(captured.Findings) != 1 {
		t.Fatal("company without departments must keep the original flat layout")
	}
}
