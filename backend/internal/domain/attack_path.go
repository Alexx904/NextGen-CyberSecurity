package domain

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
)

// AttackNodeType categorizza il ruolo del vertice all'interno del grafo topologico.
type AttackNodeType string

const (
	NodeTypeEntrypoint    AttackNodeType = "ENTRYPOINT"
	NodeTypeSubnet        AttackNodeType = "SUBNET"
	NodeTypeHost          AttackNodeType = "HOST"
	NodeTypeService       AttackNodeType = "SERVICE"
	NodeTypeVulnerability AttackNodeType = "VULNERABILITY"
)

// AttackEdgeType definisce la semantica della relazione diretta tra due nodi.
type AttackEdgeType string

const (
	EdgeTypeNetworkAccess   AttackEdgeType = "NETWORK_ACCESS"
	EdgeTypeExploits        AttackEdgeType = "EXPLOITS"
	EdgeTypePivotsTo        AttackEdgeType = "PIVOTS_TO"
	EdgeTypeTrustedRelation AttackEdgeType = "TRUSTED_RELATION"
)

// PathSeverity esprime la gravità aggregata del cammino di penetrazione.
type PathSeverity string

const (
	PathSeverityCritical PathSeverity = "CRITICAL"
	PathSeverityHigh     PathSeverity = "HIGH"
	PathSeverityMedium   PathSeverity = "MEDIUM"
	PathSeverityLow      PathSeverity = "LOW"
)

