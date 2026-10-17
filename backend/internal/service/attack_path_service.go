package service

import (
	"context"
	"fmt"
	"math"
	"net/netip"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"nextgen-cybersec/internal/domain"
)

type AttackPathService interface {
	GenerateAttackGraph(context.Context, uuid.UUID, float64) (*domain.AttackGraph, error)
	GetChokepoints(context.Context, uuid.UUID) ([]domain.Chokepoint, error)
}

type attackPathService struct {
	targetRepo  domain.TargetRepository
	findingRepo domain.FindingRepository
}

// SSH inventory is deliberately not evidence of trust between two remote hosts.
func NewAttackPathService(targets domain.TargetRepository, findings domain.FindingRepository, _ domain.SSHServerRepository) AttackPathService {
	return &attackPathService{targetRepo: targets, findingRepo: findings}
}

func (s *attackPathService) GenerateAttackGraph(ctx context.Context, tenant uuid.UUID, minCVSS float64) (*domain.AttackGraph, error) {
	if tenant == uuid.Nil || math.IsNaN(minCVSS) || math.IsInf(minCVSS, 0) || minCVSS < 0 || minCVSS > 10 {
		return nil, domain.ErrValidation
	}
	const maxPaths = 50
	g := &domain.AttackGraph{TenantID: tenant, GeneratedAt: time.Now().UTC(), Nodes: []domain.AttackNode{}, Edges: []domain.AttackEdge{}, Paths: []domain.AttackPath{}, Chokepoints: []domain.Chokepoint{}, Analysis: domain.AttackGraphAnalysis{
		Model: "heuristic-review-v2", MaxPaths: maxPaths, Warnings: []string{
			"Accesso iniziale ipotetico: esposizione Internet, firewall e raggiungibilità fra host non verificati.",
			"Una CVE correlata e un CVSS alto non provano exploit, RCE o compromissione. Le ipotesi laterali richiedono verifica di rete e prerequisiti.",
			"Indice 0–100 euristico, non probabilità. I chokepoint riguardano solo i percorsi modellati, non garantiscono neutralizzazione.",
			"Finding storici: ultima osservazione per endpoint, senza garanzia che il servizio sia ancora presente. FQDN e IPv6 non producono ipotesi di vicinanza IPv4.",
		},
	}}
	targets := []*domain.Target{}
	for offset := 0; offset < 2000; {
		page, total, err := s.targetRepo.List(ctx, tenant, domain.TargetFilter{Limit: 200, Offset: offset})
		if err != nil {
			return nil, fmt.Errorf("lettura target: %w", err)
		}
		targets = append(targets, page...)
		offset += len(page)
		if offset >= total {
			break
		}
		if len(page) == 0 || offset >= 2000 {
			g.Analysis.InventoryTruncated = true
			break
		}
	}
	findings := []*domain.Finding{}
	for offset := 0; offset < 10000; {
		page, total, err := s.findingRepo.List(ctx, tenant, domain.FindingFilter{Limit: 500, Offset: offset})
		if err != nil {
			return nil, fmt.Errorf("lettura finding: %w", err)
		}
		findings = append(findings, page...)
		offset += len(page)
		if offset >= total {
			break
		}
		if len(page) == 0 || offset >= 10000 {
			g.Analysis.InventoryTruncated = true
			break
		}
	}
	g.Analysis.TargetCount = len(targets)
	g.Analysis.FindingCount = len(findings)
	if g.Analysis.InventoryTruncated {
		g.Analysis.Warnings = append(g.Analysis.Warnings, "Inventario parziale: raggiunto il limite di lettura. Assenza di percorsi non significa assenza di rischio.")
	}
	root := "entry-internet" // Keep the identifier compatible; the label no longer asserts Internet exposure.
	g.Nodes = append(g.Nodes, domain.AttackNode{ID: root, Type: domain.NodeTypeEntrypoint, Label: "Origine di analisi (accesso ipotetico)", Severity: "INFO"})
	hosts := map[string]string{}
	for _, t := range targets {
		if t == nil || t.TenantID != uuid.Nil && t.TenantID != tenant {
			continue
		}
		if t.Type == domain.TargetTypeIP || t.Type == domain.TargetTypeFQDN {
			hosts[t.Value] = t.Value
			if t.Description != nil && *t.Description != "" {
				hosts[t.Value] = *t.Description + " (" + t.Value + ")"
			}
		}
	}
	// Select the newest scan observed for each endpoint. Deduplicate repeated CVEs inside it.
	endpoint := func(f *domain.Finding) string {
		return fmt.Sprintf("%s|%s|%d", f.HostIP, strings.ToLower(f.Protocol), f.Port)
	}
	latest := map[string]*domain.Finding{}
	for _, f := range findings {
		if f == nil || f.TenantID != uuid.Nil && f.TenantID != tenant || f.HostIP == "" {
			continue
		}
		key := endpoint(f)
		old := latest[key]
		if old == nil || f.CreatedAt.After(old.CreatedAt) || f.CreatedAt.Equal(old.CreatedAt) && f.ScanID.String() > old.ScanID.String() {
			latest[key] = f
		}
	}
	unique := map[string]*domain.Finding{}
	for _, f := range findings {
		if f == nil || f.TenantID != uuid.Nil && f.TenantID != tenant {
			continue
		}
		last := latest[endpoint(f)]
		if last == nil || f.ScanID != last.ScanID {
			continue
		}
		if f.Port < 1 || f.Port > 65535 {
			continue
		}
		proto := strings.ToLower(f.Protocol)
		if proto != "tcp" && proto != "udp" {
			continue
		}
		if f.PortState != "" && f.PortState != "open" {
			continue
		}
		if _, ok := hosts[f.HostIP]; !ok {
			hosts[f.HostIP] = f.HostIP
		}
		cve := ""
		if f.CVEID != nil {
			cve = *f.CVEID
		}
		key := endpoint(f) + "|" + cve
		old := unique[key]
		if old == nil || f.CreatedAt.After(old.CreatedAt) || f.CreatedAt.Equal(old.CreatedAt) && f.ID.String() > old.ID.String() {
			unique[key] = f
		}
	}
	sortedHosts := make([]string, 0, len(hosts))
	for h := range hosts {
		sortedHosts = append(sortedHosts, h)
	}
	sort.Strings(sortedHosts)
	fs := make([]*domain.Finding, 0, len(unique))
	for _, f := range unique {
		fs = append(fs, f)
	}
	score := func(f *domain.Finding) float64 {
		if f.CVSSScore != nil && !math.IsNaN(*f.CVSSScore) && !math.IsInf(*f.CVSSScore, 0) {
			return math.Max(0, math.Min(10, *f.CVSSScore))
		}
		return 0
	}
	sort.Slice(fs, func(i, j int) bool {
		if score(fs[i]) != score(fs[j]) {
			return score(fs[i]) > score(fs[j])
		}
		return endpoint(fs[i])+stringValue(fs[i].CVEID) < endpoint(fs[j])+stringValue(fs[j].CVEID)
	})
	nodeIndex := map[string]int{}
	addNode := func(n domain.AttackNode) {
		if i, ok := nodeIndex[n.ID]; ok {
			if n.CVSS > g.Nodes[i].CVSS {
				g.Nodes[i].CVSS = n.CVSS
				g.Nodes[i].Severity = n.Severity
			}
			return
		}
		nodeIndex[n.ID] = len(g.Nodes)
		g.Nodes = append(g.Nodes, n)
	}
	nodeIndex[root] = 0
	edges := map[string]string{}
	addEdge := func(src, dst string, kind domain.AttackEdgeType, label, basis string, inferred bool) string {
		key := src + "\x00" + dst
		if id, ok := edges[key]; ok {
			return id
		}
		id := fmt.Sprintf("edge-%d", len(g.Edges)+1)
		edges[key] = id
		g.Edges = append(g.Edges, domain.AttackEdge{ID: id, Source: src, Target: dst, Type: kind, Label: label, Weight: 1, Inferred: inferred, InferenceBasis: basis})
		return id
	}
	hostID := func(h string) string { return "host-" + url.PathEscape(h) }
	for _, h := range sortedHosts {
		addNode(domain.AttackNode{ID: hostID(h), Type: domain.NodeTypeHost, IP: h, Label: hosts[h], Severity: "INFO"})
		addEdge(root, hostID(h), domain.EdgeTypeNetworkAccess, "Accesso da verificare", "Origine di scansione e policy di ingresso non disponibili; nessuna esposizione Internet dimostrata", true)
	}
	addPath := func(ids []string, name, target, description string) {
		if len(g.Paths) >= maxPaths {
			g.Analysis.PathsTruncated = true
			return
		}
		nodes := make([]domain.AttackNode, 0, len(ids))
		edgeIDs := make([]string, 0, len(ids)-1)
		for i, id := range ids {
			nodes = append(nodes, g.Nodes[nodeIndex[id]])
			if i > 0 {
				edgeIDs = append(edgeIDs, edges[ids[i-1]+"\x00"+id])
			}
		}
		risk := domain.CalculateCumulativeRisk(nodes, len(ids)-1)
		g.Paths = append(g.Paths, domain.AttackPath{ID: fmt.Sprintf("path-%d", len(g.Paths)+1), Name: name, NodeIDs: ids, EdgeIDs: edgeIDs, CumulativeRisk: risk, Severity: domain.DeterminePathSeverity(risk), HopCount: len(ids) - 1, TargetAsset: target, Description: description})
	}
	// First build all observed associations. Inferred access is kept visually separate.
	for _, f := range fs {
		proto := strings.ToLower(f.Protocol)
		srv := fmt.Sprintf("srv-%s-%d-%s", url.PathEscape(f.HostIP), f.Port, proto)
		addNode(domain.AttackNode{ID: srv, Type: domain.NodeTypeService, IP: f.HostIP, Port: f.Port, Protocol: proto, Label: fmt.Sprintf("%s %d/%s", stringValue(f.ServiceName), f.Port, proto), CVSS: score(f), Severity: domain.AttackSeverityFromCVSS(score(f)), Metadata: map[string]interface{}{"observed_at": f.CreatedAt, "scan_id": f.ScanID, "port_state": f.PortState}})
		addNode(domain.AttackNode{ID: hostID(f.HostIP), Type: domain.NodeTypeHost, CVSS: score(f), Severity: domain.AttackSeverityFromCVSS(score(f))})
		addEdge(hostID(f.HostIP), srv, domain.EdgeTypeNetworkAccess, "Servizio rilevato", "", false)
		if stringValue(f.CVEID) == "" || score(f) < minCVSS {
			continue
		}
		vuln := srv + "-" + url.PathEscape(*f.CVEID)
		addNode(domain.AttackNode{ID: vuln, Type: domain.NodeTypeVulnerability, IP: f.HostIP, Port: f.Port, Protocol: proto, Label: *f.CVEID, CVSS: score(f), Severity: domain.AttackSeverityFromCVSS(score(f)), Metadata: map[string]interface{}{"cve_id": *f.CVEID, "host_ip": f.HostIP, "confidence": f.Confidence, "detection_method": f.DetectionMethod, "scan_id": f.ScanID, "observed_at": f.CreatedAt}})
		addEdge(srv, vuln, domain.EdgeTypeExploits, "CVE correlata (non exploit)", "Associazione del finding al servizio; sfruttabilità e prerequisiti non verificati", true)
	}
	// Direct CVE review paths get priority over speculative lateral hypotheses under the cap.
	for _, f := range fs {
		if stringValue(f.CVEID) == "" || score(f) < minCVSS {
			continue
		}
		srv := fmt.Sprintf("srv-%s-%d-%s", url.PathEscape(f.HostIP), f.Port, strings.ToLower(f.Protocol))
		vuln := srv + "-" + url.PathEscape(*f.CVEID)
		addPath([]string{root, hostID(f.HostIP), srv, vuln}, "Analisi "+f.HostIP+": "+*f.CVEID, f.HostIP, "Verificare versione, configurazione e prerequisiti della CVE. Accesso iniziale ipotetico; nessun exploit eseguito.")
	}
	for _, f := range fs {
		if stringValue(f.CVEID) == "" || score(f) < math.Max(7, minCVSS) {
			continue
		}
		srv := fmt.Sprintf("srv-%s-%d-%s", url.PathEscape(f.HostIP), f.Port, strings.ToLower(f.Protocol))
		vuln := srv + "-" + url.PathEscape(*f.CVEID)
		for _, h := range sortedHosts {
			if !sameCandidateSegment(f.HostIP, h) {
				continue
			}
			if len(g.Paths) >= maxPaths {
				g.Analysis.PathsTruncated = true
				break
			}
			addEdge(vuln, hostID(h), domain.EdgeTypePivotsTo, "Ipotesi laterale /24", "Indirizzi nella stessa /24 IPv4 convenzionale; maschera, routing, ACL, credenziali e capacità di esecuzione non verificati", true)
			addPath([]string{root, hostID(f.HostIP), srv, vuln, hostID(h)}, "Ipotesi laterale: "+f.HostIP+" → "+h, h, "Solo vicinanza di indirizzi /24, non prova di rete condivisa o lateral movement. Verificare controllo dell'host sorgente, policy e accesso al target.")
		}
	}
	if len(g.Paths) == 0 && minCVSS == 0 {
		for _, h := range sortedHosts {
			addPath([]string{root, hostID(h)}, "Inventario: "+h, h, "Asset censito; accessibilità e compromissione non dimostrate.")
		}
	}
	sort.SliceStable(g.Paths, func(i, j int) bool { return g.Paths[i].CumulativeRisk > g.Paths[j].CumulativeRisk })
	g.Chokepoints = domain.CalculateChokepoints(g.Paths, g.Nodes)
	if g.Chokepoints == nil {
		g.Chokepoints = []domain.Chokepoint{}
	}
	g.Summary = domain.AttackGraphSummary{TotalNodes: len(g.Nodes), TotalEdges: len(g.Edges), TotalPaths: len(g.Paths)}
	for _, p := range g.Paths {
		if p.Severity == domain.PathSeverityCritical {
			g.Summary.CriticalPaths++
		}
		g.Summary.HighestRiskScore = math.Max(g.Summary.HighestRiskScore, p.CumulativeRisk)
	}
	if len(g.Chokepoints) > 0 {
		g.Summary.TopChokepoint = g.Chokepoints[0].Label
	}
	if err := g.Validate(); err != nil {
		return nil, fmt.Errorf("validazione grafo: %w", err)
	}
	return g, nil
}

func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
func sameCandidateSegment(a, b string) bool {
	x, ex := netip.ParseAddr(a)
	y, ey := netip.ParseAddr(b)
	if ex != nil || ey != nil || !x.Is4() || !y.Is4() || x == y || !x.IsGlobalUnicast() || !y.IsGlobalUnicast() {
		return false
	}
	return netip.PrefixFrom(x, 24).Masked().Contains(y)
}
func (s *attackPathService) GetChokepoints(ctx context.Context, tenant uuid.UUID) ([]domain.Chokepoint, error) {
	g, e := s.GenerateAttackGraph(ctx, tenant, 0)
	if e != nil {
		return nil, e
	}
	return g.Chokepoints, nil
}
