import { act, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { ScanItem } from '../../types/scan';
import { ScanBatchProgress, summarizeBatch } from './ScanBatchProgress';
import { apiClient } from '../../api/client';

vi.mock('../../api/client', () => ({ apiClient: { get: vi.fn() } }));
const now = Date.parse('2026-10-16T10:00:00Z');
const scan = (id: string, status: ScanItem['status'], percentage?: number): ScanItem => ({
  id, status, started_at: new Date(now - 60000).toISOString(), progress: percentage === undefined ? undefined : {
    scan_id: id, percentage, message: '', observed_at: new Date(now).toISOString(), remaining_seconds: 120,
  },
} as ScanItem);
afterEach(() => { vi.useRealTimers(); vi.resetAllMocks(); });

describe('Group progress and ETA', () => {
  it('uses only batch IDs, includes queued work, and counts errors as terminated not successful', () => {
    const summary = summarizeBatch(['a', 'b', 'c', 'd'], [scan('a', 'COMPLETED'), scan('b', 'RUNNING', 50), scan('c', 'QUEUED'), scan('d', 'FAILED'), scan('outside', 'COMPLETED')], now);
    expect(summary.percentage).toBe(62.5); expect(summary.completed).toBe(1); expect(summary.failed).toBe(1); expect(summary.finished).toBe(2);
    expect(summary.remaining).toBe(300); expect(summary.done).toBe(false);
  });
  it('has no ETA for queued-only, missing, stale or exhausted telemetry', () => {
    expect(summarizeBatch(['a'], [scan('a', 'QUEUED')], now).remaining).toBeUndefined();
    expect(summarizeBatch(['a', 'b'], [scan('a', 'RUNNING', 50)], now).remaining).toBeUndefined();
    expect(summarizeBatch(['a'], [scan('a', 'RUNNING', 50)], now + 31000).remaining).toBeUndefined();
    expect(summarizeBatch(['a'], [scan('a', 'RUNNING', 50)], now + 121000).remaining).toBeUndefined();
  });
  it('only reaches completed work when all members are terminal', () => {
    expect(summarizeBatch(['a'], [scan('a', 'RUNNING', 100)], now).percentage).toBe(99);
    const result = summarizeBatch(['a', 'b'], [scan('a', 'CANCELLED'), scan('b', 'FAILED')], now);
    expect(result.percentage).toBe(100); expect(result.completed).toBe(0); expect(result.done).toBe(true);
  });
  it('polls a single batch snapshot and renders total progress', async () => {
    vi.mocked(apiClient.get).mockResolvedValue({ data: { success: true, data: { items: [scan('a', 'COMPLETED'), scan('b', 'FAILED')] } } });
    render(<ScanBatchProgress ids={['a', 'b']} />);
    await screen.findByText('Lotto terminato.');
    expect(screen.getByRole('progressbar').getAttribute('aria-valuenow')).toBe('100');
    expect(apiClient.get).toHaveBeenCalledWith('/scans?ids=a,b', expect.objectContaining({ signal: expect.any(AbortSignal) }));
    expect(screen.getByText(/1 riusciti · 1 falliti/)).toBeTruthy();
  });
  it('retries failed polls without inventing an ETA and stops after unmount', async () => {
    vi.useFakeTimers();
    vi.mocked(apiClient.get).mockRejectedValue(new Error('offline'));
    const view = render(<ScanBatchProgress ids={['a']} />);
    await act(async () => { await Promise.resolve(); });
    expect(screen.getByRole('alert')).toBeTruthy();
    expect(screen.getByText('Tempo rimanente del gruppo: in calcolo')).toBeTruthy();
    await act(async () => { await vi.advanceTimersByTimeAsync(5000); });
    expect(apiClient.get).toHaveBeenCalledTimes(2);
    view.unmount(); await vi.advanceTimersByTimeAsync(5000);
    expect(apiClient.get).toHaveBeenCalledTimes(2);
  });
  it('chunks large groups and never repolls terminated jobs', async () => {
    const ids = Array.from({ length: 51 }, (_, i) => `${i}`);
    vi.mocked(apiClient.get).mockImplementation(async (url) => ({ data: { success: true, data: { items: String(url).split('ids=')[1].split(',').map((id) => scan(id, 'COMPLETED')) } } }));
    render(<ScanBatchProgress ids={ids} />);
    await screen.findByText('Lotto terminato.'); expect(apiClient.get).toHaveBeenCalledTimes(2);
  });
});
