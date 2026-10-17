package service

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"nextgen-cybersec/internal/domain"
)

func reviewFinding(host, protocol, cve string, score float64) *domain.Finding {
	return &domain.Finding{HostIP: host, Port: 53, Protocol: protocol, CVEID: &cve, CVSSScore: &score, PortState: "open"}
}

func TestAttackReviewEveryPathFollowsDirectedEdges(t *testing.T) {
	svc := NewAttackPathService(&mockAttackPathTargetRepo{targets: []*domain.Target{{Type: domain.TargetTypeIP, Value: "10.0.0.20"}}}, &mockAttackPathFindingRepo{findings: []*domain.Finding{reviewFinding("10.0.0.10", "tcp", "CVE-TEST", 9.8)}}, &mockAttackPathSSHRepo{})
	g, err := svc.GenerateAttackGraph(context.Background(), uuid.New(), 0)
	if err != nil {
		t.Fatal(err)
	}
	edges := map[string]domain.AttackEdge{}
	pivots := 0
	for _, e := range g.Edges {
		edges[e.ID] = e
		if e.Type == domain.EdgeTypePivotsTo {
			pivots++
			if !e.Inferred || e.InferenceBasis == "" || e.Confidence != 0 {
				t.Fatalf("unjustified lateral edge: %+v", e)
			}
		}
	}
	if pivots != 1 {
		t.Fatalf("expected one explicitly hypothetical /24 relation, got %d", pivots)
	}
	for _, p := range g.Paths {
		if len(p.EdgeIDs) != len(p.NodeIDs)-1 {
			t.Fatalf("path %s has no complete edge sequence: %+v", p.ID, p)
		}
		for i, id := range p.EdgeIDs {
			e, ok := edges[id]
			if !ok || e.Source != p.NodeIDs[i] || e.Target != p.NodeIDs[i+1] {
				t.Fatalf("disconnected path %+v", p)
			}
		}
	}
}

func TestAttackReviewHistoryAndClosedEndpoint(t *testing.T) {
	old := reviewFinding("10.0.0.10", "tcp", "CVE-OLD", 9)
	old.ScanID = uuid.New()
	old.CreatedAt = time.Now().Add(-time.Hour)
	current := reviewFinding("10.0.0.10", "tcp", "CVE-CURRENT", 5)
	current.ScanID = uuid.New()
	current.CreatedAt = time.Now()
	closed := reviewFinding("10.0.0.20", "tcp", "CVE-CLOSED", 10)
	closed.PortState = "closed"
	svc := NewAttackPathService(&mockAttackPathTargetRepo{}, &mockAttackPathFindingRepo{findings: []*domain.Finding{old, current, closed}}, &mockAttackPathSSHRepo{})
	g, err := svc.GenerateAttackGraph(context.Background(), uuid.New(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(g.Paths) != 1 || g.Paths[0].TargetAsset != "10.0.0.10" {
		t.Fatalf("unexpected history/closed paths: %+v", g.Paths)
	}
	for _, n := range g.Nodes {
		if n.Label == "CVE-OLD" || n.Label == "CVE-CLOSED" {
			t.Fatalf("stale node: %+v", n)
		}
	}
}

func TestAttackReviewPaginationAndPathLimit(t *testing.T) {
	targets := []*domain.Target{}
	findings := []*domain.Finding{}
	for i := 0; i < 205; i++ {
		targets = append(targets, &domain.Target{Type: domain.TargetTypeIP, Value: fmt.Sprintf("10.1.%d.1", i)})
	}
	for i := 0; i < 505; i++ {
		f := reviewFinding("10.2.0.1", "tcp", fmt.Sprintf("CVE-%04d", i), 8)
		f.Port = 1000 + i
		findings = append(findings, f)
	}
	svc := NewAttackPathService(&mockAttackPathTargetRepo{targets}, &mockAttackPathFindingRepo{findings}, &mockAttackPathSSHRepo{})
	g, err := svc.GenerateAttackGraph(context.Background(), uuid.New(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if g.Analysis.TargetCount != 205 || g.Analysis.FindingCount != 505 || g.Analysis.InventoryTruncated {
		t.Fatalf("pagination lost inventory: %+v", g.Analysis)
	}
	if len(g.Paths) != 50 || !g.Analysis.PathsTruncated {
		t.Fatalf("path limit not disclosed: %+v", g.Analysis)
	}
}

func TestAttackReviewSSHInventoryDoesNotCreateTrust(t *testing.T) {
	targets := []*domain.Target{{Type: domain.TargetTypeIP, Value: "10.0.0.10"}, {Type: domain.TargetTypeIP, Value: "10.0.0.20"}}
	svc := NewAttackPathService(&mockAttackPathTargetRepo{targets}, &mockAttackPathFindingRepo{}, &mockAttackPathSSHRepo{servers: []*domain.SSHServer{{Host: "10.0.0.10"}}})
	g, err := svc.GenerateAttackGraph(context.Background(), uuid.New(), 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range g.Edges {
		if e.Type == domain.EdgeTypePivotsTo || e.Type == domain.EdgeTypeTrustedRelation {
			t.Fatalf("SSH inventory invented trust: %+v", e)
		}
	}
}

func TestAttackReviewNoInventedInternetOrCrossSubnetPivot(t *testing.T) {
	svc := NewAttackPathService(&mockAttackPathTargetRepo{targets: []*domain.Target{{Type: domain.TargetTypeIP, Value: "10.9.0.20"}}}, &mockAttackPathFindingRepo{findings: []*domain.Finding{reviewFinding("192.168.1.10", "tcp", "CVE-TEST", 9.8)}}, &mockAttackPathSSHRepo{servers: []*domain.SSHServer{{Host: "10.9.0.20"}}})
	g, err := svc.GenerateAttackGraph(context.Background(), uuid.New(), 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range g.Edges {
		if e.Source == "entry-internet" && !e.Inferred {
			t.Fatalf("invented Internet access: %+v", e)
		}
		if e.Type == domain.EdgeTypePivotsTo {
			t.Fatalf("invented cross-subnet pivot: %+v", e)
		}
	}
}

func TestAttackReviewProtocolsDeduplicationAndDeterminism(t *testing.T) {
	findings := []*domain.Finding{reviewFinding("10.0.0.10", "tcp", "CVE-TEST", 8), reviewFinding("10.0.0.10", "udp", "CVE-TEST", 9), reviewFinding("10.0.0.10", "tcp", "CVE-TEST", 8)}
	svc := NewAttackPathService(&mockAttackPathTargetRepo{}, &mockAttackPathFindingRepo{findings: findings}, &mockAttackPathSSHRepo{})
	tenant := uuid.New()
	g, err := svc.GenerateAttackGraph(context.Background(), tenant, 0)
	if err != nil {
		t.Fatal(err)
	}
	services := 0
	for _, n := range g.Nodes {
		if n.Type == domain.NodeTypeService {
			services++
		}
	}
	if services != 2 || len(g.Paths) != 2 {
		t.Fatalf("expected distinct tcp/udp and deduped paths, services=%d paths=%d", services, len(g.Paths))
	}
	again, _ := svc.GenerateAttackGraph(context.Background(), tenant, 0)
	if !reflect.DeepEqual(g.Nodes, again.Nodes) || !reflect.DeepEqual(g.Edges, again.Edges) || !reflect.DeepEqual(g.Paths, again.Paths) || !reflect.DeepEqual(g.Chokepoints, again.Chokepoints) {
		t.Fatal("graph is nondeterministic")
	}
}
