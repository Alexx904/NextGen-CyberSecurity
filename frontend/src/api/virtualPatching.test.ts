import { beforeEach, expect, it, vi } from 'vitest';
import { apiClient } from './client';
import { getVirtualPatches } from './virtualPatching';
vi.mock('./client', () => ({ apiClient: { get: vi.fn() } }));
beforeEach(() => vi.resetAllMocks());
it('requests a tenant-scoped finding and rejects mismatched or unsuccessful results', async () => {
  vi.mocked(apiClient.get).mockResolvedValue({ data: { success: true, data: { finding_id: 'id' } } });
  await expect(getVirtualPatches('id')).resolves.toEqual({ finding_id: 'id' });
  expect(apiClient.get).toHaveBeenCalledWith('/findings/id/virtual-patches');
  await expect(getVirtualPatches('other')).rejects.toThrow('non coerenti');
  vi.mocked(apiClient.get).mockResolvedValue({ data: { success: false } });
  await expect(getVirtualPatches('id')).rejects.toThrow('non disponibili');
});
