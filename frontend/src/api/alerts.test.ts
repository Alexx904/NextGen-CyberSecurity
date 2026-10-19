import { beforeEach, expect, it, vi } from 'vitest';
import { apiClient } from './client';
import { alertsApi } from './alerts';
vi.mock('./client', () => ({ apiClient: { get: vi.fn(), post: vi.fn(), put: vi.fn(), delete: vi.fn() } }));
beforeEach(() => vi.resetAllMocks());
it('posts a synthetic acknowledged test separately from save and preserves omitted credentials', async () => {
  vi.mocked(apiClient.post).mockResolvedValue({ data: { success: true, data: { accepted: true } } });
  await alertsApi.test('webhook', { endpoint: 'https://example.test/fixture', hmac_secret: 'synthetic-only' });
  expect(apiClient.post).toHaveBeenCalledWith('/alert-channels/test', { kind: 'webhook', config: { endpoint: 'https://example.test/fixture', hmac_secret: 'synthetic-only' }, confirmed_external_send: true });
  vi.mocked(apiClient.put).mockResolvedValue({ data: { success: true, data: { id: 'channel' } } });
  const metadata = { name: 'Renamed', kind: 'webhook' as const, enabled: false, events: ['finding.critical'], expected_updated_at: 'revision' };
  await alertsApi.save(metadata, 'channel'); expect(apiClient.put).toHaveBeenCalledWith('/alert-channels/channel', metadata);
  expect(apiClient.post).toHaveBeenCalledTimes(1);
  vi.mocked(apiClient.get).mockResolvedValue({ data: { success: false } });
  await expect(alertsApi.list()).rejects.toThrow('non completata');
});
