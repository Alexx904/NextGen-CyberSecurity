import type { AttackGraph, AttackNode, AttackPath } from '../../types/attackPath';

function ipv4(value: string): number | null {
  const parts = value.split('.');
  if (parts.length !== 4 || parts.some(p => !/^\d{1,3}$/.test(p) || Number(p) > 255)) return null;
  return parts.reduce((n, p) => (n * 256 + Number(p)) >>> 0, 0);
}

export function matchesScope(host: string, values: Set<string>): boolean {
  if (values.has(host)) return true;
  const address = ipv4(host);
  if (address === null) return false;
  for (const value of values) {
    if (value.includes('-')) {
      const [start, end] = value.split('-');
      const first = ipv4(start);
      const last = ipv4(end.includes('.') ? end : start.slice(0, start.lastIndexOf('.') + 1) + end);
      if (first !== null && last !== null && address >= first && address <= last) return true;
      continue;
    }
    const [network, bits] = value.split('/');
    if (bits === undefined || !/^\d+$/.test(bits) || Number(bits) > 32) continue;
    const base = ipv4(network);
    const mask = Number(bits) === 0 ? 0 : (0xffffffff << (32 - Number(bits))) >>> 0;
    if (base !== null && (address & mask) === (base & mask)) return true;
  }
  return false;
}

// Company scope is a display projection; tenant authorization remains server-side.
export function scopeAttackGraph(graph: AttackGraph, values: Set<string>): AttackGraph {
  const nodes = graph.nodes.filter(n => n.type === 'ENTRYPOINT' || matchesScope(n.ip || String(n.metadata?.host_ip || ''), values));
  const ids = new Set(nodes.map(n => n.id));
  const edges = graph.edges.filter(e => ids.has(e.source) && ids.has(e.target));
  const paths = graph.paths.filter(p => p.node_ids.every(id => ids.has(id)));
  const counts = new Map<string, { count: number; risk: number }>();
  paths.forEach(p => new Set(p.node_ids.slice(1, -1)).forEach(id => {
    const stat = counts.get(id) || { count: 0, risk: 0 };
    counts.set(id, { count: stat.count + 1, risk: stat.risk + p.cumulative_risk });
  }));
  const chokepoints = graph.chokepoints.filter(c => counts.has(c.node_id)).map(c => ({
    ...c, mitigated_paths: counts.get(c.node_id)!.count,
    mitigated_risk: Math.round(counts.get(c.node_id)!.risk * 10) / 10,
  })).sort((a, b) => b.mitigated_risk - a.mitigated_risk || b.mitigated_paths - a.mitigated_paths || a.node_id.localeCompare(b.node_id));
  return { ...graph, nodes, edges, paths, chokepoints, summary: {
    total_nodes: nodes.length, total_edges: edges.length, total_paths: paths.length,
    critical_paths: paths.filter(p => p.severity === 'CRITICAL').length,
    highest_risk_score: Math.max(0, ...paths.map(p => p.cumulative_risk)),
    top_chokepoint: chokepoints[0]?.label,
  } };
}

export interface PositionedAttackNode extends AttackNode { x: number; y: number }

export function layoutAttackGraph(graph: AttackGraph, path?: AttackPath | null): PositionedAttackNode[] {
  if (path) {
    const index = new Map(graph.nodes.map(n => [n.id, n]));
    return path.node_ids.flatMap((id, i) => {
      const node = index.get(id);
      return node ? [{ ...node, x: 40 + i * 300, y: 110 }] : [];
    });
  }
  const layers = new Map<string, number>();
  const queue: string[] = [];
  graph.nodes.filter(n => n.type === 'ENTRYPOINT').forEach(n => { layers.set(n.id, 0); queue.push(n.id); });
  const adjacency = new Map<string, string[]>();
  graph.edges.forEach(e => adjacency.set(e.source, [...(adjacency.get(e.source) || []), e.target]));
  // Each vertex is visited once, including graphs containing lateral cycles.
  for (let i = 0; i < queue.length; i++) {
    for (const next of adjacency.get(queue[i]) || []) {
      if (!layers.has(next)) { layers.set(next, layers.get(queue[i])! + 1); queue.push(next); }
    }
  }
  const rows = new Map<number, number>();
  return [...graph.nodes].sort((a, b) => a.id.localeCompare(b.id)).map(n => {
    const fallback = { ENTRYPOINT: 0, HOST: 1, SUBNET: 1, SERVICE: 2, VULNERABILITY: 3 };
    const layer = layers.get(n.id) ?? fallback[n.type];
    const row = rows.get(layer) ?? 0;
    rows.set(layer, row + 1);
    return { ...n, x: 40 + layer * 300, y: 90 + row * 116 };
  });
}
