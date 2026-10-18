import { expect, it } from 'vitest';
import { discoveryCSV, discoveryJSON } from './discoveryExport';
import type { DiscoveryResult } from '../types/discovery';

const result: DiscoveryResult = { id: 'session', tenant_id: 'tenant', subnet: '192.0.2.0/30', devices: [{ ip: '192.0.2.1', status: 'up', hostname: ' =HYPERLINK("bad")', vendor: 'Vendor, Inc.', latency_ms: 0.2, open_ports: [22, 80], scanned_at: '2026-10-18T00:00:00Z' }], total_discovered: 1, duration_seconds: 0, started_at: '', completed_at: '' };

it('quotes CSV text and neutralizes a formula after leading whitespace', () => {
  const csv = discoveryCSV(result.devices);
  expect(csv).toContain('"\' =HYPERLINK(""bad"")"');
  expect(csv).toContain('"Vendor, Inc."');
  expect(csv).toContain('"22;80"');
});

it('keeps partial snapshot metadata and exports only the explicitly selected set', () => {
  const data = JSON.parse(discoveryJSON(result, []));
  expect(data.snapshot_partial).toBe(true);
  expect(data.total_discovered).toBe(1);
  expect(data.exported_count).toBe(0);
  expect(data.devices).toEqual([]);
});
