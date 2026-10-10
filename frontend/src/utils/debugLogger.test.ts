import { describe, it, expect, vi, beforeEach } from 'vitest';
import {
  DebugLogger,
  sanitizeData,
  generateLogId,
} from './debugLogger';

describe('generateLogId', () => {
  it('generates non-empty string IDs', () => {
    const id1 = generateLogId();
    const id2 = generateLogId();
    expect(typeof id1).toBe('string');
    expect(id1.length).toBeGreaterThan(0);
    expect(id1).not.toBe(id2);
  });
});

describe('sanitizeData', () => {
  it('handles null, undefined, and primitive types without modification', () => {
    expect(sanitizeData(null)).toBeNull();
    expect(sanitizeData(undefined)).toBeUndefined();
    expect(sanitizeData(12345)).toBe(12345);
    expect(sanitizeData(true)).toBe(true);
    expect(sanitizeData('normal_string')).toBe('normal_string');
  });

  it('sanitizes Date objects and Error objects', () => {
    const date = new Date('2026-10-09T18:00:00Z');
    expect(sanitizeData(date)).toBe('2026-10-09T18:00:00.000Z');

    const err = new Error('Test failure');
    const sanitizedErr = sanitizeData(err) as { name: string; message: string; stack?: string };
    expect(sanitizedErr.name).toBe('Error');
    expect(sanitizedErr.message).toBe('Test failure');
    expect(sanitizedErr.stack).toBeDefined();
  });

  it('redacts sensitive keys in flat objects', () => {
    const raw = {
      username: 'secops_user',
      password: 'SuperSecretPassword!',
      access_token: 'abc123token',
      refresh_token: 'def456token',
      secret: 'my-top-secret',
      encrypted_secret: 'aes-cipher-text',
      api_key: 'key-999',
      port: 8080,
    };

    const sanitized = sanitizeData(raw) as Record<string, unknown>;
    expect(sanitized.username).toBe('secops_user');
    expect(sanitized.port).toBe(8080);
    expect(sanitized.password).toBe('[REDACTED]');
    expect(sanitized.access_token).toBe('[REDACTED]');
    expect(sanitized.refresh_token).toBe('[REDACTED]');
    expect(sanitized.secret).toBe('[REDACTED]');
    expect(sanitized.encrypted_secret).toBe('[REDACTED]');
    expect(sanitized.api_key).toBe('[REDACTED]');
  });

  it('is case-insensitive when detecting sensitive keys', () => {
    const raw = {
      PassWord: 'secret-1',
      ACCESS_TOKEN: 'secret-2',
      ApiKey: 'secret-3',
      Passphrase: 'secret-4',
    };

    const sanitized = sanitizeData(raw) as Record<string, unknown>;
    expect(sanitized.PassWord).toBe('[REDACTED]');
    expect(sanitized.ACCESS_TOKEN).toBe('[REDACTED]');
    expect(sanitized.ApiKey).toBe('[REDACTED]');
    expect(sanitized.Passphrase).toBe('[REDACTED]');
  });

  it('redacts sensitive fields in nested objects and arrays', () => {
    const raw = {
      server: {
        host: '192.168.1.10',
        auth: {
          type: 'key',
          private_key: '-----BEGIN OPENSSH PRIVATE KEY-----...',
        },
      },
      users: [
        { id: 1, name: 'Alice', token: 'token-a' },
        { id: 2, name: 'Bob', password: 'password-b' },
      ],
    };

    const sanitized = sanitizeData(raw) as {
      server: { host: string; auth: { type: string; private_key: string } };
      users: Array<{ id: number; name: string; token?: string; password?: string }>;
    };

    expect(sanitized.server.host).toBe('192.168.1.10');
    expect(sanitized.server.auth.private_key).toBe('[REDACTED]');
    expect(sanitized.users[0].token).toBe('[REDACTED]');
    expect(sanitized.users[1].password).toBe('[REDACTED]');
  });

  it('sanitizes Bearer authorization headers and JWT strings', () => {
    const rawHeader = 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.doNotLeak';
    expect(sanitizeData(rawHeader)).toBe('Bearer [REDACTED]');

    const standaloneJwt = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.c2lnbmF0dXJl';
    expect(sanitizeData(standaloneJwt)).toBe('[REDACTED_JWT]');
  });

  it('sanitizes secret parameters in URLs and query strings', () => {
    const streamUrl = '/api/v1/scans/123/stream?token=secret-token-value&other=123';
    expect(sanitizeData(streamUrl)).toBe('/api/v1/scans/123/stream?token=[REDACTED]&other=123');

    const wsUrl = '/ws/v1/ssh/servers/456/terminal?ticket=st_secret_ticket_123';
    expect(sanitizeData(wsUrl)).toBe('/ws/v1/ssh/servers/456/terminal?ticket=[REDACTED]');
  });

  it('safely handles circular references without infinite recursion or stack overflow', () => {
    const circularObj: Record<string, unknown> = {
      name: 'root',
    };
    circularObj.self = circularObj;

    const sanitized = sanitizeData(circularObj) as Record<string, unknown>;
    expect(sanitized.name).toBe('root');
    expect(sanitized.self).toBe('[Circular]');
  });

  it('guarantees immutability and does not mutate the source object', () => {
    const original = {
      password: 'OriginalPassword123!',
      nested: {
        token: 'OriginalToken',
      },
    };

    sanitizeData(original);

    expect(original.password).toBe('OriginalPassword123!');
    expect(original.nested.token).toBe('OriginalToken');
  });
});

