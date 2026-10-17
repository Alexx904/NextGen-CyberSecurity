import { describe, it, expect } from 'vitest';
import { layoutAttackGraph, matchesScope, scopeAttackGraph } from './attackPathModel';
import type { AttackGraph } from '../../types/attackPath';

const graph: AttackGraph = {
  tenant_id: 'test', generated_at: '2026-10-17T10:00:00Z',
  nodes: [
    { id: 'root', type: 'ENTRYPOINT', label: 'Origine' },
    { id: 'a', type: 'HOST', label: 'A', ip: '10.0.0.1' },
    { id: 's', type: 'SERVICE', label: 'HTTPS', ip: '10.0.0.1', port: 443 },
    { id: 'v', type: 'VULNERABILITY', label: 'CVE', metadata: { host_ip: '10.0.0.1' } },
    { id: 'b', type: 'HOST', label: 'B', ip: '10.9.0.1' },
  ],
  edges: [
    { id: 'e1', source: 'root', target: 'a', type: 'NETWORK_ACCESS', label: 'Ipotesi', weight: 1 },
    { id: 'e2', source: 'a', target: 's', type: 'NETWORK_ACCESS', label: 'Osservato', weight: 1 },
    { id: 'e3', source: 's', target: 'v', type: 'EXPLOITS', label: 'CVE', weight: 1 },
    { id: 'e4', source: 'v', target: 'b', type: 'PIVOTS_TO', label: 'Ipotesi', weight: 1 },
    { id: 'cycle', source: 'b', target: 'a', type: 'PIVOTS_TO', label: 'Ritorno', weight: 1 },
  ],
  paths: [
    { id: 'p1', name: 'Diretto', node_ids: ['root', 'a', 's', 'v'], edge_ids: ['e1', 'e2', 'e3'], cumulative_risk: 80, severity: 'CRITICAL', hop_count: 3, target_asset: '10.0.0.1:443' },
    { id: 'p2', name: 'Cross scope', node_ids: ['root', 'a', 's', 'v', 'b'], edge_ids: ['e1', 'e2', 'e3', 'e4'], cumulative_risk: 75, severity: 'CRITICAL', hop_count: 4, target_asset: '10.9.0.1' },
  ],
  chokepoints: [{ node_id: 'a', label: 'A', mitigated_paths: 2, mitigated_risk: 155, remediation: 'Valutare ACL' }],
  summary: { total_nodes: 5, total_edges: 5, total_paths: 2, critical_paths: 2, highest_risk_score: 80 },
};

describe('Attack graph scope and cycle-safe layout', () => {
  it('empty company has no global fallback', () => {
    const result = scopeAttackGraph(graph, new Set());
    expect(result.nodes.map(n => n.id)).toEqual(['root']);
    expect(result.paths).toEqual([]);
    expect(result.edges).toEqual([]);
    expect(result.chokepoints).toEqual([]);
    expect(result.summary.total_paths).toBe(0);
  });
  it('scopes the entire graph and recomputes KPI/chokepoints including host:port targets', () => {
    const result = scopeAttackGraph(graph, new Set(['10.0.0.0/24']));
    expect(result.paths.map(p => p.id)).toEqual(['p1']);
    expect(result.nodes.map(n => n.id)).toEqual(['root', 'a', 's', 'v']);
    expect(result.edges.map(e => e.id)).toEqual(['e1', 'e2', 'e3']);
    expect(result.chokepoints[0].mitigated_risk).toBe(80);
    expect(result.chokepoints[0].mitigated_paths).toBe(1);
    expect(result.summary.total_nodes).toBe(4);
    expect(result.summary.critical_paths).toBe(1);
  });
  it('never includes a chain originating outside the selected company', () => {
    expect(scopeAttackGraph(graph, new Set(['10.9.0.1'])).paths).toEqual([]);
  });
  it('supports exact FQDN/IPv6 and validates IPv4 CIDR boundaries', () => {
    expect(matchesScope('example.test', new Set(['example.test']))).toBe(true);
    expect(matchesScope('::1', new Set(['::1']))).toBe(true);
    expect(matchesScope('10.0.0.255', new Set(['10.0.0.0/24']))).toBe(true);
    expect(matchesScope('10.0.1.0', new Set(['10.0.0.0/24']))).toBe(false);
    expect(matchesScope('10.0.0.1', new Set(['10.0.0.0/33']))).toBe(false);
    expect(matchesScope('10.0.0.10', new Set(['10.0.0.1-20']))).toBe(true);
    expect(matchesScope('10.0.0.21', new Set(['10.0.0.1-10.0.0.20']))).toBe(false);
  });
  it('terminates on reachable cycles and preserves the entrypoint layer', () => {
    const positions = layoutAttackGraph(graph);
    expect(positions).toHaveLength(5);
    expect(positions.find(n => n.id === 'root')!.x).toBe(40);
    expect(positions.every(n => Number.isFinite(n.x) && Number.isFinite(n.y))).toBe(true);
  });
  it('lays out a selected path in exact sequence with no overlapping cards', () => {
    const positions = layoutAttackGraph(graph, graph.paths[1]);
    expect(positions.map(n => n.id)).toEqual(graph.paths[1].node_ids);
    expect(positions.every((n, i) => i === 0 || n.x - positions[i - 1].x >= 300)).toBe(true);
  });
});
