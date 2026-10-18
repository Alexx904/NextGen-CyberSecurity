import { afterEach, describe, expect, it, vi } from 'vitest';
import { discoveryApi } from './discovery';
import { apiClient } from './client';

afterEach(() => vi.restoreAllMocks());

describe('discovery persisted results', () => {
  it('returns an empty archive only for HTTP 404', async () => {
    vi.spyOn(apiClient, 'get').mockRejectedValue({ isAxiosError: true, response: { status: 404 } });
    expect(await discoveryApi.getLastResults()).toBeNull();
  });

  it.each([401, 403, 500, 503])('preserves HTTP %s failures for visible recovery errors', async (status) => {
    const error = { isAxiosError: true, response: { status } };
    vi.spyOn(apiClient, 'get').mockRejectedValue(error);
    await expect(discoveryApi.getLastResults()).rejects.toBe(error);
  });

  it('preserves network failures', async () => {
    const error = new Error('Network unavailable');
    vi.spyOn(apiClient, 'get').mockRejectedValue(error);
    await expect(discoveryApi.getLastResults()).rejects.toBe(error);
  });
});