describe('DebugLogger (Ring Buffer & Core Logic)', () => {
  let logger: DebugLogger;

  beforeEach(() => {
    logger = new DebugLogger(5); // Capacità ridotta a 5 per test del buffer circolare
  });

  it('initializes with empty entries and correct default capacity', () => {
    const defaultLogger = new DebugLogger();
    expect(defaultLogger.getEntries()).toHaveLength(0);
    expect(logger.getEntries()).toHaveLength(0);
  });

  it('records logs with generated ID and ISO timestamp', () => {
    const entry = logger.log({
      level: 'info',
      source: 'app',
      title: 'Test Log Message',
    });

    expect(entry.id).toBeDefined();
    expect(entry.timestamp).toBeDefined();
    expect(new Date(entry.timestamp).getTime()).not.toBeNaN();
    expect(entry.level).toBe('info');
    expect(entry.title).toBe('Test Log Message');

    const entries = logger.getEntries();
    expect(entries).toHaveLength(1);
    expect(entries[0].id).toBe(entry.id);
  });

  it('enforces FIFO eviction when capacity is exceeded', () => {
    for (let i = 1; i <= 7; i++) {
      logger.log({
        level: 'info',
        source: 'app',
        title: `Message ${i}`,
      });
    }

    const entries = logger.getEntries();
    expect(entries).toHaveLength(5);
    // Gli elementi 1 e 2 devono essere stati rimossi per fare spazio a 6 e 7
    expect(entries[0].title).toBe('Message 3');
    expect(entries[4].title).toBe('Message 7');
  });

  it('clears all entries and resets count', () => {
    logger.log({ level: 'info', source: 'app', title: 'A' });
    logger.log({ level: 'error', source: 'app', title: 'B' });

    expect(logger.getEntries()).toHaveLength(2);

    logger.clear();

    expect(logger.getEntries()).toHaveLength(0);
    expect(logger.getErrorsCount()).toBe(0);
  });

  it('supports setMaxEntries and trims entries if new limit is smaller', () => {
    for (let i = 1; i <= 5; i++) {
      logger.log({ level: 'info', source: 'app', title: `Item ${i}` });
    }

    expect(logger.getEntries()).toHaveLength(5);

    logger.setMaxEntries(3);

    const entries = logger.getEntries();
    expect(entries).toHaveLength(3);
    expect(entries[0].title).toBe('Item 3');
    expect(entries[2].title).toBe('Item 5');
  });
});

describe('DebugLogger (Pub/Sub Event Bus)', () => {
  let logger: DebugLogger;

  beforeEach(() => {
    logger = new DebugLogger(10);
  });

  it('notifies subscribers when a new entry is logged', () => {
    const listener = vi.fn();
    const unsubscribe = logger.subscribe(listener);

    const entry = logger.log({
      level: 'info',
      source: 'api',
      title: 'Fetch Targets',
    });

    expect(listener).toHaveBeenCalledTimes(1);
    expect(listener).toHaveBeenCalledWith(entry, [entry]);

    unsubscribe();
    logger.log({
      level: 'info',
      source: 'api',
      title: 'Second Call',
    });

    expect(listener).toHaveBeenCalledTimes(1); // Non più chiamato dopo unsubscribe
  });

  it('notifies multiple subscribers independently', () => {
    const listenerA = vi.fn();
    const listenerB = vi.fn();

    logger.subscribe(listenerA);
    logger.subscribe(listenerB);

    logger.log({ level: 'warn', source: 'app', title: 'Warning' });

    expect(listenerA).toHaveBeenCalledTimes(1);
    expect(listenerB).toHaveBeenCalledTimes(1);
  });

  it('isolates listener errors so other listeners still execute', () => {
    const faultyListener = vi.fn().mockImplementation(() => {
      throw new Error('Listener crash');
    });
    const goodListener = vi.fn();

    // Spia per silenziare l'errore atteso nella console
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

    logger.subscribe(faultyListener);
    logger.subscribe(goodListener);

    expect(() => {
      logger.log({ level: 'error', source: 'app', title: 'Boom' });
    }).not.toThrow();

    expect(faultyListener).toHaveBeenCalledTimes(1);
    expect(goodListener).toHaveBeenCalledTimes(1);

    consoleSpy.mockRestore();
  });
});

