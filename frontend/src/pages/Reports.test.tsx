import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { Reports } from './Reports';
import { reportsApi } from '../api/reports';
import { apiClient } from '../api/client';
import { ReportItem } from '../types/report';
import { TargetGroupNode } from '../types/target';

const scope = vi.hoisted(() => ({ company: null as string | null }));
vi.mock('../context/CompanyContext', () => ({ useCompany: () => ({
  selectedCompanyId: scope.company, selectedCompany: null, departments: [],
  companyGroupIds: new Set(scope.company ? [scope.company] : []),
}) }));

vi.mock('../api/reports', () => ({
  reportsApi: {
    list: vi.fn(),
    download: vi.fn(),
    downloadCompany: vi.fn(),
  },
}));

vi.mock('../api/client', () => ({
  apiClient: {
    get: vi.fn(),
    post: vi.fn(),
    delete: vi.fn(),
  },
}));

describe('Reports Page - Multi-Company Scope & Target IP Inspection', () => {
  const mockReports: ReportItem[] = [
    {
      scan_id: 'scan-alfa-01',
      tenant_id: 'tenant-1',
      target_id: 'target-01',
      target_value: '192.168.1.100',
      target_type: 'IP',
      group_id: 'group-alfa',
      group_name: 'Azienda Alfa S.r.l.',
      profile_id: 'profile-fast',
      profile_name: 'Fast Network Audit',
      status: 'COMPLETED',
      trigger_type: 'MANUAL',
      summary_stats: {
        total_hosts: 1,
        up_hosts: 1,
        total_open_ports: 2,
        critical_count: 1,
        high_count: 0,
        medium_count: 0,
        low_count: 0,
        info_count: 0,
        duration_seconds: 15.0,
      },
      max_severity: 'CRITICAL',
      posture_score: 45,
      posture_grade: 'D',
      created_at: '2026-10-09T10:00:00Z',
      completed_at: '2026-10-09T10:05:00Z',
    },
    {
      scan_id: 'scan-beta-01',
      tenant_id: 'tenant-1',
      target_id: 'target-02',
      target_value: '10.0.0.50',
      target_type: 'IP',
      group_id: 'group-beta',
      group_name: 'Azienda Beta S.p.A.',
      profile_id: 'profile-full',
      profile_name: 'Full Port Audit',
      status: 'COMPLETED',
      trigger_type: 'SCHEDULED',
      summary_stats: {
        total_hosts: 1,
        up_hosts: 1,
        total_open_ports: 1,
        critical_count: 0,
        high_count: 0,
        medium_count: 0,
        low_count: 0,
        info_count: 0,
        duration_seconds: 30.0,
      },
      max_severity: 'INFO',
      posture_score: 95,
      posture_grade: 'A',
      created_at: '2026-10-09T11:00:00Z',
      completed_at: '2026-10-09T11:05:00Z',
    },
  ];

  const mockGroups: TargetGroupNode[] = [
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
      aggregate_risk: 'LOW',
      severity_counts: { INFO: 1 },
      created_at: '2026-10-08T09:30:00Z',
      updated_at: '2026-10-08T09:30:00Z',
    },
  ];

  beforeEach(() => {
    vi.clearAllMocks();
    scope.company = null;

    vi.mocked(reportsApi.list).mockResolvedValue({
      items: mockReports,
      total: mockReports.length,
      limit: 100,
      offset: 0,
    });

    vi.mocked(reportsApi.download).mockResolvedValue(undefined);

    vi.mocked(apiClient.get).mockImplementation((url: string) => {
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

  it('renders company scope groups with posture metrics and specific target IPs', async () => {
    render(
      <MemoryRouter>
        <Reports />
      </MemoryRouter>
    );

    // Verifica la presenza dei gruppi aziendali
    await waitFor(() => {
      expect(screen.getAllByText('Azienda Alfa S.r.l.').length).toBeGreaterThan(0);
      expect(screen.getAllByText('Azienda Beta S.p.A.').length).toBeGreaterThan(0);
    });

    // Per Alfa, verifica il Target IP e il Posture Score
    expect(screen.getByText('192.168.1.100')).toBeDefined();
    expect(screen.getByText('45/100')).toBeDefined();
    expect(screen.getByText('GRADE D')).toBeDefined();

    // Per Beta, verifica l'IP e il Posture Score
    expect(screen.getByText('10.0.0.50')).toBeDefined();
    expect(screen.getByText('95/100')).toBeDefined();
    expect(screen.getByText('GRADE A')).toBeDefined();
  });

  it('allows collapsing and expanding company accordions', async () => {
    render(
      <MemoryRouter>
        <Reports />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('192.168.1.100')).toBeDefined();
    });

    // Clicca sull'header di Alfa per collassare l'accordion
    const alfaHeader = screen.getByRole('heading', { name: 'Azienda Alfa S.r.l.' });
    fireEvent.click(alfaHeader);

    // L'IP di Alfa non deve essere più visibile nella tabella espansa
    expect(screen.queryByText('192.168.1.100')).toBeNull();

    // L'IP di Beta rimane visibile
    expect(screen.getByText('10.0.0.50')).toBeDefined();

    // Clicca nuovamente per riaprire
    fireEvent.click(alfaHeader);
    expect(screen.getByText('192.168.1.100')).toBeDefined();
  });

  it('allows switching to flat chronological archive view', async () => {
    render(
      <MemoryRouter>
        <Reports />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getAllByText('Per Azienda').length).toBeGreaterThan(0);
    });

    // Clicca sul pulsante Cronologico
    const cronologicoBtn = screen.getByText('Cronologico');
    fireEvent.click(cronologicoBtn);

    // Deve chiamare reportsApi.list con la paginazione corretta
    await waitFor(() => {
      expect(reportsApi.list).toHaveBeenCalledWith(
        expect.objectContaining({ limit: 10, offset: 0 })
      );
    });
  });

  it('triggers 1-click download when clicking download button', async () => {
    render(
      <MemoryRouter>
        <Reports />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('192.168.1.100')).toBeDefined();
    });

    // Trova i pulsanti download per Executive PDF
    const downloadExecutiveButtons = screen.getAllByTitle(/Scarica Report PDF Executive/i);
    expect(downloadExecutiveButtons.length).toBeGreaterThan(0);

    await waitFor(async () => {
      fireEvent.click(downloadExecutiveButtons[0]);
    });

    expect(reportsApi.downloadCompany).toHaveBeenCalledWith(
      'group-alfa',
      'Azienda Alfa S.r.l.',
      'pdf',
      'executive',
      expect.any(Object)
    );
  });

  it('combines profile and inclusive date filters, with a visible invalid-period error', async () => {
    render(<Reports />);
    await screen.findByRole('heading', { name: 'Azienda Alfa S.r.l.' });
    fireEvent.change(screen.getByLabelText('Profilo'), { target: { value: 'profile-fast' } });
    expect(screen.queryByRole('heading', { name: 'Azienda Beta S.p.A.' })).toBeNull();
    fireEvent.change(screen.getByLabelText('Dal giorno (UTC)'), { target: { value: '2026-10-10' } });
    fireEvent.change(screen.getByLabelText('Al giorno (UTC)'), { target: { value: '2026-10-09' } });
    expect(screen.getByRole('alert').textContent).toContain('data iniziale');
    expect(screen.queryByRole('heading', { name: 'Azienda Alfa S.r.l.' })).toBeNull();
    fireEvent.click(screen.getByRole('button', { name: 'Azzera filtri' }));
    expect(screen.getByRole('heading', { name: 'Azienda Beta S.p.A.' })).toBeDefined();
  });

  it('applies company scope to the chronological archive', async () => {
    scope.company = 'group-alfa';
    render(<Reports />);
    await screen.findByText('192.168.1.100');
    fireEvent.click(screen.getByText('Cronologico'));
    await waitFor(() => expect(screen.getByText('192.168.1.100')).toBeDefined());
    expect(screen.queryByText('10.0.0.50')).toBeNull();
  });

  it('loads all API pages and makes chronological pages navigable', async () => {
    const rows = Array.from({ length: 12 }, (_, index) => ({ ...mockReports[0], scan_id: `scan-${index}`, target_id: `target-${index}`, target_value: `192.0.2.${index + 1}` }));
    vi.mocked(reportsApi.list).mockImplementation(async params => ({ items: rows.slice(params?.offset || 0, (params?.offset || 0) + (params?.limit || 100)), total: rows.length, limit: params?.limit || 100, offset: params?.offset || 0 }));
    render(<Reports />);
    await screen.findByText('192.0.2.12');
    fireEvent.click(screen.getByText('Cronologico'));
    await screen.findByRole('button', { name: 'Pagina successiva' });
    expect(screen.queryByText('192.0.2.12')).toBeNull();
    fireEvent.click(screen.getByRole('button', { name: 'Pagina successiva' }));
    expect(screen.getByText('192.0.2.12')).toBeDefined();
    expect(reportsApi.list).toHaveBeenCalledWith({ limit: 10, offset: 10 });
  });
});
