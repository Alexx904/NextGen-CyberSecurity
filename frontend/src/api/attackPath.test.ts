import { beforeEach, expect, it, vi } from 'vitest';
import { attackPathApi } from './attackPath';
import { apiClient } from './client';

vi.mock('./client', () => ({ apiClient: { post: vi.fn() } }));
beforeEach(() => { vi.resetAllMocks(); });

it('encodes the chokepoint identifier and submits only the chosen server', async () => {
  vi.mocked(apiClient.post).mockResolvedValue({ data: { success: true, data: { node_id: 'node' } } });
  await attackPathApi.previewRemediation('service/?:#', 'server');
  await attackPathApi.remediate('service/?:#', 'server');
  expect(apiClient.post).toHaveBeenNthCalledWith(1, '/attack-paths/chokepoints/service%2F%3F%3A%23/preview', { server_id: 'server' });
  expect(apiClient.post).toHaveBeenNthCalledWith(2, '/attack-paths/chokepoints/service%2F%3F%3A%23/remediate', { server_id: 'server' });
});

it('rejects a failed remediation envelope instead of acknowledging application', async () => {
  vi.mocked(apiClient.post).mockResolvedValue({ data: { success: false, error: { message: 'Chokepoint obsoleto' } } });
  await expect(attackPathApi.remediate('node', 'server')).rejects.toThrow('Chokepoint obsoleto');
});