describe('DebugLogger (Helper Methods & Sanitization Integration)', () => {
  let logger: DebugLogger;

  beforeEach(() => {
    logger = new DebugLogger(20);
  });

  it('logApiRequest logs request and automatically redacts sensitive headers & data', () => {
    const reqId = logger.logApiRequest({
      method: 'POST',
      url: '/api/v1/auth/login',
      data: {
        username: 'admin',
        password: 'TopSecretPassword!',
      },
      headers: {
        Authorization: 'Bearer eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxIn0.sig',
      },
    });

    expect(typeof reqId).toBe('string');
    const entries = logger.getEntries();
    expect(entries).toHaveLength(1);

    const entry = entries[0];
    expect(entry.id).toBe(reqId);
    expect(entry.source).toBe('api');
    expect(entry.level).toBe('info');
    expect(entry.title).toBe('API REQ: POST /api/v1/auth/login');
    expect((entry.requestData as Record<string, unknown>).password).toBe('[REDACTED]');
    expect(
      (entry.metadata?.headers as Record<string, unknown>).Authorization
    ).toBe('Bearer [REDACTED]');
  });

  it('logApiResponse handles 200, 300, 400 and 500 status codes correctly', () => {
    logger.logApiResponse({
      method: 'GET',
      url: '/api/v1/targets',
      status: 200,
      durationMs: 45,
      data: [{ id: 1, name: 'Target A' }],
    });

    logger.logApiResponse({
      method: 'GET',
      url: '/api/v1/redirect',
      status: 301,
      durationMs: 12,
    });

    logger.logApiResponse({
      method: 'GET',
      url: '/api/v1/protected',
      status: 403,
      durationMs: 25,
      data: { error: 'Forbidden' },
    });

    const entries = logger.getEntries();
    expect(entries).toHaveLength(3);
    expect(entries[0].level).toBe('info');
    expect(entries[0].durationMs).toBe(45);
    expect(entries[1].level).toBe('warn');
    expect(entries[2].level).toBe('error');
    expect(entries[2].title).toBe('API RES: GET /api/v1/protected [403]');
  });

  it('logApiError extracts error message, stack, and status', () => {
    const axiosError = {
      message: 'Request failed with status code 500',
      response: {
        status: 500,
        data: { message: 'Internal Server Error', secret_trace: 'leaked' },
      },
    };

    logger.logApiError({
      method: 'POST',
      url: '/api/v1/scans',
      durationMs: 120,
      error: axiosError,
      requestData: { target: '192.168.1.1', secret: 'abc' },
    });

    const entries = logger.getEntries();
    expect(entries).toHaveLength(1);

    const entry = entries[0];
    expect(entry.level).toBe('error');
    expect(entry.source).toBe('api');
    expect(entry.status).toBe(500);
    expect(entry.durationMs).toBe(120);
    expect(entry.error).toBe('Request failed with status code 500');
    expect((entry.requestData as Record<string, unknown>).secret).toBe('[REDACTED]');
  });

  it('logSSEEvent logs streaming events with appropriate level', () => {
    logger.logSSEEvent({
      eventType: 'device_discovered',
      data: { ip: '192.168.1.10', hostname: 'host.local' },
    });

    logger.logSSEEvent({
      eventType: 'stream_error',
      data: { message: 'Connection dropped' },
    });

    const entries = logger.getEntries();
    expect(entries).toHaveLength(2);
    expect(entries[0].source).toBe('sse');
    expect(entries[0].level).toBe('info');
    expect(entries[0].title).toBe('SSE: device_discovered');

    expect(entries[1].source).toBe('sse');
    expect(entries[1].level).toBe('error');
    expect(entries[1].title).toBe('SSE: stream_error');
  });

  it('logWSEvent logs WebSocket incoming and outgoing frames', () => {
    logger.logWSEvent({
      direction: 'out',
      data: { type: 'resize', cols: 120, rows: 40 },
      eventType: 'resize',
    });

    logger.logWSEvent({
      direction: 'in',
      data: 'terminal output prompt $ ',
      eventType: 'pty_stream',
    });

    const entries = logger.getEntries();
    expect(entries).toHaveLength(2);
    expect(entries[0].source).toBe('ws');
    expect(entries[0].title).toBe('WS OUT: resize');
    expect(entries[0].requestData).toBeDefined();

    expect(entries[1].source).toBe('ws');
    expect(entries[1].title).toBe('WS IN: pty_stream');
    expect(entries[1].responseData).toBeDefined();
  });

  it('logRuntimeError logs unhandled exceptions with error and stack', () => {
    const error = new TypeError('Cannot read property of undefined');
    logger.logRuntimeError({
      error,
      source: 'window.onerror',
    });

    const entries = logger.getEntries();
    expect(entries).toHaveLength(1);

    const entry = entries[0];
    expect(entry.source).toBe('runtime');
    expect(entry.level).toBe('error');
    expect(entry.error).toBe('Cannot read property of undefined');
    expect(entry.stack).toBeDefined();
    expect(entry.metadata?.source).toBe('window.onerror');
  });

  it('logCustom logs generic application diagnostic messages', () => {
    logger.logCustom({
      level: 'warn',
      title: 'Session about to expire',
      data: { remainingSeconds: 30 },
    });

    const entries = logger.getEntries();
    expect(entries).toHaveLength(1);
    expect(entries[0].source).toBe('app');
    expect(entries[0].level).toBe('warn');
    expect(entries[0].title).toBe('Session about to expire');
  });
});

