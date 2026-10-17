package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"nextgen-cybersec/internal/domain"
)

// In-Memory Mocks per Target, Finding e SSHServer Repository
type mockAttackPathTargetRepo struct {
	targets []*domain.Target
}

func (m *mockAttackPathTargetRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Target, error) {
	return nil, domain.ErrNotFound
}
func (m *mockAttackPathTargetRepo) GetByValue(ctx context.Context, tenantID uuid.UUID, value string) (*domain.Target, error) {
	return nil, domain.ErrNotFound
}
func (m *mockAttackPathTargetRepo) List(ctx context.Context, tenantID uuid.UUID, filter domain.TargetFilter) ([]*domain.Target, int, error) {
	end := filter.Offset + filter.Limit
	if end > len(m.targets) {
		end = len(m.targets)
	}
	if filter.Offset >= len(m.targets) {
		return nil, len(m.targets), nil
	}
	return m.targets[filter.Offset:end], len(m.targets), nil
}
func (m *mockAttackPathTargetRepo) Create(ctx context.Context, target *domain.Target) error {
	return nil
}
func (m *mockAttackPathTargetRepo) Update(ctx context.Context, target *domain.Target) error {
	return nil
}
func (m *mockAttackPathTargetRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	return nil
}
func (m *mockAttackPathTargetRepo) IsBlacklisted(ctx context.Context, tenantID uuid.UUID, value string) (bool, error) {
	return false, nil
}

type mockAttackPathFindingRepo struct {
	findings []*domain.Finding
}

func (m *mockAttackPathFindingRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Finding, error) {
	return nil, domain.ErrNotFound
}
func (m *mockAttackPathFindingRepo) List(ctx context.Context, tenantID uuid.UUID, filter domain.FindingFilter) ([]*domain.Finding, int, error) {
	res := make([]*domain.Finding, 0)
	for _, f := range m.findings {
		if filter.HostIP != "" && f.HostIP != filter.HostIP {
			continue
		}
		res = append(res, f)
	}
	total := len(res)
	end := filter.Offset + filter.Limit
	if end > total {
		end = total
	}
	if filter.Offset >= total {
		return nil, total, nil
	}
	return res[filter.Offset:end], total, nil
}
func (m *mockAttackPathFindingRepo) CreateBatch(ctx context.Context, findings []*domain.Finding) error {
	return nil
}
func (m *mockAttackPathFindingRepo) GetSeverityCountsByScan(ctx context.Context, tenantID, scanID uuid.UUID) (map[domain.SeverityLevel]int, error) {
	return nil, nil
}

type mockAttackPathSSHRepo struct {
	servers []*domain.SSHServer
}

