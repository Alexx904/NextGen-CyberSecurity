package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCalculateCumulativeRisk(t *testing.T) {
	// Caso 1: Nessun nodo
	if risk := CalculateCumulativeRisk(nil, 0); risk != 0.0 {
		t.Errorf("expected 0.0 for empty nodes, got %f", risk)
	}

	// Caso 2: Solo nodi topologici (senza vulnerabilità)
	topologicalNodes := []AttackNode{
		{ID: "n1", Type: NodeTypeEntrypoint},
		{ID: "n2", Type: NodeTypeHost},
	}
	if risk := CalculateCumulativeRisk(topologicalNodes, 2); risk <= 0.0 || risk > 30.0 {
		t.Errorf("expected low topological risk score, got %f", risk)
	}

	// Caso 3: Vulnerabilità media
	vulnNodes := []AttackNode{
		{ID: "n1", Type: NodeTypeEntrypoint},
		{ID: "n2", Type: NodeTypeHost},
		{ID: "n3", Type: NodeTypeVulnerability, CVSS: 7.5},
		{ID: "n4", Type: NodeTypeHost},
	}
	risk := CalculateCumulativeRisk(vulnNodes, 3)
	if risk < 50.0 || risk > 90.0 {
		t.Errorf("expected medium/high risk score for CVSS 7.5, got %f", risk)
	}

	// Caso 4: Vulnerabilità critica (CVSS >= 9.0) deve garantire almeno 80.0
	criticalVuln := []AttackNode{
		{ID: "n1", Type: NodeTypeVulnerability, CVSS: 9.8},
	}
	if critRisk := CalculateCumulativeRisk(criticalVuln, 1); critRisk < 80.0 {
		t.Errorf("expected at least 80.0 for critical vuln, got %f", critRisk)
	}

	// Caso 5: Cumulo estremo clamped a 100.0
	extremeVulns := []AttackNode{
		{ID: "v1", Type: NodeTypeVulnerability, CVSS: 10.0},
		{ID: "v2", Type: NodeTypeVulnerability, CVSS: 10.0},
		{ID: "v3", Type: NodeTypeVulnerability, CVSS: 10.0},
	}
	if maxRisk := CalculateCumulativeRisk(extremeVulns, 1); maxRisk != 100.0 {
		t.Errorf("expected max risk 100.0, got %f", maxRisk)
	}
}

func TestDeterminePathSeverity(t *testing.T) {
	tests := []struct {
		risk     float64
		expected PathSeverity
	}{
		{90.0, PathSeverityCritical},
		{75.0, PathSeverityCritical},
		{74.9, PathSeverityHigh},
		{55.0, PathSeverityHigh},
		{54.9, PathSeverityMedium},
		{35.0, PathSeverityMedium},
		{34.9, PathSeverityLow},
		{10.0, PathSeverityLow},
	}

	for _, tt := range tests {
		got := DeterminePathSeverity(tt.risk)
		if got != tt.expected {
			t.Errorf("risk %f: expected %s, got %s", tt.risk, tt.expected, got)
		}
	}
}

func TestCalculateChokepoints(t *testing.T) {
	nodes := []AttackNode{
		{ID: "entry", Type: NodeTypeEntrypoint, Label: "Internet"},
		{ID: "web", Type: NodeTypeHost, Label: "WebServer-DMZ"},
		{ID: "vuln-rce", Type: NodeTypeVulnerability, Label: "Apache-RCE", CVSS: 9.0},
		{ID: "db", Type: NodeTypeHost, Label: "Database-Internal"},
		{ID: "ssh", Type: NodeTypeService, Label: "SSH-Port22", Port: 22},
	}

	// Due percorsi entrambi passanti per "web"
	paths := []AttackPath{
		{
			ID:             "p1",
			NodeIDs:        []string{"entry", "web", "vuln-rce", "db"},
			CumulativeRisk: 85.0,
		},
		{
			ID:             "p2",
			NodeIDs:        []string{"entry", "web", "ssh", "db"},
			CumulativeRisk: 70.0,
		},
	}

	chokepoints := CalculateChokepoints(paths, nodes)
	if len(chokepoints) == 0 {
		t.Fatal("expected chokepoints, got none")
	}

	// Il nodo "web" compare in 2 cammini (somma rischio 155.0) e deve essere al primo posto
	if chokepoints[0].NodeID != "web" {
		t.Errorf("expected top chokepoint to be 'web', got %s", chokepoints[0].NodeID)
	}
	if chokepoints[0].MitigatedPaths != 2 {
		t.Errorf("expected 2 mitigated paths for 'web', got %d", chokepoints[0].MitigatedPaths)
	}
	if chokepoints[0].MitigatedRisk != 155.0 {
		t.Errorf("expected 155.0 mitigated risk, got %f", chokepoints[0].MitigatedRisk)
	}
	if chokepoints[0].Remediation == "" {
		t.Error("expected non-empty remediation text")
	}
}

func TestAttackGraph_Validate(t *testing.T) {
	tenantID := uuid.New()

	// 1. Grafo valido
	validGraph := AttackGraph{
		TenantID:    tenantID,
		GeneratedAt: time.Now(),
		Nodes: []AttackNode{
			{ID: "n1", Type: NodeTypeEntrypoint, Label: "Entry"},
			{ID: "n2", Type: NodeTypeHost, Label: "Host"},
		},
		Edges: []AttackEdge{
			{ID: "e1", Source: "n1", Target: "n2", Type: EdgeTypeNetworkAccess},
		},
		Paths: []AttackPath{
			{ID: "p1", NodeIDs: []string{"n1", "n2"}, EdgeIDs: []string{"e1"}, HopCount: 1},
		},
	}
	if err := validGraph.Validate(); err != nil {
		t.Fatalf("expected valid graph, got error: %v", err)
	}

	// 2. Tenant ID nullo
	nilTenantGraph := validGraph
	nilTenantGraph.TenantID = uuid.Nil
	if err := nilTenantGraph.Validate(); err == nil {
		t.Error("expected error for nil tenant_id")
	}

	// 3. Nodo con ID duplicato
	dupNodeGraph := validGraph
	dupNodeGraph.Nodes = []AttackNode{
		{ID: "n1", Type: NodeTypeEntrypoint},
		{ID: "n1", Type: NodeTypeHost},
	}
	if err := dupNodeGraph.Validate(); err == nil {
		t.Error("expected error for duplicate node id")
	}

	// 4. Arco verso nodo inesistente
	invalidEdgeGraph := validGraph
	invalidEdgeGraph.Edges = []AttackEdge{
		{ID: "e1", Source: "n1", Target: "non-existent"},
	}
	if err := invalidEdgeGraph.Validate(); err == nil {
		t.Error("expected error for edge referencing unknown node")
	}

	// 5. Percorso verso nodo inesistente
	invalidPathGraph := validGraph
	invalidPathGraph.Paths = []AttackPath{
		{ID: "p1", NodeIDs: []string{"n1", "ghost"}},
	}
	if err := invalidPathGraph.Validate(); err == nil {
		t.Error("expected error for path referencing unknown node")
	}
}
