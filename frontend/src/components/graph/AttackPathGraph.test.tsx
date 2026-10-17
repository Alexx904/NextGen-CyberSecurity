import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { AttackPathGraph } from './AttackPathGraph';
import type { AttackGraph } from '../../types/attackPath';

const graph: AttackGraph = {
  tenant_id: 'test', generated_at: '2026-10-17T10:00:00Z',
  nodes: [
    { id: 'root', type: 'ENTRYPOINT', label: 'Origine ipotetica' },
    { id: 'host', type: 'HOST', label: 'Server', ip: '10.0.0.10' },
    { id: 'service', type: 'SERVICE', label: 'HTTPS', ip: '10.0.0.10', port: 443, protocol: 'tcp' },
    { id: 'other', type: 'HOST', label: 'Altro host' },
  ],
  edges: [
    { id: 'e1', source: 'root', target: 'host', type: 'NETWORK_ACCESS', label: 'Accesso da verificare', inferred: true, weight: 1 },
    { id: 'e2', source: 'host', target: 'service', type: 'NETWORK_ACCESS', label: 'Servizio osservato', weight: 1 },
    { id: 'parallel', source: 'host', target: 'service', type: 'PIVOTS_TO', label: 'Altro arco', inferred: true, weight: 1 },
    { id: 'cycle', source: 'service', target: 'host', type: 'PIVOTS_TO', label: 'Ciclo', inferred: true, weight: 1 },
  ],
  paths: [{ id: 'p', name: 'Analisi', node_ids: ['root', 'host', 'service'], edge_ids: ['e1', 'e2'], hop_count: 2, severity: 'LOW', cumulative_risk: 22.5, target_asset: '10.0.0.10' }],
  chokepoints: [], summary: { total_nodes: 4, total_edges: 4, total_paths: 1, critical_paths: 0, highest_risk_score: 22.5 },
};

describe('Attack graph interaction regressions', () => {
  it('shows only exact path edges and preserves observed/inferred line semantics under highlight', () => {
    render(<AttackPathGraph graph={graph} selectedPathId="p" />);
    expect(document.querySelectorAll('[data-edge-id]')).toHaveLength(2);
    expect(document.querySelector('[data-edge-id="e1"]')?.getAttribute('stroke-dasharray')).toBe('7 5');
    expect(document.querySelector('[data-edge-id="e2"]')?.hasAttribute('stroke-dasharray')).toBe(false);
    fireEvent.click(screen.getByRole('button', { name: 'Mappa completa' }));
    expect(document.querySelectorAll('[data-edge-id]')).toHaveLength(4);
    expect(document.querySelector('[data-edge-id="parallel"]')?.getAttribute('stroke-width')).toBe('1.7');
  });
  it('allows keyboard inspection and resets the zoom', () => {
    const select = vi.fn();
    render(<AttackPathGraph graph={graph} selectedPathId="p" onSelectNode={select} />);
    fireEvent.keyDown(screen.getByRole('button', { name: 'Ispeziona HTTPS' }), { key: 'Enter' });
    expect(select).toHaveBeenCalledWith(expect.objectContaining({ id: 'service', protocol: 'tcp' }));
    fireEvent.click(screen.getByRole('button', { name: 'Zoom Avanti' }));
    expect(screen.getByText('120%')).toBeDefined();
    fireEvent.click(screen.getByRole('button', { name: 'Adatta vista' }));
    expect(screen.getByText('100%')).toBeDefined();
  });
});
