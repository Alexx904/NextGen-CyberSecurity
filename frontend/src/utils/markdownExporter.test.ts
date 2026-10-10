import { describe, it, expect, vi } from 'vitest';
import {
  generateAiMarkdownReport,
  exportLogsToJson,
  copyTextToClipboard,
  triggerFileDownload,
} from './markdownExporter';
import { DebugEntry, DebugLoggerStats } from '../types/debug';

describe('markdownExporter', () => {
  const sampleStats: DebugLoggerStats = {
    total: 3,
    errors: 1,
    warnings: 1,
    info: 1,
    oldestTimestamp: '2026-10-09T18:00:00Z',
    newestTimestamp: '2026-10-09T18:01:00Z',
  };

  const sampleEntries: DebugEntry[] = [
    {
      id: 'log-1',
      timestamp: '2026-10-09T18:00:00.000Z',
      level: 'info',
      source: 'api',
      title: 'GET /api/v1/targets',
      url: '/api/v1/targets',
      method: 'GET',
      status: 200,
      durationMs: 40,
    },
    {
      id: 'log-2',
      timestamp: '2026-10-09T18:00:30.000Z',
      level: 'warn',
      source: 'app',
      title: 'Memory high warning',
    },
    {
      id: 'log-3',
      timestamp: '2026-10-09T18:01:00.000Z',
      level: 'error',
      source: 'api',
      title: 'POST /api/v1/scans [500]',
      url: '/api/v1/scans',
      method: 'POST',
      status: 500,
      durationMs: 150,
      error: 'Internal Server Error',
      requestData: { target: '192.168.1.1' },
      responseData: { message: 'Database connection failed' },
      stack: 'Error: Database connection failed\n    at query (db.go:42)',
    },
  ];

  it('generates structured Markdown report with session summary, error table, and deep-dive cards', () => {
    const md = generateAiMarkdownReport(sampleEntries, sampleStats, '/scans');

    expect(md).toContain('# 🛡️ NextGen CyberSecurity - Diagnostic Session Report');
    expect(md).toContain('## 1. Riepilogo Sessione');
    expect(md).toContain('/scans');
    expect(md).toContain('**1**'); // 1 error

    expect(md).toContain('## 2. Tabella Sinottica degli Errori');
    expect(md).toContain('POST /api/v1/scans');
    expect(md).toContain('500');
    expect(md).toContain('150ms');

    expect(md).toContain('## 3. Schede Dettagliate degli Errori (Deep-Dive)');
    expect(md).toContain('Errore #1');
    expect(md).toContain('"target": "192.168.1.1"');
    expect(md).toContain('"message": "Database connection failed"');
    expect(md).toContain('at query (db.go:42)');

    expect(md).toContain('## 4. Cronologia Ultimi Eventi Recenti (Contesto)');
  });

  it('generates clean report when no errors exist', () => {
    const cleanEntries: DebugEntry[] = [
      {
        id: 'log-1',
        timestamp: '2026-10-09T18:00:00.000Z',
        level: 'info',
        source: 'api',
        title: 'GET /api/v1/health',
      },
    ];
    const cleanStats: DebugLoggerStats = {
      total: 1,
      errors: 0,
      warnings: 0,
      info: 1,
    };

    const md = generateAiMarkdownReport(cleanEntries, cleanStats, '/');

    expect(md).toContain('Nessun errore critico rilevato');
    expect(md).not.toContain('Schede Dettagliate degli Errori (Deep-Dive)');
  });

  it('exports logs to valid formatted JSON string', () => {
    const jsonStr = exportLogsToJson(sampleEntries);
    expect(typeof jsonStr).toBe('string');

    const parsed = JSON.parse(jsonStr);
    expect(Array.isArray(parsed)).toBe(true);
    expect(parsed).toHaveLength(3);
    expect(parsed[0].id).toBe('log-1');
  });

  it('copies text to clipboard using navigator.clipboard', async () => {
    const writeTextMock = vi.fn().mockResolvedValue(undefined);
    Object.defineProperty(navigator, 'clipboard', {
      value: {
        writeText: writeTextMock,
      },
      configurable: true,
      writable: true,
    });

    const success = await copyTextToClipboard('test copy');
    expect(success).toBe(true);
    expect(writeTextMock).toHaveBeenCalledWith('test copy');
  });

  it('triggers file download creating object URL and clicking anchor', () => {
    const createObjectURLMock = vi.fn().mockReturnValue('blob:mock-url');
    const revokeObjectURLMock = vi.fn();
    window.URL.createObjectURL = createObjectURLMock;
    window.URL.revokeObjectURL = revokeObjectURLMock;

    const clickSpy = vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => {});

    triggerFileDownload('{"a": 1}', 'logs.json', 'application/json');

    expect(createObjectURLMock).toHaveBeenCalled();
    expect(clickSpy).toHaveBeenCalled();
    expect(revokeObjectURLMock).toHaveBeenCalledWith('blob:mock-url');

    clickSpy.mockRestore();
  });

  it('redacts tokens and sensitive query params from generated markdown report', () => {
    const sensitiveEntries: DebugEntry[] = [
      {
        id: 'log-sec',
        timestamp: '2026-10-09T18:00:00.000Z',
        level: 'error',
        source: 'sse',
        title: 'SSE Connection Failed to /api/v1/scans/1/stream?token=secret-jwt-123',
        url: '/api/v1/scans/1/stream?token=secret-jwt-123',
        error: 'Authorization header Bearer secret-bearer-token was rejected',
      },
    ];
    const md = generateAiMarkdownReport(sensitiveEntries, sampleStats, '/scans?token=my-secret-token');

    expect(md).not.toContain('secret-jwt-123');
    expect(md).not.toContain('secret-bearer-token');
    expect(md).not.toContain('my-secret-token');
    expect(md).toContain('token=[REDACTED]');
    expect(md).toContain('Bearer [REDACTED]');
  });
});
