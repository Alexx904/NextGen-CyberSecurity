import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useSSE } from './useSSE';
import { debugLogger } from '../utils/debugLogger';

// Mock MockEventSource
class MockEventSource {
  public url: string;
  public readyState: number = 0;
  public onopen: ((e: Event) => void) | null = null;
  public onerror: ((e: Event) => void) | null = null;
  public onmessage: ((e: MessageEvent) => void) | null = null;
  public listeners: Record<string, ((e: any) => void)[]> = {};
  public isClosed: boolean = false;

  constructor(url: string) {
    this.url = url;
    MockEventSource.instances.push(this);
  }

  public addEventListener(event: string, cb: (e: any) => void) {
    if (!this.listeners[event]) {
      this.listeners[event] = [];
    }
    this.listeners[event].push(cb);
  }

  public removeEventListener(event: string, cb: (e: any) => void) {
    if (this.listeners[event]) {
      this.listeners[event] = this.listeners[event].filter((fn) => fn !== cb);
    }
  }

  public close() {
    this.isClosed = true;
    this.readyState = 2;
  }

  // Helper per simulare eventi nei test
  public simulateOpen() {
    this.readyState = 1;
    if (this.onopen) {
      this.onopen(new Event('open'));
    }
  }

  public simulateError() {
    this.readyState = 2;
    if (this.onerror) {
      this.onerror(new Event('error'));
    }
  }

  public simulateEvent(type: string, data: any) {
    const dataStr = typeof data === 'string' ? data : JSON.stringify(data);
    const evt = { data: dataStr } as MessageEvent;
    if (type === 'message' && this.onmessage) {
      this.onmessage(evt);
    }
    if (this.listeners[type]) {
      this.listeners[type].forEach((cb) => cb(evt));
    }
  }

  static instances: MockEventSource[] = [];
  static reset() {
    MockEventSource.instances = [];
  }
}

describe('useSSE Hook', () => {
  const originalEventSource = globalThis.EventSource;

  beforeEach(() => {
    MockEventSource.reset();
    (globalThis as any).EventSource = MockEventSource;
    localStorage.clear();
    debugLogger.clear();
  });

  afterEach(() => {
    (globalThis as any).EventSource = originalEventSource;
    vi.restoreAllMocks();
  });

  it('does not establish connection when url is null or enabled is false', () => {
    const { result, rerender } = renderHook(
      ({ url, enabled }) => useSSE({ url, enabled }),
      { initialProps: { url: null as string | null, enabled: true } }
    );

    expect(result.current.isConnected).toBe(false);
    expect(MockEventSource.instances.length).toBe(0);

    rerender({ url: '/api/v1/stream', enabled: false });
    expect(result.current.isConnected).toBe(false);
    expect(MockEventSource.instances.length).toBe(0);
  });

  it('connects to EventSource and injects stored JWT token', () => {
    localStorage.setItem('nextgen_access_token', 'mock-jwt-token-123');

    const onOpen = vi.fn();
    const { result } = renderHook(() =>
      useSSE({
        url: '/api/v1/network-discovery/stream?session_id=abc',
        onOpen,
      })
    );

    expect(MockEventSource.instances.length).toBe(1);
    const es = MockEventSource.instances[0];
    expect(es.url).toContain('token=mock-jwt-token-123');
    expect(result.current.isConnecting).toBe(true);

    act(() => {
      es.simulateOpen();
    });

    expect(result.current.isConnected).toBe(true);
    expect(result.current.isConnecting).toBe(false);
    expect(onOpen).toHaveBeenCalledTimes(1);
  });

  it('receives custom SSE events, parses JSON, and logs to debugLogger', () => {
    const onEvent = vi.fn();
    renderHook(() =>
      useSSE({
        url: '/api/v1/network-discovery/stream?session_id=123',
        onEvent,
      })
    );

    const es = MockEventSource.instances[0];
    act(() => {
      es.simulateOpen();
    });

    const payload = {
      session_id: '123',
      device: {
        ip: '192.168.1.55',
        hostname: 'nas-server',
        latency_ms: 1.2,
        status: 'up',
      },
    };

    act(() => {
      es.simulateEvent('device_discovered', payload);
    });

    expect(onEvent).toHaveBeenCalledWith('device_discovered', payload);

    // Verifica logging in debugLogger
    const logs = debugLogger.getEntries();
    const sseLog = logs.find((l) => l.source === 'sse' && l.title === 'SSE: device_discovered');
    expect(sseLog).toBeDefined();
    expect(sseLog?.responseData).toEqual(payload);
  });

  it('handles error events and updates error state', () => {
    const onError = vi.fn();
    const { result } = renderHook(() =>
      useSSE({
        url: '/api/v1/network-discovery/stream?session_id=123',
        onError,
      })
    );

    const es = MockEventSource.instances[0];
    act(() => {
      es.simulateError();
    });

    expect(result.current.isConnected).toBe(false);
    expect(result.current.error).toContain('Connessione SSE interrotta');
    expect(onError).toHaveBeenCalledTimes(1);
  });

  it('closes connection on close() call and on unmount', () => {
    const { result, unmount } = renderHook(() =>
      useSSE({
        url: '/api/v1/network-discovery/stream?session_id=123',
      })
    );

    const es = MockEventSource.instances[0];
    expect(es.isClosed).toBe(false);

    act(() => {
      result.current.close();
    });

    expect(es.isClosed).toBe(true);
    expect(result.current.isConnected).toBe(false);

    unmount();
    expect(es.isClosed).toBe(true);
  });
});