// AttackNode rappresenta un vertice all'interno della topologia di minaccia.
type AttackNode struct {
	ID        string                 `json:"id"`
	Type      AttackNodeType         `json:"type"`
	Label     string                 `json:"label"`
	IP        string                 `json:"ip,omitempty"`
	Port      int                    `json:"port,omitempty"`
	Protocol  string                 `json:"protocol,omitempty"`
	Severity  string                 `json:"severity,omitempty"`
	CVSS      float64                `json:"cvss,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// AttackSeverityFromCVSS converte uno score CVSS nella corretta severità allineata allo standard (SEC-P2-01)
func AttackSeverityFromCVSS(score float64) string {
	switch {
	case score >= 9.0:
		return string(SeverityCritical)
	case score >= 7.0:
		return string(SeverityHigh)
	case score >= 4.0:
		return string(SeverityMedium)
	case score > 0.0:
		return string(SeverityLow)
	default:
		return string(SeverityInfo)
	}
}

// AttackEdge rappresenta una connessione orientata tra due nodi del grafo con supporto a relazioni inferite (SEC-P2-01).
type AttackEdge struct {
	ID             string         `json:"id"`
	Source         string         `json:"source"`
	Target         string         `json:"target"`
	Type           AttackEdgeType `json:"type"`
	Label          string         `json:"label"`
	Weight         float64        `json:"weight"`
	Inferred       bool           `json:"inferred,omitempty"`
	InferenceBasis string         `json:"inference_basis,omitempty"`
	Confidence     float64        `json:"confidence,omitempty"`
}

// AttackPath descrive una catena completa e ordinata di compromissione da un entrypoint a un asset interno.
type AttackPath struct {
	ID             string       `json:"id"`
	Name           string       `json:"name"`
	NodeIDs        []string     `json:"node_ids"`
	EdgeIDs        []string     `json:"edge_ids"`
	CumulativeRisk float64      `json:"cumulative_risk"`
	Severity       PathSeverity `json:"severity"`
	HopCount       int          `json:"hop_count"`
	TargetAsset    string       `json:"target_asset"`
	Description    string       `json:"description,omitempty"`
}

// Chokepoint rappresenta un nodo o servizio di massima centralità la cui remediation neutralizza molteplici cammini.
type Chokepoint struct {
	NodeID         string  `json:"node_id"`
	Label          string  `json:"label"`
	MitigatedPaths int     `json:"mitigated_paths"`
	MitigatedRisk  float64 `json:"mitigated_risk"`
	Remediation    string  `json:"remediation"`
}

// AttackGraphSummary riassume gli indicatori sintetici del grafo.
type AttackGraphSummary struct {
	TotalNodes       int     `json:"total_nodes"`
	TotalEdges       int     `json:"total_edges"`
	TotalPaths       int     `json:"total_paths"`
	CriticalPaths    int     `json:"critical_paths"`
	HighestRiskScore float64 `json:"highest_risk_score"`
	TopChokepoint    string  `json:"top_chokepoint,omitempty"`
}

// AttackGraph modella l'intera struttura del grafo di attacco per il tenant.
type AttackGraph struct {
	TenantID    uuid.UUID          `json:"tenant_id"`
	GeneratedAt time.Time          `json:"generated_at"`
	Nodes       []AttackNode       `json:"nodes"`
	Edges       []AttackEdge       `json:"edges"`
	Paths       []AttackPath       `json:"paths"`
	Chokepoints []Chokepoint       `json:"chokepoints"`
	Summary     AttackGraphSummary `json:"summary"`
}

// CalculateCumulativeRisk calcola il punteggio di rischio probabilistico (0.0 - 100.0) di una sequenza di nodi (SEC-P2-02).
func CalculateCumulativeRisk(pathNodes []AttackNode, hopCount int) float64 {
	if len(pathNodes) == 0 {
		return 0.0
	}

	var hasCriticalVuln bool
	var hasHighVuln bool
	vulnCount := 0

	// Probabilità congiunta che nessun exploit riesca: prod(1 - p_i)
	probNone := 1.0

	for _, node := range pathNodes {
		if node.Type == NodeTypeVulnerability && node.CVSS > 0 {
			vulnCount++
			if node.CVSS >= 9.0 {
				hasCriticalVuln = true
			} else if node.CVSS >= 7.0 {
				hasHighVuln = true
			}
			// Normalizzazione probabilistica per vulnerabilità singola (p_i in [0.05, 0.95])
			p := math.Min(0.95, math.Max(0.05, (node.CVSS/10.0)*0.85))
			probNone *= (1.0 - p)
		}
	}

	var baseRisk float64
	if vulnCount > 0 {
		// P = 1 - prod(1 - p_i)
		combinedProb := 1.0 - probNone
		baseRisk = combinedProb * 100.0
	} else {
		// Percorso senza vulnerabilità note (puramente topologico)
		baseRisk = 15.0
	}

	// Modulatore per hop: hop minori = maggiore esposizione diretta dell'asset
	effectiveHops := math.Max(1.0, float64(hopCount))
	hopFactor := 1.0 + (0.5 / math.Sqrt(effectiveHops))
	risk := baseRisk * hopFactor

	if hasCriticalVuln && risk < 85.0 {
		risk = 85.0
	} else if hasHighVuln && risk < 65.0 {
		risk = 65.0
	}

	if risk > 100.0 {
		risk = 100.0
	}
	return math.Round(risk*10) / 10
}

// DeterminePathSeverity restituisce il livello di severità in base al punteggio numerico di rischio.
func DeterminePathSeverity(riskScore float64) PathSeverity {
	switch {
	case riskScore >= 75.0:
		return PathSeverityCritical
	case riskScore >= 55.0:
		return PathSeverityHigh
	case riskScore >= 35.0:
		return PathSeverityMedium
	default:
		return PathSeverityLow
	}
}

// CalculateChokepoints individua i vertici intermedi presenti nel maggior numero di percorsi critici.
func CalculateChokepoints(paths []AttackPath, nodes []AttackNode) []Chokepoint {
	if len(paths) == 0 || len(nodes) == 0 {
		return nil
	}

	nodeMap := make(map[string]AttackNode, len(nodes))
	for _, n := range nodes {
		nodeMap[n.ID] = n
	}

	type stats struct {
		pathsCount int
		riskSum    float64
	}
	nodeStats := make(map[string]*stats)

	for _, path := range paths {
		if len(path.NodeIDs) <= 2 {
			continue // Percorsi senza nodi intermedi non hanno chokepoint candidati
		}
		// Considera solo i nodi intermedi (esclude il primo entrypoint e l'asset finale)
		intermediates := path.NodeIDs[1 : len(path.NodeIDs)-1]
		seen := make(map[string]bool)
		for _, nodeID := range intermediates {
			if seen[nodeID] {
				continue
			}
			seen[nodeID] = true
			if nodeStats[nodeID] == nil {
				nodeStats[nodeID] = &stats{}
			}
			nodeStats[nodeID].pathsCount++
			nodeStats[nodeID].riskSum += path.CumulativeRisk
		}
	}

	chokepoints := make([]Chokepoint, 0, len(nodeStats))
	for nodeID, st := range nodeStats {
		n, exists := nodeMap[nodeID]
		if !exists {
			continue
		}
		var remediation string
		switch n.Type {
		case NodeTypeVulnerability:
			remediation = fmt.Sprintf("Applicare patch immediata o aggiornamento del servizio per sanare la vulnerabilità '%s'.", n.Label)
		case NodeTypeHost:
			remediation = fmt.Sprintf("Isolare l'host '%s' o restringere le regole firewall per bloccare il lateral movement verso la rete interna.", n.Label)
		case NodeTypeService:
			remediation = fmt.Sprintf("Chiudere la porta %d o abilitare autenticazione forte e whitelist IP sul servizio '%s'.", n.Port, n.Label)
		default:
			remediation = fmt.Sprintf("Rimuovere o mitigare il nodo di transito '%s'.", n.Label)
		}

		chokepoints = append(chokepoints, Chokepoint{
			NodeID:         nodeID,
			Label:          n.Label,
			MitigatedPaths: st.pathsCount,
			MitigatedRisk:  math.Round(st.riskSum*10) / 10,
			Remediation:    remediation,
		})
	}

	// Ordina per MitigatedRisk decrescente, poi MitigatedPaths decrescente
	sort.Slice(chokepoints, func(i, j int) bool {
		if chokepoints[i].MitigatedRisk == chokepoints[j].MitigatedRisk {
			return chokepoints[i].MitigatedPaths > chokepoints[j].MitigatedPaths
		}
		return chokepoints[i].MitigatedRisk > chokepoints[j].MitigatedRisk
	})

	return chokepoints
}

// Validate esegue la validazione di consistenza interna del grafo di attacco.
func (g *AttackGraph) Validate() error {
	if g.TenantID == uuid.Nil {
		return errors.New("attack graph tenant_id must not be nil")
	}

	nodeSet := make(map[string]bool, len(g.Nodes))
	for _, n := range g.Nodes {
		if n.ID == "" {
			return errors.New("attack node id must not be empty")
		}
		if nodeSet[n.ID] {
			return fmt.Errorf("duplicate attack node id: %s", n.ID)
		}
		nodeSet[n.ID] = true
	}

	for _, e := range g.Edges {
		if e.ID == "" {
			return errors.New("attack edge id must not be empty")
		}
		if !nodeSet[e.Source] {
			return fmt.Errorf("edge %s references unknown source node: %s", e.ID, e.Source)
		}
		if !nodeSet[e.Target] {
			return fmt.Errorf("edge %s references unknown target node: %s", e.ID, e.Target)
		}
	}

	for _, p := range g.Paths {
		if p.ID == "" {
			return errors.New("attack path id must not be empty")
		}
		for _, nodeID := range p.NodeIDs {
			if !nodeSet[nodeID] {
				return fmt.Errorf("path %s references unknown node: %s", p.ID, nodeID)
			}
		}
	}

	return nil
}

