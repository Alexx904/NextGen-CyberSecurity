import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { Findings } from './Findings';
import { apiClient } from '../api/client';

vi.mock('../api/client', () => ({
  apiClient: {
    get: vi.fn(),
    post: vi.fn(),
    delete: vi.fn(),
  },
}));

describe('Findings Page - Multi-Company Scope & IP Inspection', () => {
  const mockScans = [
    {
      id: 'scan-alfa-01',
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
        total_open_ports: 2,
        critical_count: 1,
        high_count: 0,
        medium_count: 0,
        low_count: 0,
        info_count: 0,
        duration_seconds: 12.0,
      },
    },
    {
      id: 'scan-beta-01',
      tenant_id: 'tenant-1',
      target_id: 'target-02',
      profile_id: 'profile-full',
      trigger_type: 'SCHEDULED',
      status: 'COMPLETED',
      created_at: '2026-10-09T11:00:00Z',
      updated_at: '2026-10-09T11:05:00Z',
      summary_stats: {
        total_hosts: 1,
        up_hosts: 1,
        total_open_ports: 1,
        critical_count: 0,
        high_count: 1,
        medium_count: 0,
        low_count: 0,
        info_count: 0,
        duration_seconds: 30.0,
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
      severity_counts: { CRITICAL: 1 },
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
      severity_counts: { HIGH: 1 },
      created_at: '2026-10-08T09:30:00Z',
      updated_at: '2026-10-08T09:30:00Z',
    },
  ];

  const mockFindingsAlfa = [
    {
      id: 'finding-01',
      scan_id: 'scan-alfa-01',
      tenant_id: 'tenant-1',
      host_ip: '192.168.1.100',
      hostname: 'srv-web.alfa.local',
      port: 80,
      protocol: 'tcp',
      service_name: 'http',
      service_version: 'Apache 2.4.49',
      cve_id: 'CVE-2021-41773',
      cvss_score: 9.8,
      severity: 'CRITICAL',
      raw_banner: 'Apache/2.4.49 (Unix)',
      remediation: 'Aggiorna ad Apache 2.4.51 o superiore.',
      created_at: '2026-10-09T10:05:00Z',
    },
  ];

  const mockFindingsBeta = [
    {
      id: 'finding-02',
      scan_id: 'scan-beta-01',
      tenant_id: 'tenant-1',
      host_ip: '10.0.0.50',
      hostname: 'db.beta.internal',
      port: 3306,
      protocol: 'tcp',
      service_name: 'mysql',
      service_version: 'MySQL 5.7.20',
      cve_id: 'CVE-2018-2767',
      cvss_score: 7.5,
      severity: 'HIGH',
      raw_banner: 'MySQL 5.7.20 Community Server',
      remediation: 'Aggiorna alla versione supportata di MySQL.',
      created_at: '2026-10-09T11:05:00Z',
    },
  ];

  beforeEach(() => {
    vi.clearAllMocks();

    vi.mocked(apiClient.get).mockImplementation((url: string) => {
      if (url.includes('/scans/scan-alfa-01/findings')) {
        return Promise.resolve({
          data: {
            success: true,
            data: { items: mockFindingsAlfa, total: 1, limit: 100, offset: 0 },
          },
        });
      }
      if (url.includes('/scans/scan-beta-01/findings')) {
        return Promise.resolve({
          data: {
            success: true,
            data: { items: mockFindingsBeta, total: 1, limit: 100, offset: 0 },
          },
        });
      }
      if (url.startsWith('/scans')) {
        return Promise.resolve({
          data: {
            success: true,
            data: { items: mockScans, total: mockScans.length, limit: 100, offset: 0 },
          },
        });
      }
      if (url.startsWith('/targets')) {
        return Promise.resolve({
          data: {
            success: true,
            data: { items: mockTargets, total: mockTargets.length, limit: 200, offset: 0 },
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
      return Promise.reject(new Error(`Unhandled URL: ${url}`));
    });
  });

  it('renders company cards and displays findings for the active company and target IP', async () => {
    render(
      <MemoryRouter>
        <Findings />
      </MemoryRouter>
    );

    // Verifica la presenza delle aziende nelle card
    await waitFor(() => {
      expect(screen.getAllByText('Azienda Alfa S.r.l.').length).toBeGreaterThan(0);
      expect(screen.getAllByText('Azienda Beta S.p.A.').length).toBeGreaterThan(0);
    });

    // Per Alfa (selezionata di default), l'IP 192.168.1.100 è visualizzato
    expect(screen.getByText('192.168.1.100')).toBeDefined();

    // Verifica il finding di Alfa
    await waitFor(() => {
      expect(screen.getByText('CVE-2021-41773')).toBeDefined();
    });
  });

  it('allows switching companies to inspect a different client scope and findings', async () => {
    render(
      <MemoryRouter>
        <Findings />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getAllByText('Azienda Beta S.p.A.').length).toBeGreaterThan(0);
    });

    // Clicca sulla card di Azienda Beta
    const betaCards = screen.getAllByText('Azienda Beta S.p.A.');
    fireEvent.click(betaCards[0]);

    // Ora deve comparire l'IP di Beta e il suo finding
    await waitFor(() => {
      expect(screen.getAllByText('10.0.0.50').length).toBeGreaterThan(0);
      expect(screen.getByText('CVE-2018-2767')).toBeDefined();
    });

    // Il finding di Alfa non deve essere più presente nella tabella
    expect(screen.queryByText('CVE-2021-41773')).toBeNull();
  });

  it('allows selecting a specific IP pill to isolate findings for that target only', async () => {
    render(
      <MemoryRouter>
        <Findings />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getAllByText('Azienda Alfa S.r.l.').length).toBeGreaterThan(0);
    });

    // Clicca sul pill IP 192.168.1.100
    const ipPills = screen.getAllByText('192.168.1.100');
    fireEvent.click(ipPills[0]);

    // Verifica che mostri la sezione di dettaglio target
    await waitFor(() => {
      expect(screen.getByText('Target IP selezionato:')).toBeDefined();
      expect(screen.getByText('CVE-2021-41773')).toBeDefined();
    });
  });
});
