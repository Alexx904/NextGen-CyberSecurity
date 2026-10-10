export type AttackNodeType = 'ENTRYPOINT' | 'SUBNET' | 'HOST' | 'SERVICE' | 'VULNERABILITY';
export type AttackEdgeType = 'NETWORK_ACCESS' | 'EXPLOITS' | 'PIVOTS_TO' | 'TRUSTED_RELATION';
export type PathSeverity = 'CRITICAL' | 'HIGH' | 'MEDIUM' | 'LOW';

export interface AttackNode {
  id: string;
  type: AttackNodeType;
  label: string;
  ip?: string;
  port?: number;
  protocol?: string;
  severity?: string;
  cvss?: number;
  metadata?: Record<string, unknown>;
}

export interface AttackEdge {
  id: string;
  source: string;
  target: string;
  type: AttackEdgeType;
  label: string;
  weight: number;
}

export interface AttackPath {
  id: string;
  name: string;
  node_ids: string[];
  edge_ids: string[];
  cumulative_risk: number;
  severity: PathSeverity;
  hop_count: number;
  target_asset: string;
  description?: string;
}

export interface Chokepoint {
  node_id: string;
  label: string;
  mitigated_paths: number;
  mitigated_risk: number;
  remediation: string;
}

export interface AttackGraphSummary {
  total_nodes: number;
  total_edges: number;
  total_paths: number;
  critical_paths: number;
  highest_risk_score: number;
  top_chokepoint?: string;
}

export interface AttackGraph {
  tenant_id: string;
  generated_at: string;
  nodes: AttackNode[];
  edges: AttackEdge[];
  paths: AttackPath[];
  chokepoints: Chokepoint[];
  summary: AttackGraphSummary;
}

export interface AttackGraphFilterParams {
  min_cvss?: number;
}