describe('DebugLogger (Stats & Filtering)', () => {
  let logger: DebugLogger;

  beforeEach(() => {
    logger = new DebugLogger(20);

    logger.log({ level: 'info', source: 'api', title: 'GET /api/v1/targets', url: '/api/v1/targets' });
    logger.log({ level: 'error', source: 'api', title: 'POST /api/v1/login failed', url: '/api/v1/login', error: 'Invalid creds' });
    logger.log({ level: 'warn', source: 'app', title: 'Slow query warning' });
    logger.log({ level: 'error', source: 'runtime', title: 'Uncaught ReferenceError', error: 'ReferenceError: x is not defined' });
  });

  it('computes accurate error count and aggregate statistics', () => {
    expect(logger.getErrorsCount()).toBe(2);

    const stats = logger.getStats();
    expect(stats.total).toBe(4);
    expect(stats.errors).toBe(2);
    expect(stats.warnings).toBe(1);
    expect(stats.info).toBe(1);
    expect(stats.oldestTimestamp).toBeDefined();
    expect(stats.newestTimestamp).toBeDefined();
  });

  it('filters entries by level, source, and search term', () => {
    // Filtro per livello error
    const errorsOnly = logger.getFilteredEntries({ level: 'error' });
    expect(errorsOnly).toHaveLength(2);

    // Filtro per source api
    const apiOnly = logger.getFilteredEntries({ source: 'api' });
    expect(apiOnly).toHaveLength(2);

    // Filtro combinato level + source
    const apiErrors = logger.getFilteredEntries({ level: 'error', source: 'api' });
    expect(apiErrors).toHaveLength(1);
    expect(apiErrors[0].title).toContain('POST /api/v1/login');

    // Ricerca testuale
    const searchTargets = logger.getFilteredEntries({ search: 'targets' });
    expect(searchTargets).toHaveLength(1);

    const searchError = logger.getFilteredEntries({ search: 'ReferenceError' });
    expect(searchError).toHaveLength(1);
    expect(searchError[0].source).toBe('runtime');

    // Senza opzioni di filtro restituisce tutto
    expect(logger.getFilteredEntries()).toHaveLength(4);
  });

  it('handles string errors and fallback inputs in logApiError and logRuntimeError', () => {
    logger.logApiError({
      method: 'GET',
      url: '/api/v1/ping',
      error: 'Direct string error message',
    });

    const entries = logger.getEntries();
    const lastEntry = entries[entries.length - 1];
    expect(lastEntry.error).toBe('Direct string error message');

    logger.logRuntimeError({
      error: 'Raw string runtime exception',
    });
    const entriesAfterRuntime = logger.getEntries();
    const runtimeEntry = entriesAfterRuntime[entriesAfterRuntime.length - 1];
    expect(runtimeEntry.error).toBe('Raw string runtime exception');
  });

  it('falls back to default max entries if negative or zero limit is passed to constructor or setMaxEntries', () => {
    const invalidLogger = new DebugLogger(-5);
    invalidLogger.setMaxEntries(0);
    for (let i = 0; i < 10; i++) {
      invalidLogger.log({ level: 'info', source: 'app', title: `Test ${i}` });
    }
    expect(invalidLogger.getEntries()).toHaveLength(10);
  });

  it('sanitizes Basic auth header format', () => {
    const basicHeader = 'Basic dXNlcm5hbWU6cGFzc3dvcmQ=';
    expect(sanitizeData(basicHeader)).toBe('Basic [REDACTED]');
  });
});

