import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { Dashboard } from './Dashboard';
import { apiClient } from '../api/client';
import { TargetGroupNode, TargetItem } from '../types/target';
import { ReportItem } from '../types/report';
import { ScanItem } from '../types/scan';

vi.mock('../api/client', () => ({
  apiClient: {
    get: vi.fn(),
    post: vi.fn(),
    delete: vi.fn(),
  },
}));

vi.mock('../context/AuthContext', () => ({
  useAuth: () => ({
    user: { id: 'u1', email: 'admin@secops.local', tenant_id: 'tenant-1', role: 'ADMIN' },
  }),
}));

describe('Dashboard Page - Per-Company Posture Overview & Workflow Stepper', () => {
  const mockGroups: TargetGroupNode[] = [
    {
      id: 'grp-acme',
      tenant_id: 'tenant-1',
      name: 'Acme Industries Corp',
      description: 'Sede centrale e server farm',
      color_tag: '#06b6d4',
      children: [],
      direct_target_count: 2,
      target_count: 2,
      aggregate_risk: 'CRITICAL',
      severity_counts: { CRITICAL: 1, HIGH: 2 },
      created_at: '2026-10-09T10:00:00Z',
      updated_at: '2026-10-09T10:00:00Z',
    },
    {
      id: 'grp-globex',
      tenant_id: 'tenant-1',
      name: 'Globex Logistics',
      description: 'Filiale logistica e magazzini',
      color_tag: '#8b5cf6',
      children: [],
      direct_target_count: 1,
      target_count: 1,
      aggregate_risk: 'LOW',
      severity_counts: { LOW: 1 },
      created_at: '2026-10-09T10:00:00Z',
      updated_at: '2026-10-09T10:00:00Z',
    },
  ];

  const mockTargets: TargetItem[] = [
    {
      id: 'tgt-1',
      tenant_id: 'tenant-1',
      group_id: 'grp-acme',
      type: 'IP',
      value: '192.168.10.10',
      is_blacklisted: false,
      created_at: '2026-10-09T10:00:00Z',
      updated_at: '2026-10-09T10:00:00Z',
    },
    {
      id: 'tgt-2',
      tenant_id: 'tenant-1',
      group_id: 'grp-acme',
      type: 'IP',
      value: '192.168.10.20',
      is_blacklisted: false,
      created_at: '2026-10-09T10:00:00Z',
      updated_at: '2026-10-09T10:00:00Z',
    },
    {
      id: 'tgt-3',
      tenant_id: 'tenant-1',
      group_id: 'grp-globex',
      type: 'IP',
      value: '10.20.30.40',
      is_blacklisted: false,
      created_at: '2026-10-09T10:00:00Z',
      updated_at: '2026-10-09T10:00:00Z',
    },
    {
      id: 'tgt-unassigned',
      tenant_id: 'tenant-1',
      type: 'IP',
      value: '172.16.0.100',
      is_blacklisted: false,
      created_at: '2026-10-09T10:00:00Z',
      updated_at: '2026-10-09T10:00:00Z',
    },
  ];

  const mockReports: ReportItem[] = [
    {
      scan_id: 'scan-acme-1',
      tenant_id: 'tenant-1',
      target_id: 'tgt-1',
      target_value: '192.168.10.10',
      target_type: 'IP',
      group_id: 'grp-acme',
      group_name: 'Acme Industries Corp',
      profile_id: 'p1',
      profile_name: 'All-in-One 360',
      status: 'COMPLETED',
      trigger_type: 'MANUAL',
      summary_stats: {
        total_hosts: 1,
        up_hosts: 1,
        total_open_ports: 3,
        critical_count: 1,
        high_count: 2,
        medium_count: 0,
        low_count: 0,
        info_count: 0,
        duration_seconds: 30,
      },
      max_severity: 'CRITICAL',
      posture_score: 45,
      posture_grade: 'D',
      created_at: '2026-10-09T10:00:00Z',
      completed_at: '2026-10-09T10:15:00Z',
    },
  ];

  const mockScans: ScanItem[] = [
    {
      id: 'scan-acme-1',
      tenant_id: 'tenant-1',
      target_id: 'tgt-1',
      profile_id: 'p1',
      trigger_type: 'MANUAL',
      status: 'COMPLETED',
      summary_stats: {
        total_hosts: 1,
        up_hosts: 1,
        total_open_ports: 3,
        critical_count: 1,
        high_count: 2,
        medium_count: 0,
        low_count: 0,
        info_count: 0,
        duration_seconds: 30,
      },
      created_at: '2026-10-09T10:00:00Z',
      completed_at: '2026-10-09T10:15:00Z',
      updated_at: '2026-10-09T10:15:00Z',
    },
  ];

  beforeEach(() => {
    vi.clearAllMocks();

    vi.mocked(apiClient.get).mockImplementation((url: string) => {
      if (url.startsWith('/target-groups')) {
        return Promise.resolve({ data: { success: true, data: mockGroups } });
      }
      if (url.startsWith('/targets')) {
        return Promise.resolve({ data: { success: true, data: { items: mockTargets, total: mockTargets.length } } });
      }
      if (url.startsWith('/reports')) {
        return Promise.resolve({ data: { success: true, data: { items: mockReports, total: mockReports.length } } });
      }
      if (url.startsWith('/scans')) {
        return Promise.resolve({ data: { success: true, data: { items: mockScans, total: mockScans.length } } });
      }
      return Promise.reject(new Error(`Unhandled URL: ${url}`));
    });
  });

  it('renders linear workflow stepper and company posture cards', async () => {
    render(
      <MemoryRouter>
        <Dashboard />
      </MemoryRouter>
    );

    // Verifica la presenza degli step operativi lineari
    expect(screen.getByText('Flusso Operativo Lineare della Piattaforma')).toBeDefined();
    expect(screen.getByText('Network Discovery')).toBeDefined();
    expect(screen.getByText('Target & Aziende')).toBeDefined();
    expect(screen.getByText('Scansioni Nmap')).toBeDefined();
    expect(screen.getByText('Vulnerabilità CVE')).toBeDefined();
    expect(screen.getByText('Report Aziendali')).toBeDefined();

    // Attendi caricamento dati aziende
    await waitFor(() => {
      expect(screen.getByText('Acme Industries Corp')).toBeDefined();
      expect(screen.getByText('Globex Logistics')).toBeDefined();
    });

    // Verifica card Acme
    expect(screen.getByText('192.168.10.10')).toBeDefined();
    expect(screen.getByText('192.168.10.20')).toBeDefined();
    expect(screen.getByText('GRADE D')).toBeDefined();
    expect(screen.getByText('45/100')).toBeDefined();

    // Verifica card Globex
    expect(screen.getByText('10.20.30.40')).toBeDefined();
    expect(screen.getByText('GRADE A')).toBeDefined();

    // Verifica callout bersagli non assegnati
    expect(screen.getByText(/1 bersagli non sono ancora assegnati/i)).toBeDefined();
  });

  it('filters companies by search query', async () => {
    render(
      <MemoryRouter>
        <Dashboard />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('Acme Industries Corp')).toBeDefined();
      expect(screen.getByText('Globex Logistics')).toBeDefined();
    });

    // Inserisci "Globex" nel campo di ricerca
    const searchInput = screen.getByPlaceholderText(/Filtra aziende o IP perimetrali/i);
    fireEvent.change(searchInput, { target: { value: 'Globex' } });

    // Acme scompare, Globex rimane
    expect(screen.queryByText('Acme Industries Corp')).toBeNull();
    expect(screen.getByText('Globex Logistics')).toBeDefined();
  });
});

