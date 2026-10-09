package service

import (
	"context"
	"fmt"
	"math"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"nextgen-cybersec/internal/domain"
)

// AttackPathService definisce le operazioni per la generazione e consultazione del grafo delle minacce.
type AttackPathService interface {
	GenerateAttackGraph(ctx context.Context, tenantID uuid.UUID, minCVSS float64) (*domain.AttackGraph, error)
	GetChokepoints(ctx context.Context, tenantID uuid.UUID) ([]domain.Chokepoint, error)
}

type attackPathService struct {
	targetRepo    domain.TargetRepository
	findingRepo   domain.FindingRepository
	sshServerRepo domain.SSHServerRepository
}

// NewAttackPathService istanzia un nuovo servizio per l'analisi dei percorsi di attacco.
func NewAttackPathService(
	targetRepo domain.TargetRepository,
	findingRepo domain.FindingRepository,
	sshServerRepo domain.SSHServerRepository,
) AttackPathService {
	return &attackPathService{
		targetRepo:    targetRepo,
		findingRepo:   findingRepo,
		sshServerRepo: sshServerRepo,
	}
}

// GenerateAttackGraph genera il grafo orientato delle minacce e correla i percorsi di penetrazione multi-hop.
func (s *attackPathService) GenerateAttackGraph(ctx context.Context, tenantID uuid.UUID, minCVSS float64) (*domain.AttackGraph, error) {
	if tenantID == uuid.Nil {
		return nil, domain.ErrValidation
	}

	// 1. Recupero dei target registrati per il tenant
	targets, _, err := s.targetRepo.List(ctx, tenantID, domain.TargetFilter{Limit: 200})
	if err != nil {
		return nil, fmt.Errorf("recupero target per attack graph fallito: %w", err)
	}

	// 2. Recupero dei finding per il tenant (senza filtro scan, limit elevato)
	findings, _, err := s.findingRepo.List(ctx, tenantID, domain.FindingFilter{Limit: 500})
	if err != nil {
		return nil, fmt.Errorf("recupero finding per attack graph fallito: %w", err)
	}

	// 3. Recupero dei server SSH remoti per rilevare trust relationships
	sshServers, err := s.sshServerRepo.List(ctx, tenantID)
	if err != nil {
		// Log e fallback con slice vuota
		sshServers = []*domain.SSHServer{}
	}

	graph := &domain.AttackGraph{
		TenantID:    tenantID,
		GeneratedAt: time.Now().UTC(),
		Nodes:       make([]domain.AttackNode, 0),
		Edges:       make([]domain.AttackEdge, 0),
		Paths:       make([]domain.AttackPath, 0),
		Chokepoints: make([]domain.Chokepoint, 0),
	}

	// Nodo radice fisso: Internet Perimeter
	entrypointNode := domain.AttackNode{
		ID:       "entry-internet",
		Type:     domain.NodeTypeEntrypoint,
		Label:    "Perimetro Internet / Attaccante Esterno",
		Severity: string(domain.SeverityLow),
	}
	graph.Nodes = append(graph.Nodes, entrypointNode)

	// Se non ci sono target né finding, restituiamo il grafo minimale
	if len(targets) == 0 && len(findings) == 0 {
		graph.Summary = domain.AttackGraphSummary{
			TotalNodes: 1,
		}
		return graph, nil
	}

	// Mappa host IP -> lista findings
	hostFindingsMap := make(map[string][]*domain.Finding)
	for _, f := range findings {
		hostFindingsMap[f.HostIP] = append(hostFindingsMap[f.HostIP], f)
	}

	// Mappa host IP -> target information
	targetMap := make(map[string]*domain.Target)
	for _, t := range targets {
		targetMap[t.Value] = t
	}

	// Insieme di tutti gli host distinti presenti nei target o nei finding
	allHosts := make(map[string]bool)
	for _, t := range targets {
		if t.Type == domain.TargetTypeIP || t.Type == domain.TargetTypeFQDN {
			allHosts[t.Value] = true
		}
	}
	for ip := range hostFindingsMap {
		allHosts[ip] = true
	}

	// Identificazione degli host esterni vs interni
	externalHosts := make([]string, 0)
	internalHosts := make([]string, 0)

	for host := range allHosts {
		if isPrivateIP(host) {
			internalHosts = append(internalHosts, host)
		} else {
			externalHosts = append(externalHosts, host)
		}
	}

	// Se non ci sono host pubblici identificati, consideriamo i primi host o target come punto di contatto iniziale
	if len(externalHosts) == 0 && len(allHosts) > 0 {
		// Ordina per determinismo
		sortedHosts := make([]string, 0, len(allHosts))
		for h := range allHosts {
			sortedHosts = append(sortedHosts, h)
		}
		sort.Strings(sortedHosts)
		externalHosts = append(externalHosts, sortedHosts[0])
		if len(sortedHosts) > 1 {
			internalHosts = sortedHosts[1:]
		}
	}

	edgeCounter := 1
	nextEdgeID := func() string {
		id := fmt.Sprintf("edge-%d", edgeCounter)
		edgeCounter++
		return id
	}

	addedNodes := make(map[string]bool)
	addedNodes[entrypointNode.ID] = true

	// Host classificati come pivot da cui è possibile estendere l'attacco
	pivotCapableHosts := make(map[string]bool)

	// Creazione nodi Host, Servizi e Vulnerabilità
	for host := range allHosts {
		hostNodeID := fmt.Sprintf("host-%s", sanitizeID(host))
		hostLabel := host
		if t, ok := targetMap[host]; ok && t.Description != nil && *t.Description != "" {
			hostLabel = fmt.Sprintf("%s (%s)", *t.Description, host)
		}

		hostSeverity := string(domain.SeverityLow)
		hostFindings := hostFindingsMap[host]

		// Calcolo massima severità host
		maxCVSS := 0.0
		for _, f := range hostFindings {
			if f.CVSSScore != nil && *f.CVSSScore > maxCVSS {
				maxCVSS = *f.CVSSScore
				hostSeverity = string(f.Severity)
			}
		}

		if !addedNodes[hostNodeID] {
			graph.Nodes = append(graph.Nodes, domain.AttackNode{
				ID:       hostNodeID,
				Type:     domain.NodeTypeHost,
				Label:    hostLabel,
				IP:       host,
				Severity: hostSeverity,
				CVSS:     maxCVSS,
			})
			addedNodes[hostNodeID] = true
		}

		// Se l'host è nel perimetro esterno, colleghiamolo dall'entrypoint
		if sliceContains(externalHosts, host) {
			graph.Edges = append(graph.Edges, domain.AttackEdge{
				ID:     nextEdgeID(),
				Source: entrypointNode.ID,
				Target: hostNodeID,
				Type:   domain.EdgeTypeNetworkAccess,
				Label:  "Esposizione Rete Esterna",
				Weight: 1.0,
			})
		}

		// Aggiunta servizi e vulnerabilità per l'host
		serviceMap := make(map[int]bool)
		for _, f := range hostFindings {
			serviceNodeID := fmt.Sprintf("srv-%s-%d", sanitizeID(host), f.Port)
			serviceLabel := fmt.Sprintf("Port %d/%s", f.Port, f.Protocol)
			if f.ServiceName != nil && *f.ServiceName != "" {
				serviceLabel = fmt.Sprintf("%s (%d)", *f.ServiceName, f.Port)
			}

			if !serviceMap[f.Port] {
				serviceMap[f.Port] = true
				if !addedNodes[serviceNodeID] {
					graph.Nodes = append(graph.Nodes, domain.AttackNode{
						ID:       serviceNodeID,
						Type:     domain.NodeTypeService,
						Label:    serviceLabel,
						IP:       host,
						Port:     f.Port,
						Protocol: f.Protocol,
						Severity: string(f.Severity),
					})
					addedNodes[serviceNodeID] = true

					// Arco Host -> Service
					graph.Edges = append(graph.Edges, domain.AttackEdge{
						ID:     nextEdgeID(),
						Source: hostNodeID,
						Target: serviceNodeID,
						Type:   domain.EdgeTypeNetworkAccess,
						Label:  "Porta Aperta",
						Weight: 1.0,
					})
				}
			}

			// Se il finding ha una CVE con CVSS >= minCVSS, creiamo il nodo vulnerabilità
			cvss := 0.0
			if f.CVSSScore != nil {
				cvss = *f.CVSSScore
			}

			if f.CVEID != nil && *f.CVEID != "" && cvss >= minCVSS {
				vulnNodeID := fmt.Sprintf("vuln-%s-%d-%s", sanitizeID(host), f.Port, sanitizeID(*f.CVEID))
				vulnLabel := *f.CVEID
				if f.ServiceName != nil && *f.ServiceName != "" {
					vulnLabel = fmt.Sprintf("%s: %s", *f.ServiceName, *f.CVEID)
				}

				if !addedNodes[vulnNodeID] {
					graph.Nodes = append(graph.Nodes, domain.AttackNode{
						ID:       vulnNodeID,
						Type:     domain.NodeTypeVulnerability,
						Label:    vulnLabel,
						Severity: string(f.Severity),
						CVSS:     cvss,
						Metadata: map[string]interface{}{
							"cve_id":   *f.CVEID,
							"host_ip":  host,
							"port":     f.Port,
							"protocol": f.Protocol,
						},
					})
					addedNodes[vulnNodeID] = true

					// Arco Service -> Vulnerability
					graph.Edges = append(graph.Edges, domain.AttackEdge{
						ID:     nextEdgeID(),
						Source: serviceNodeID,
						Target: vulnNodeID,
						Type:   domain.EdgeTypeExploits,
						Label:  fmt.Sprintf("Exploit CVSS %.1f", cvss),
						Weight: math.Max(0.1, 10.0-cvss),
					})
				}

				// Se la vulnerabilità è grave (CVSS >= 7.0), l'host diventa vettore di pivot
				if cvss >= 7.0 {
					pivotCapableHosts[host] = true
				}
			}
		}
	}

	// Aggiunta relazioni di fiducia basate sui Server SSH configurati
	for _, srv := range sshServers {
		srvHost := srv.Host
		if allHosts[srvHost] {
			pivotCapableHosts[srvHost] = true
		}
	}

	// Connessioni di Pivot tra host esterni compromessi e host interni sensibili
	for pivotHost := range pivotCapableHosts {
		pivotNodeID := fmt.Sprintf("host-%s", sanitizeID(pivotHost))
		for _, intHost := range internalHosts {
			if intHost == pivotHost {
				continue
			}
			intNodeID := fmt.Sprintf("host-%s", sanitizeID(intHost))

			// Arco di Pivot
			graph.Edges = append(graph.Edges, domain.AttackEdge{
				ID:     nextEdgeID(),
				Source: pivotNodeID,
				Target: intNodeID,
				Type:   domain.EdgeTypePivotsTo,
				Label:  "Lateral Movement / Pivot Rete",
				Weight: 1.5,
			})
		}
	}

	// 4. Costruzione dei Percorsi di Attacco (Attack Paths)
	nodeIndex := make(map[string]domain.AttackNode, len(graph.Nodes))
	for _, n := range graph.Nodes {
		nodeIndex[n.ID] = n
	}

	pathCounter := 1
	// Percorsi diretti verso host esterni e relative vulnerabilità
	for _, extHost := range externalHosts {
		extHostNodeID := fmt.Sprintf("host-%s", sanitizeID(extHost))
		extFindings := hostFindingsMap[extHost]

		for _, f := range extFindings {
			if f.CVEID == nil || (f.CVSSScore != nil && *f.CVSSScore < minCVSS) {
				continue
			}
			srvNodeID := fmt.Sprintf("srv-%s-%d", sanitizeID(extHost), f.Port)
			vulnNodeID := fmt.Sprintf("vuln-%s-%d-%s", sanitizeID(extHost), f.Port, sanitizeID(*f.CVEID))

			pathNodeIDs := []string{entrypointNode.ID, extHostNodeID, srvNodeID, vulnNodeID}
			pathNodes := make([]domain.AttackNode, 0, len(pathNodeIDs))
			for _, nid := range pathNodeIDs {
				if n, ok := nodeIndex[nid]; ok {
					pathNodes = append(pathNodes, n)
				}
			}

			risk := domain.CalculateCumulativeRisk(pathNodes, len(pathNodeIDs)-1)
			pathName := fmt.Sprintf("Perimetro Esterno -> %s (Porta %d: %s)", extHost, f.Port, *f.CVEID)

			graph.Paths = append(graph.Paths, domain.AttackPath{
				ID:             fmt.Sprintf("path-%d", pathCounter),
				Name:           pathName,
				NodeIDs:        pathNodeIDs,
				CumulativeRisk: risk,
				Severity:       domain.DeterminePathSeverity(risk),
				HopCount:       len(pathNodeIDs) - 1,
				TargetAsset:    fmt.Sprintf("%s:%d", extHost, f.Port),
				Description:    fmt.Sprintf("Compromissione diretta del servizio su %s tramite exploit della vulnerabilità %s.", extHost, *f.CVEID),
			})
			pathCounter++

			// Se la vulnerabilità è critica (>= 7.0) e ci sono host interni, generiamo percorsi multi-hop
			if f.CVSSScore != nil && *f.CVSSScore >= 7.0 {
				for _, intHost := range internalHosts {
					if intHost == extHost {
						continue
					}
					intHostNodeID := fmt.Sprintf("host-%s", sanitizeID(intHost))

					multiHopNodeIDs := []string{entrypointNode.ID, extHostNodeID, vulnNodeID, intHostNodeID}
					multiHopNodes := make([]domain.AttackNode, 0, len(multiHopNodeIDs))
					for _, nid := range multiHopNodeIDs {
						if n, ok := nodeIndex[nid]; ok {
							multiHopNodes = append(multiHopNodes, n)
						}
					}

					hopRisk := domain.CalculateCumulativeRisk(multiHopNodes, len(multiHopNodeIDs)-1)
					multiHopName := fmt.Sprintf("Catena Multi-Hop: %s (%s) -> Pivot su %s", extHost, *f.CVEID, intHost)

					graph.Paths = append(graph.Paths, domain.AttackPath{
						ID:             fmt.Sprintf("path-%d", pathCounter),
						Name:           multiHopName,
						NodeIDs:        multiHopNodeIDs,
						CumulativeRisk: hopRisk,
						Severity:       domain.DeterminePathSeverity(hopRisk),
						HopCount:       len(multiHopNodeIDs) - 1,
						TargetAsset:    intHost,
						Description:    fmt.Sprintf("Penetrazione dalla DMZ (%s) tramite %s e successivo lateral movement verso l'asset interno %s.", extHost, *f.CVEID, intHost),
					})
					pathCounter++
				}
			}
		}
	}

	// Se non sono stati individuati percorsi con vulnerabilità, generiamo cammini topologici base
	if len(graph.Paths) == 0 && len(externalHosts) > 0 {
		for _, extHost := range externalHosts {
			extHostNodeID := fmt.Sprintf("host-%s", sanitizeID(extHost))
			pathNodeIDs := []string{entrypointNode.ID, extHostNodeID}
			pathNodes := []domain.AttackNode{entrypointNode, nodeIndex[extHostNodeID]}
			risk := domain.CalculateCumulativeRisk(pathNodes, 1)

			graph.Paths = append(graph.Paths, domain.AttackPath{
				ID:             fmt.Sprintf("path-%d", pathCounter),
				Name:           fmt.Sprintf("Accesso di Rete Esterna verso %s", extHost),
				NodeIDs:        pathNodeIDs,
				CumulativeRisk: risk,
				Severity:       domain.DeterminePathSeverity(risk),
				HopCount:       1,
				TargetAsset:    extHost,
				Description:    fmt.Sprintf("Esposizione dell'host %s al perimetro di rete.", extHost),
			})
			pathCounter++
		}
	}

	// Ordina percorsi per CumulativeRisk decrescente
	sort.Slice(graph.Paths, func(i, j int) bool {
		return graph.Paths[i].CumulativeRisk > graph.Paths[j].CumulativeRisk
	})

	// 5. Calcolo dei Chokepoints
	graph.Chokepoints = domain.CalculateChokepoints(graph.Paths, graph.Nodes)

	// 6. Calcolo del Summary
	criticalCount := 0
	highestRisk := 0.0
	for _, p := range graph.Paths {
		if p.CumulativeRisk > highestRisk {
			highestRisk = p.CumulativeRisk
		}
		if p.Severity == domain.PathSeverityCritical {
			criticalCount++
		}
	}

	topChokepoint := ""
	if len(graph.Chokepoints) > 0 {
		topChokepoint = graph.Chokepoints[0].Label
	}

	graph.Summary = domain.AttackGraphSummary{
		TotalNodes:       len(graph.Nodes),
		TotalEdges:       len(graph.Edges),
		TotalPaths:       len(graph.Paths),
		CriticalPaths:    criticalCount,
		HighestRiskScore: highestRisk,
		TopChokepoint:    topChokepoint,
	}

	// 7. Validazione finale
	if err := graph.Validate(); err != nil {
		return nil, fmt.Errorf("validazione grafo generato fallita: %w", err)
	}

	return graph, nil
}

// GetChokepoints estrae l'elenco prioritario dei punti di mitigazione (Chokepoints).
func (s *attackPathService) GetChokepoints(ctx context.Context, tenantID uuid.UUID) ([]domain.Chokepoint, error) {
	graph, err := s.GenerateAttackGraph(ctx, tenantID, 0.0)
	if err != nil {
		return nil, err
	}
	return graph.Chokepoints, nil
}

// Helper per identificare se un indirizzo IP è privato (RFC 1918)
func isPrivateIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		// Se è un dominio (FQDN) o stringa non IP, assumiamo esterno
		return false
	}
	return ip.IsPrivate() || ip.IsLoopback()
}

func sanitizeID(s string) string {
	s = strings.ReplaceAll(s, ".", "-")
	s = strings.ReplaceAll(s, ":", "-")
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ReplaceAll(s, " ", "-")
	return s
}

func sliceContains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

