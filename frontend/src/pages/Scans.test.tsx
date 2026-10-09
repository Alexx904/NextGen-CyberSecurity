import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { Scans } from './Scans';
import { apiClient } from '../api/client';

vi.mock('../api/client', () => ({
  apiClient: {
    get: vi.fn(),
    post: vi.fn(),
    delete: vi.fn(),
  },
}));

describe('Scans Page - Workstation Multi-Company Grouped View', () => {
  const mockScans = [
    {
      id: 'scan-001-alfa-web',
      tenant_id: 'tenant-1',
      target_id: 'target-01',
      profile_id: 'profile-fast',
      trigger_type: 'MANUAL',
      status: 'COMPLETED',
      created_at: '2026-10-09T10:00:00Z',
      updated_at: '2026-10-09T10:05:00Z',
      summary_stats: {
        total_hosts: 1,
        up_hosts: 1,
        total_open_ports: 3,
        critical_count: 2,
        high_count: 1,
        medium_count: 0,
        low_count: 0,
        info_count: 0,
        duration_seconds: 14.5,
      },
    },
    {
      id: 'scan-002-beta-db',
      tenant_id: 'tenant-1',
      target_id: 'target-02',
      profile_id: 'profile-full',
      trigger_type: 'SCHEDULED',
      status: 'RUNNING',
      created_at: '2026-10-09T11:00:00Z',
      updated_at: '2026-10-09T11:02:00Z',
      summary_stats: {
        total_hosts: 1,
        up_hosts: 1,
        total_open_ports: 2,
        critical_count: 0,
        high_count: 3,
        medium_count: 1,
        low_count: 0,
        info_count: 0,
        duration_seconds: 45.0,
      },
    },
  ];

  const mockTargets = [
    {
      id: 'target-01',
      tenant_id: 'tenant-1',
      group_id: 'group-alfa',
      type: 'IP',
      value: '192.168.1.100',
      description: 'Web Server Alfa',
      is_blacklisted: false,
      created_at: '2026-10-08T10:00:00Z',
      updated_at: '2026-10-08T10:00:00Z',
    },
    {
      id: 'target-02',
      tenant_id: 'tenant-1',
      group_id: 'group-beta',
      type: 'IP',
      value: '10.0.0.50',
      description: 'Database Host Beta',
      is_blacklisted: false,
      created_at: '2026-10-08T11:00:00Z',
      updated_at: '2026-10-08T11:00:00Z',
    },
  ];

  const mockGroups = [
    {
      id: 'group-alfa',
      tenant_id: 'tenant-1',
      name: 'Azienda Alfa S.r.l.',
      color_tag: '#06b6d4',
      children: [],
      target_count: 1,
      direct_target_count: 1,
      aggregate_risk: 'CRITICAL',
      severity_counts: { CRITICAL: 2, HIGH: 1 },
      created_at: '2026-10-08T09:00:00Z',
      updated_at: '2026-10-08T09:00:00Z',
    },
    {
      id: 'group-beta',
      tenant_id: 'tenant-1',
      name: 'Azienda Beta S.p.A.',
      color_tag: '#8b5cf6',
      children: [],
      target_count: 1,
      direct_target_count: 1,
      aggregate_risk: 'HIGH',
      severity_counts: { HIGH: 3 },
      created_at: '2026-10-08T09:30:00Z',
      updated_at: '2026-10-08T09:30:00Z',
    },
  ];

  beforeEach(() => {
    vi.clearAllMocks();

    vi.mocked(apiClient.get).mockImplementation((url: string) => {
      if (url.startsWith('/scans')) {
        return Promise.resolve({
          data: {
            success: true,
            data: {
              items: mockScans,
              total: mockScans.length,
              limit: 50,
              offset: 0,
            },
          },
        });
      }
      if (url.startsWith('/targets')) {
        return Promise.resolve({
          data: {
            success: true,
            data: {
              items: mockTargets,
              total: mockTargets.length,
              limit: 200,
              offset: 0,
            },
          },
        });
      }
      if (url.startsWith('/target-groups')) {
        return Promise.resolve({
          data: {
            success: true,
            data: mockGroups,
          },
        });
      }
      if (url.startsWith('/schedules')) {
        return Promise.resolve({
          data: {
            success: true,
            data: {
              items: [],
              total: 0,
            },
          },
        });
      }
      return Promise.reject(new Error(`Unhandled URL: ${url}`));
    });
  });

  it('renders company groups and specific target IPs in grouped view by default', async () => {
    render(
      <MemoryRouter>
        <Scans />
      </MemoryRouter>
    );

    // Aspetta il caricamento dei dati
    await waitFor(() => {
      expect(screen.getByText('Azienda Alfa S.r.l.')).toBeDefined();
      expect(screen.getByText('Azienda Beta S.p.A.')).toBeDefined();
    });

    // Verifica che gli IP dei target specifici siano renderizzati sotto le rispettive aziende
    expect(screen.getByText('192.168.1.100')).toBeDefined();
    expect(screen.getByText('10.0.0.50')).toBeDefined();

    // Verifica la presenza delle scansioni ID parziali
    expect(screen.getByText('scan-001...')).toBeDefined();
    expect(screen.getByText('scan-002...')).toBeDefined();

    // Verifica badge metriche per azienda
    expect(screen.getByText('2 Crit')).toBeDefined();
    expect(screen.getByText('3 High')).toBeDefined();
  });

  it('allows collapsing and expanding groups and controls', async () => {
    render(
      <MemoryRouter>
        <Scans />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('Azienda Alfa S.r.l.')).toBeDefined();
    });

    // Clicca per comprimere tutti i gruppi
    const collapseAllBtn = screen.getByText('Comprimi tutti');
    fireEvent.click(collapseAllBtn);

    // Quando compresso, l'IP interno non è presente nel DOM
    expect(screen.queryByText('192.168.1.100')).toBeNull();

    // Clicca "Espandi tutti"
    const expandAllBtn = screen.getByText('Espandi tutti');
    fireEvent.click(expandAllBtn);

    expect(screen.getByText('192.168.1.100')).toBeDefined();
    expect(screen.getByText('10.0.0.50')).toBeDefined();
  });

  it('switches between grouped view and linear table view seamlessly', async () => {
    render(
      <MemoryRouter>
        <Scans />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('Azienda Alfa S.r.l.')).toBeDefined();
    });

    // Clicca sul pulsante "Tabella Lineare"
    const tableBtn = screen.getByText('Tabella Lineare');
    fireEvent.click(tableBtn);

    // Nella tabella lineare classica, l'header delle colonne è presente
    expect(screen.getByText('Scansione ID')).toBeDefined();
    expect(screen.getByText('Data Creazione')).toBeDefined();

    // Torna alla vista per cartella
    const groupedBtn = screen.getByText('Vista per Cartella/Azienda');
    fireEvent.click(groupedBtn);

    expect(screen.getByText('Azienda Alfa S.r.l.')).toBeDefined();
  });

  it('filters scans and targets by search filter input', async () => {
    render(
      <MemoryRouter>
        <Scans />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('192.168.1.100')).toBeDefined();
      expect(screen.getByText('10.0.0.50')).toBeDefined();
    });

    const searchInput = screen.getByPlaceholderText('Cerca per IP, host, cartella scope o ID scansione...');
    fireEvent.change(searchInput, { target: { value: '192.168.1.100' } });

    // Mostra solo Alfa
    expect(screen.getByText('192.168.1.100')).toBeDefined();
    expect(screen.queryByText('10.0.0.50')).toBeNull();
  });
});
