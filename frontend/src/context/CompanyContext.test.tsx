import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import React from 'react';
import {
  CompanyProvider,
  useCompany,
  getAllDescendantGroupIds,
  flattenDepartments,
} from './CompanyContext';
import { apiClient } from '../api/client';
import { TargetGroupNode, TargetItem } from '../types/target';
import { ScanItem } from '../types/scan';
import { FindingItem } from '../types/finding';

vi.mock('../api/client', () => ({
  apiClient: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
}));

const mockTree: TargetGroupNode[] = [
  {
    id: 'comp-1',
    name: 'Acme Security Corp',
    tenant_id: 'tenant-1',
    parent_id: null,
    color_tag: '#3b82f6',
    direct_target_count: 1,
    target_count: 3,
    aggregate_risk: 'HIGH',
    severity_counts: { HIGH: 2 },
    created_at: '2026-10-10T10:00:00Z',
    updated_at: '2026-10-10T10:00:00Z',
    children: [
      {
        id: 'dept-1',
        name: 'Reparto IT',
        tenant_id: 'tenant-1',
        parent_id: 'comp-1',
        color_tag: '#10b981',
        direct_target_count: 2,
        target_count: 2,
        aggregate_risk: 'MEDIUM',
        severity_counts: { MEDIUM: 1 },
        created_at: '2026-10-10T10:00:00Z',
        updated_at: '2026-10-10T10:00:00Z',
        children: [],
      },
    ],
  },
  {
    id: 'comp-2',
    name: 'Beta Finance',
    tenant_id: 'tenant-1',
    parent_id: null,
    color_tag: '#ef4444',
    direct_target_count: 1,
    target_count: 1,
    aggregate_risk: 'LOW',
    severity_counts: { LOW: 1 },
    created_at: '2026-10-10T10:00:00Z',
    updated_at: '2026-10-10T10:00:00Z',
    children: [],
  },
];

const mockTargets: TargetItem[] = [
  {
    id: 'target-1',
    tenant_id: 'tenant-1',
    group_id: 'comp-1',
    type: 'IP',
    value: '192.168.1.10',
    is_blacklisted: false,
    created_at: '2026-10-10T10:00:00Z',
    updated_at: '2026-10-10T10:00:00Z',
  },
  {
    id: 'target-2',
    tenant_id: 'tenant-1',
    group_id: 'dept-1',
    type: 'IP',
    value: '192.168.1.20',
    is_blacklisted: false,
    created_at: '2026-10-10T10:00:00Z',
    updated_at: '2026-10-10T10:00:00Z',
  },
  {
    id: 'target-3',
    tenant_id: 'tenant-1',
    group_id: 'comp-2',
    type: 'IP',
    value: '10.0.0.5',
    is_blacklisted: false,
    created_at: '2026-10-10T10:00:00Z',
    updated_at: '2026-10-10T10:00:00Z',
  },
];

