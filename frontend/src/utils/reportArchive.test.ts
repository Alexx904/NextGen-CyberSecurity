import { describe, expect, it, vi } from 'vitest';
import { filterReportArchive, loadReportArchive, type ArchiveFilters } from './reportArchive';
import type { ReportItem } from '../types/report';

const row: ReportItem = {
  scan_id: 'scan', tenant_id: 'tenant', target_id: 'target', target_value: '192.0.2.1', target_type: 'IP',
  group_id: 'group', group_name: 'Client', profile_id: 'profile', profile_name: 'Audit', status: 'COMPLETED', trigger_type: 'MANUAL',
  summary_stats: { total_hosts: 1, up_hosts: 1, total_open_ports: 0, critical_count: 0, high_count: 0, medium_count: 0, low_count: 0, info_count: 0, duration_seconds: 1 },
  max_severity: 'HIGH', posture_score: 80, posture_grade: 'B', created_at: '2026-10-01T00:00:00Z', completed_at: '2026-10-18T23:59:59.999Z',
};
const empty: ArchiveFilters = { dateFrom: '', dateTo: '', profile: '', target: '', group: '', severity: '', search: '' };

describe('report archive', () => {
  it('loads later pages before returning the archive', async () => {
    const list = vi.fn().mockResolvedValueOnce({ items: [row], total: 2 }).mockResolvedValueOnce({ items: [{ ...row, scan_id: 'later' }], total: 2 });
    expect(await loadReportArchive(list, 100, () => true)).toHaveLength(2);
    expect(list).toHaveBeenNthCalledWith(2, { limit: 100, offset: 1 });
  });
  it('rejects incomplete, duplicate and changing archives', async () => {
    for (const page of [{ items: [], total: 2 }, { items: [row], total: 2 }, { items: [{ ...row, scan_id: 'later' }], total: 3 }]) {
      const list = vi.fn().mockResolvedValueOnce({ items: [row], total: 2 }).mockResolvedValueOnce(page);
      await expect(loadReportArchive(list, 100, () => true)).rejects.toThrow();
    }
  });
  it('does not continue a superseded request', async () => {
    const list = vi.fn().mockResolvedValue({ items: [row], total: 2 });
    expect(await loadReportArchive(list, 100, () => false)).toEqual([]);
    expect(list).toHaveBeenCalledTimes(1);
  });
  it('combines inclusive UTC dates and exact identity/severity filters', () => {
    const filters = { dateFrom: '2026-10-18', dateTo: '2026-10-18', profile: 'profile', target: 'target', group: 'group', severity: 'HIGH', search: 'AUDIT' };
    expect(filterReportArchive([row], filters)).toEqual([row]);
    for (const field of ['profile', 'target', 'group', 'severity'] as const) {
      expect(filterReportArchive([row], { ...filters, [field]: 'other' })).toEqual([]);
    }
    expect(filterReportArchive([row], { ...empty, dateTo: '2026-10-17' })).toEqual([]);
    expect(filterReportArchive([row], { ...empty, dateFrom: '2026-10-19', dateTo: '2026-10-18' })).toEqual([]);
    expect(filterReportArchive([{ ...row, completed_at: undefined }], { ...empty, dateTo: '2026-10-01' })).toHaveLength(1);
  });
});
