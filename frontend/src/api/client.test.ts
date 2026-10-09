import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { apiClient, storeTokens, purgeStoredTokens } from './client';
import { debugLogger } from '../utils/debugLogger';

describe('apiClient Debug Interceptors', () => {
  const originalAdapter = apiClient.defaults.adapter;

  beforeEach(() => {
    debugLogger.clear();
    purgeStoredTokens();
  });

  afterEach(() => {
    apiClient.defaults.adapter = originalAdapter;
    vi.restoreAllMocks();
  });

  it('injects Bearer token and logs request and response to debugLogger', async () => {
    storeTokens({
      access_token: 'valid_access_token_jwt',
      refresh_token: 'valid_refresh_token_jwt',
      expires_at: '2026-10-09T22:00:00Z',
    });

    apiClient.defaults.adapter = async (config) => {
      return {
        data: { success: true, data: 'ok' },
        status: 200,
        statusText: 'OK',
        headers: {},
        config,
      };
    };

    const response = await apiClient.get('/test-endpoint');
    expect(response.status).toBe(200);

    const entries = debugLogger.getEntries();
    // Dovrebbe esserci 1 log di REQ e 1 log di RES
    expect(entries.length).toBeGreaterThanOrEqual(2);

    const reqEntry = entries.find((e) => e.title.includes('API REQ'));
    expect(reqEntry).toBeDefined();
    expect(reqEntry?.url).toBe('/test-endpoint');
    expect(reqEntry?.method).toBe('GET');

    const resEntry = entries.find((e) => e.title.includes('API RES'));
    expect(resEntry).toBeDefined();
    expect(resEntry?.status).toBe(200);
    expect(resEntry?.durationMs).toBeDefined();
    expect(typeof resEntry?.durationMs).toBe('number');
  });

  it('logs API error to debugLogger on failure', async () => {
    apiClient.defaults.adapter = async (config) => {
      const error: any = new Error('Gateway Timeout');
      error.config = config;
      error.response = { status: 504, data: { message: 'Gateway Timeout' } };
      throw error;
    };

    await expect(apiClient.get('/timeout-endpoint')).rejects.toThrow();

    const entries = debugLogger.getEntries();
    const errEntry = entries.find((e) => e.title.includes('API ERR'));
    expect(errEntry).toBeDefined();
    expect(errEntry?.status).toBe(504);
    expect(errEntry?.level).toBe('error');
    expect(errEntry?.durationMs).toBeDefined();
  });
});