func (m *mockAttackPathSSHRepo) Create(ctx context.Context, server *domain.SSHServer) error {
	return nil
}
func (m *mockAttackPathSSHRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.SSHServer, error) {
	return nil, domain.ErrNotFound
}
func (m *mockAttackPathSSHRepo) List(ctx context.Context, tenantID uuid.UUID) ([]*domain.SSHServer, error) {
	return m.servers, nil
}
func (m *mockAttackPathSSHRepo) Update(ctx context.Context, server *domain.SSHServer) error {
	return nil
}
func (m *mockAttackPathSSHRepo) UpdateStatus(ctx context.Context, tenantID, id uuid.UUID, status domain.SSHServerStatus, lastConnectedAt *time.Time) error {
	return nil
}
func (m *mockAttackPathSSHRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error { return nil }

func TestGenerateAttackGraph_Empty(t *testing.T) {
	tenantID := uuid.New()
	targetRepo := &mockAttackPathTargetRepo{targets: nil}
	findingRepo := &mockAttackPathFindingRepo{findings: nil}
	sshRepo := &mockAttackPathSSHRepo{servers: nil}

	svc := NewAttackPathService(targetRepo, findingRepo, sshRepo)

	graph, err := svc.GenerateAttackGraph(context.Background(), tenantID, 0.0)
	if err != nil {
		t.Fatalf("expected no error for empty graph, got %v", err)
	}

	if graph.TenantID != tenantID {
		t.Errorf("expected tenantID %s, got %s", tenantID, graph.TenantID)
	}
	if len(graph.Nodes) != 1 || graph.Nodes[0].Type != domain.NodeTypeEntrypoint {
		t.Errorf("expected only entrypoint node in empty graph, got %d nodes", len(graph.Nodes))
	}
	if len(graph.Edges) != 0 {
		t.Errorf("expected 0 edges, got %d", len(graph.Edges))
	}
	if len(graph.Paths) != 0 {
		t.Errorf("expected 0 paths, got %d", len(graph.Paths))
	}
}

func TestGenerateAttackGraph_NilTenant(t *testing.T) {
	svc := NewAttackPathService(&mockAttackPathTargetRepo{}, &mockAttackPathFindingRepo{}, &mockAttackPathSSHRepo{})
	_, err := svc.GenerateAttackGraph(context.Background(), uuid.Nil, 0.0)
	if err == nil {
		t.Fatal("expected error for uuid.Nil tenant")
	}
}

func TestGenerateAttackGraph_ExternalToInternalPivot(t *testing.T) {
	tenantID := uuid.New()
	cveID := "CVE-2021-41773"
	cvss := 8.5
	serviceName := "http"
	descExternal := "DMZ Web Server"
	descInternal := "Internal DB Server"

	targets := []*domain.Target{
		{
			ID:          uuid.New(),
			TenantID:    tenantID,
			Type:        domain.TargetTypeIP,
			Value:       "203.0.113.10", // Pubblico / Esterno
			Description: &descExternal,
		},
		{
			ID:          uuid.New(),
			TenantID:    tenantID,
			Type:        domain.TargetTypeIP,
			Value:       "10.0.0.50", // Privato RFC 1918 / Interno
			Description: &descInternal,
		},
	}

	findings := []*domain.Finding{
		{
			ID:          uuid.New(),
			TenantID:    tenantID,
			HostIP:      "203.0.113.10",
			Port:        80,
			Protocol:    "tcp",
			ServiceName: &serviceName,
			CVEID:       &cveID,
			CVSSScore:   &cvss,
			Severity:    domain.SeverityCritical,
		},
		{
			ID:          uuid.New(),
			TenantID:    tenantID,
			HostIP:      "10.0.0.50",
			Port:        5432,
			Protocol:    "tcp",
			ServiceName: &serviceName,
			Severity:    domain.SeverityMedium,
		},
	}

	svc := NewAttackPathService(
		&mockAttackPathTargetRepo{targets: targets},
		&mockAttackPathFindingRepo{findings: findings},
		&mockAttackPathSSHRepo{servers: nil},
	)

	graph, err := svc.GenerateAttackGraph(context.Background(), tenantID, 5.0)
	if err != nil {
		t.Fatalf("expected graph generation to succeed, got %v", err)
	}

	// 1. Verifica consistenza e validazione del grafo
	if err := graph.Validate(); err != nil {
		t.Fatalf("graph validation failed: %v", err)
	}

	// 2. Verifica nodi
	nodeTypes := make(map[domain.AttackNodeType]int)
	for _, n := range graph.Nodes {
		nodeTypes[n.Type]++
	}
	if nodeTypes[domain.NodeTypeEntrypoint] != 1 {
		t.Errorf("expected 1 entrypoint, got %d", nodeTypes[domain.NodeTypeEntrypoint])
	}
	if nodeTypes[domain.NodeTypeHost] != 2 {
		t.Errorf("expected 2 hosts, got %d", nodeTypes[domain.NodeTypeHost])
	}
	if nodeTypes[domain.NodeTypeVulnerability] < 1 {
		t.Errorf("expected at least 1 vulnerability node, got %d", nodeTypes[domain.NodeTypeVulnerability])
	}

	// 3. Verifica esistenza arco di Pivot verso la rete interna
	hasPivotEdge := false
	for _, e := range graph.Edges {
		if e.Type == domain.EdgeTypePivotsTo {
			hasPivotEdge = true
			break
		}
	}
	if hasPivotEdge {
		t.Error("unrelated public/private addresses must not establish a lateral relationship")
	}

	// 4. Verifica percorsi (Paths)
	if len(graph.Paths) == 0 {
		t.Fatal("expected attack paths, got none")
	}

	// Il primo percorso deve essere il più critico
	topPath := graph.Paths[0]
	if topPath.CumulativeRisk < 60.0 {
		t.Errorf("expected high cumulative risk for critical path, got %f", topPath.CumulativeRisk)
	}

	// 5. Verifica Chokepoints
	if len(graph.Chokepoints) == 0 {
		t.Fatal("expected at least one chokepoint, got none")
	}
	topChokepoint := graph.Chokepoints[0]
	if topChokepoint.MitigatedPaths == 0 {
		t.Errorf("expected mitigated paths > 0 for chokepoint, got %d", topChokepoint.MitigatedPaths)
	}

	// 6. Verifica Chokepoint via metodo GetChokepoints
	chokepoints, err := svc.GetChokepoints(context.Background(), tenantID)
	if err != nil {
		t.Fatalf("GetChokepoints failed: %v", err)
	}
	if len(chokepoints) == 0 {
		t.Fatal("expected chokepoints from GetChokepoints, got none")
	}
}

func TestGenerateAttackGraph_MinCVSSFiltering(t *testing.T) {
	tenantID := uuid.New()
	cveLow := "CVE-2020-0001"
	cvssLow := 4.2
	cveHigh := "CVE-2021-9999"
	cvssHigh := 9.8

	findings := []*domain.Finding{
		{
			ID:        uuid.New(),
			TenantID:  tenantID,
			HostIP:    "192.168.1.10",
			Port:      80,
			Protocol:  "tcp",
			CVEID:     &cveLow,
			CVSSScore: &cvssLow,
			Severity:  domain.SeverityMedium,
		},
		{
			ID:        uuid.New(),
			TenantID:  tenantID,
			HostIP:    "192.168.1.10",
			Port:      443,
			Protocol:  "tcp",
			CVEID:     &cveHigh,
			CVSSScore: &cvssHigh,
			Severity:  domain.SeverityCritical,
		},
	}

	svc := NewAttackPathService(
		&mockAttackPathTargetRepo{targets: nil},
		&mockAttackPathFindingRepo{findings: findings},
		&mockAttackPathSSHRepo{servers: nil},
	)

	// Filtraggio con minCVSS = 7.0: deve escludere CVE-2020-0001
	graph, err := svc.GenerateAttackGraph(context.Background(), tenantID, 7.0)
	if err != nil {
		t.Fatalf("GenerateAttackGraph failed: %v", err)
	}

	for _, n := range graph.Nodes {
		if n.Type == domain.NodeTypeVulnerability && n.CVSS < 7.0 {
			t.Errorf("found vulnerability node with CVSS %.1f below threshold 7.0: %s", n.CVSS, n.Label)
		}
	}
}