describe('CompanyContext & Helpers', () => {
  beforeEach(() => {
    localStorage.clear();
    vi.clearAllMocks();

    (apiClient.get as any).mockImplementation((url: string) => {
      if (url === '/target-groups') {
        return Promise.resolve({ data: { success: true, data: mockTree } });
      }
      if (url.startsWith('/targets')) {
        return Promise.resolve({ data: { success: true, data: { items: mockTargets } } });
      }
      return Promise.reject(new Error('not found'));
    });
  });

  it('getAllDescendantGroupIds returns company and all department ids', () => {
    const ids = getAllDescendantGroupIds(mockTree[0]);
    expect(ids).toEqual(['comp-1', 'dept-1']);
  });

  it('flattenDepartments returns child departments', () => {
    const depts = flattenDepartments(mockTree[0]);
    expect(depts).toHaveLength(1);
    expect(depts[0].name).toBe('Reparto IT');
  });

  it('initializes with all companies and filters correctly', async () => {
    const wrapper = ({ children }: { children: React.ReactNode }) => (
      <CompanyProvider>{children}</CompanyProvider>
    );

    const { result } = renderHook(() => useCompany(), { wrapper });

    // Wait for initial fetch
    await act(async () => {
      await new Promise((r) => setTimeout(r, 10));
    });

    expect(result.current.companies).toHaveLength(2);
    expect(result.current.selectedCompanyId).toBeNull();
    expect(result.current.selectedCompany).toBeNull();

    // In global mode, all targets are returned
    expect(result.current.filterTargets(mockTargets)).toHaveLength(3);

    // Select company 1
    act(() => {
      result.current.setSelectedCompanyId('comp-1');
    });

    expect(result.current.selectedCompanyId).toBe('comp-1');
    expect(result.current.selectedCompany?.name).toBe('Acme Security Corp');
    expect(result.current.departments).toHaveLength(1);
    expect(result.current.departments[0].name).toBe('Reparto IT');

    // Targets filter should return only targets of comp-1 and dept-1
    const filteredTargets = result.current.filterTargets(mockTargets);
    expect(filteredTargets).toHaveLength(2);
    expect(filteredTargets.map((t) => t.id)).toEqual(['target-1', 'target-2']);

    // Scans filter
    const mockScans: ScanItem[] = [
      {
        id: 'scan-1',
        tenant_id: 'tenant-1',
        target_id: 'target-1',
        profile_id: 'p1',
        trigger_type: 'MANUAL',
        status: 'COMPLETED',
        summary_stats: { total_hosts: 1, up_hosts: 1, total_open_ports: 2, critical_count: 0, high_count: 1, medium_count: 0, low_count: 0, info_count: 0, duration_seconds: 10 },
        created_at: '2026-10-10T10:00:00Z',
        updated_at: '2026-10-10T10:00:00Z',
      },
      {
        id: 'scan-2',
        tenant_id: 'tenant-1',
        target_id: 'target-3', // Beta Finance
        profile_id: 'p1',
        trigger_type: 'MANUAL',
        status: 'COMPLETED',
        summary_stats: { total_hosts: 1, up_hosts: 1, total_open_ports: 1, critical_count: 0, high_count: 0, medium_count: 0, low_count: 1, info_count: 0, duration_seconds: 5 },
        created_at: '2026-10-10T10:00:00Z',
        updated_at: '2026-10-10T10:00:00Z',
      },
    ];

    const filteredScans = result.current.filterScans(mockScans);
    expect(filteredScans).toHaveLength(1);
    expect(filteredScans[0].id).toBe('scan-1');

    // Findings filter
    const mockFindings: FindingItem[] = [
      {
        id: 'f-1',
        scan_id: 'scan-1',
        tenant_id: 'tenant-1',
        host_ip: '192.168.1.10',
        port: 80,
        protocol: 'tcp',
        severity: 'HIGH',
        created_at: '2026-10-10T10:00:00Z',
      },
      {
        id: 'f-2',
        scan_id: 'scan-2',
        tenant_id: 'tenant-1',
        host_ip: '10.0.0.5', // Beta Finance
        port: 443,
        protocol: 'tcp',
        severity: 'LOW',
        created_at: '2026-10-10T10:00:00Z',
      },
    ];

    const filteredFindings = result.current.filterFindings(mockFindings);
    expect(filteredFindings).toHaveLength(1);
    expect(filteredFindings[0].id).toBe('f-1');

    // Scope check
    expect(result.current.isTargetInActiveScope('192.168.1.10')).toBe(true);
    expect(result.current.isTargetInActiveScope('10.0.0.5')).toBe(false);

    // Reset to global
    act(() => {
      result.current.setSelectedCompanyId(null);
    });

    expect(result.current.selectedCompanyId).toBeNull();
    expect(result.current.filterTargets(mockTargets)).toHaveLength(3);
  });
});

