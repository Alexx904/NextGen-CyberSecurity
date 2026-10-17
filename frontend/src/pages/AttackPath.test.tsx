import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { AttackPath } from './AttackPath';
import { attackPathApi } from '../api/attackPath';
import { AttackGraph } from '../types/attackPath';
import type { TargetGroupNode, TargetItem } from '../types/target';

const companyState = vi.hoisted(() => ({
  selectedCompanyId: null as string | null,
  selectedCompany: null,
  departments: [] as TargetGroupNode[],
  allTargets: [] as TargetItem[],
  companyTargetValues: new Set<string>(),
}));
vi.mock('../context/CompanyContext', () => ({ useCompany: () => companyState }));

vi.mock('../api/attackPath', () => ({
  attackPathApi: {
    getGraph: vi.fn(),
    getChokepoints: vi.fn(),
  },
}));

describe('AttackPath Page - Visual Graph Explorer & Chokepoints', () => {
  const mockGraphData: AttackGraph = {
    tenant_id: '00000000-0000-0000-0000-000000000001',
    generated_at: '2026-10-10T12:00:00Z',
    nodes: [
      {
        id: 'entry-internet',
        type: 'ENTRYPOINT',
        label: 'Perimetro Internet / Attaccante Esterno',
        severity: 'LOW',
      },
      {
        id: 'host-192-168-1-10',
        type: 'HOST',
        label: 'Web Server DMZ (192.168.1.10)',
        ip: '192.168.1.10',
        severity: 'CRITICAL',
        cvss: 7.5,
      },
      {
        id: 'srv-192-168-1-10-80',
        type: 'SERVICE',
        label: 'http (80)',
        ip: '192.168.1.10',
        port: 80,
        protocol: 'tcp',
        severity: 'CRITICAL',
      },
      {
        id: 'vuln-cve-2021-41773',
        type: 'VULNERABILITY',
        ip: '192.168.1.10',
        label: 'Apache Path Traversal RCE (CVE-2021-41773)',
        severity: 'HIGH',
        cvss: 7.5,
      },
      {
        id: 'host-10-0-0-50',
        type: 'HOST',
        label: 'Database Interno (10.0.0.50)',
        ip: '10.0.0.50',
        severity: 'HIGH',
      },
    ],
    edges: [
      {
        id: 'edge-1',
        source: 'entry-internet',
        target: 'host-192-168-1-10',
        type: 'NETWORK_ACCESS',
        label: 'Esposizione Rete Esterna',
        weight: 1.0,
      },
      {
        id: 'edge-2',
        source: 'host-192-168-1-10',
        target: 'srv-192-168-1-10-80',
        type: 'NETWORK_ACCESS',
        label: 'Porta Aperta',
        weight: 1.0,
      },
      {
        id: 'edge-3',
        source: 'srv-192-168-1-10-80',
        target: 'vuln-cve-2021-41773',
        type: 'EXPLOITS',
        label: 'RCE Exploit',
        weight: 1.0,
      },
      {
        id: 'edge-4',
        source: 'vuln-cve-2021-41773',
        target: 'host-10-0-0-50',
        type: 'PIVOTS_TO',
        label: 'Pivot Rete Interna',
        weight: 2.0,
      },
    ],
    paths: [
      {
        id: 'path-1',
        name: 'Compromissione DMZ verso Database Interno',
        node_ids: [
          'entry-internet',
          'host-192-168-1-10',
          'srv-192-168-1-10-80',
          'vuln-cve-2021-41773',
          'host-10-0-0-50',
        ],
        edge_ids: ['edge-1', 'edge-2', 'edge-3', 'edge-4'],
        cumulative_risk: 86.4,
        severity: 'CRITICAL',
        hop_count: 4,
        target_asset: 'Database Interno (10.0.0.50)',
        description: 'Lateral movement da DMZ a LAN interna',
      },
    ],
    chokepoints: [
      {
        node_id: 'host-192-168-1-10',
        label: 'Web Server DMZ (192.168.1.10)',
        mitigated_paths: 1,
        mitigated_risk: 86.4,
        remediation:
          'Aggiornare Apache HTTP Server alla versione più recente e isolare le rotte verso 10.0.0.50.',
      },
    ],
    summary: {
      total_nodes: 5,
      total_edges: 4,
      total_paths: 1,
      critical_paths: 1,
      highest_risk_score: 86.4,
      top_chokepoint: 'Web Server DMZ (192.168.1.10)',
    },
  };

  beforeEach(() => {
    vi.clearAllMocks();
    companyState.selectedCompanyId = null;
    companyState.companyTargetValues = new Set();
    companyState.departments = [];
    companyState.allTargets = [];
  });

  it('rende correttamente i dati del grafo, i KPI e le card di riepilogo', async () => {
    vi.mocked(attackPathApi.getGraph).mockResolvedValue(mockGraphData);

    render(
      <MemoryRouter>
        <AttackPath />
      </MemoryRouter>
    );

    // Verifica stato di caricamento e attesa dati
    await waitFor(() => {
      expect(screen.getByText('Attack Path Mapping & Lateral Movement')).toBeDefined();
    });

    // Verifica KPI
    expect(screen.getByText('Nodi & Connessioni')).toBeDefined();
    expect(screen.getByText('5')).toBeDefined();
    expect(screen.getByText('(4 archi)')).toBeDefined();
    expect(screen.getByText('1 Critici')).toBeDefined();
    expect(screen.getAllByText('86.4').length).toBeGreaterThan(0);
    expect(screen.getAllByText('Web Server DMZ (192.168.1.10)').length).toBeGreaterThan(0);

    // Verifica presenza tab Percorsi
    expect(screen.getByText(/Percorsi \(1\)/i)).toBeDefined();
    expect(
      screen.getByText('Compromissione DMZ verso Database Interno')
    ).toBeDefined();
  });

  it('permette la navigazione nei tab Chokepoints e la visualizzazione della remediation prescrittiva', async () => {
    vi.mocked(attackPathApi.getGraph).mockResolvedValue(mockGraphData);

    render(
      <MemoryRouter>
        <AttackPath />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText(/Chokepoints \(1\)/i)).toBeDefined();
    });

    // Click sul tab Chokepoints
    fireEvent.click(screen.getByText(/Chokepoints \(1\)/i));

    // Verifica presenza remediation e statistiche di mitigazione
    expect(
      screen.getByText(/Aggiornare Apache HTTP Server alla versione più recente/i)
    ).toBeDefined();
    expect(screen.getByText(/1 percorsi/i)).toBeDefined();
    expect(screen.getByText(/86.4 pts/i)).toBeDefined();
  });

  it('supporta il click sul pulsante Focalizza per ispezionare il nodo selezionato', async () => {
    vi.mocked(attackPathApi.getGraph).mockResolvedValue(mockGraphData);

    render(
      <MemoryRouter>
        <AttackPath />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText(/Chokepoints \(1\)/i)).toBeDefined();
    });

    fireEvent.click(screen.getByText(/Chokepoints \(1\)/i));

    const focusBtn = screen.getByTitle('Focalizza sul Grafo');
    fireEvent.click(focusBtn);

    // Verifica apertura del tab Dettaglio Nodo
    expect(screen.getByText('Dettaglio Nodo')).toBeDefined();
    expect(screen.getByText('Indirizzo IP:')).toBeDefined();
    expect(screen.getAllByText('192.168.1.10').length).toBeGreaterThan(0);
    expect(screen.getByText('Chokepoint di Massima Priorità')).toBeDefined();
  });

  it('rimuove il filtro CVSS e carica tutte le vulnerabilità senza soglia', async () => {
    vi.mocked(attackPathApi.getGraph).mockResolvedValue(mockGraphData);

    render(
      <MemoryRouter>
        <AttackPath />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByLabelText('Reparto Attack Path')).toBeDefined();
    });

    expect(screen.queryByLabelText('Filtro CVSS')).toBeNull();
    expect(attackPathApi.getGraph).toHaveBeenCalledWith();
  });

  it('mostra correttamente lo stato vuoto quando non ci sono target né vulnerabilità scoperte', async () => {
    const emptyGraph: AttackGraph = {
      tenant_id: '00000000-0000-0000-0000-000000000001',
      generated_at: '2026-10-10T12:00:00Z',
      nodes: [
        {
          id: 'entry-internet',
          type: 'ENTRYPOINT',
          label: 'Perimetro Internet',
          severity: 'LOW',
        },
      ],
      edges: [],
      paths: [],
      chokepoints: [],
      summary: {
        total_nodes: 1,
        total_edges: 0,
        total_paths: 0,
        critical_paths: 0,
        highest_risk_score: 0,
      },
    };

    vi.mocked(attackPathApi.getGraph).mockResolvedValue(emptyGraph);

    render(
      <MemoryRouter>
        <AttackPath />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(
        screen.getByText('Nessun percorso di attacco rilevato')
      ).toBeDefined();
    });

    expect(screen.getByText('Vai a Scansioni Nmap')).toBeDefined();
  });

  it('mostra un messaggio di errore chiaro in caso di fallimento della richiesta API', async () => {
    vi.mocked(attackPathApi.getGraph).mockRejectedValue(
      new Error('Timeout di connessione al database')
    );

    render(
      <MemoryRouter>
        <AttackPath />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(
        screen.getByText('Timeout di connessione al database')
      ).toBeDefined();
    });

    expect(screen.getByText('Riprova')).toBeDefined();
  });

  it('distingue le inferenze senza presentare percentuali non calibrate come prove', async () => {
    const graphWithInferred: AttackGraph = {
      ...mockGraphData,
      edges: [
        ...mockGraphData.edges,
        {
          id: 'edge-inferred-1',
          source: 'vuln-cve-2021-41773',
          target: 'host-10-0-0-50',
          type: 'PIVOTS_TO',
          label: 'Pivot Rete Interna',
          weight: 2.0,
          inferred: true,
          confidence: 0.85,
          inference_basis: 'Condivisione stessa subnet o dual-homed host',
        },
      ],
    };

    vi.mocked(attackPathApi.getGraph).mockResolvedValue(graphWithInferred);

    render(
      <MemoryRouter>
        <AttackPath />
      </MemoryRouter>
    );

    await screen.findByRole('button', { name: 'Mappa completa' });
    fireEvent.click(screen.getByRole('button', { name: 'Mappa completa' }));
    const edge = document.querySelector('[data-edge-id="edge-inferred-1"]');
    expect(edge?.getAttribute('stroke-dasharray')).toBe('7 5');
    expect(screen.getByText('Ipotesi / correlazione da verificare')).toBeDefined();
    expect(screen.queryByText('INFERRED 85%')).toBeNull();
  });

  it('non ricarica il grafo cambiando selezione e rende disponibile la guida', async () => {
    vi.mocked(attackPathApi.getGraph).mockResolvedValue(mockGraphData);
    render(<MemoryRouter><AttackPath /></MemoryRouter>);
    const path = await screen.findByRole('button', { name: /Seleziona percorso/ });
    fireEvent.click(path);
    fireEvent.click(path);
    expect(attackPathApi.getGraph).toHaveBeenCalledTimes(1);
    fireEvent.click(screen.getByText('Guida all’uso di Attack Path'));
    expect(screen.getByText('Dalla scansione alla priorità')).toBeDefined();
  });

  it('non mostra dati globali se l’azienda attiva è vuota', async () => {
    companyState.selectedCompanyId = 'empty-company';
    vi.mocked(attackPathApi.getGraph).mockResolvedValue(mockGraphData);
    render(<MemoryRouter><AttackPath /></MemoryRouter>);
    await screen.findByText('Nessun percorso di attacco rilevato');
    expect(screen.queryByText('Compromissione DMZ verso Database Interno')).toBeNull();
    expect(screen.queryByRole('group', { name: 'Grafo dei percorsi di analisi' })).toBeNull();
  });

  it('esclude la catena cross-company anche quando un suo host è nello scope', async () => {
    companyState.selectedCompanyId = 'dmz-company';
    companyState.companyTargetValues = new Set(['192.168.1.0/24']);
    vi.mocked(attackPathApi.getGraph).mockResolvedValue(mockGraphData);
    render(<MemoryRouter><AttackPath /></MemoryRouter>);
    await screen.findByText('Percorsi (0)');
    expect(screen.queryByText('Compromissione DMZ verso Database Interno')).toBeNull();
    expect(screen.queryByRole('button', { name: /Ispeziona Database Interno/ })).toBeNull();
    expect(screen.getByText('Chokepoints (0)')).toBeDefined();
  });

  it('limita il grafo al reparto e non usa fallback quando il reparto è vuoto', async () => {
    companyState.selectedCompanyId = 'company';
    companyState.companyTargetValues = new Set(['192.168.1.10', '10.0.0.50']);
    companyState.departments = [
      { id: 'it', name: 'IT', parent_id: 'company' } as TargetGroupNode,
      { id: 'empty', name: 'Vuoto', parent_id: 'company' } as TargetGroupNode,
    ];
    companyState.allTargets = [{ group_id: 'it', value: '192.168.1.10' } as TargetItem];
    vi.mocked(attackPathApi.getGraph).mockResolvedValue(mockGraphData);
    render(<MemoryRouter><AttackPath /></MemoryRouter>);
    await screen.findByText('Percorsi (1)');
    fireEvent.change(screen.getByLabelText('Reparto Attack Path'), { target: { value: 'it' } });
    expect(screen.getByText('Percorsi (0)')).toBeDefined();
    expect(screen.queryByRole('button', { name: /Ispeziona Database Interno/ })).toBeNull();
    fireEvent.change(screen.getByLabelText('Reparto Attack Path'), { target: { value: 'empty' } });
    expect(screen.getByText('Nessun percorso di attacco rilevato')).toBeDefined();
    expect(attackPathApi.getGraph).toHaveBeenCalledTimes(1);
  });
});

