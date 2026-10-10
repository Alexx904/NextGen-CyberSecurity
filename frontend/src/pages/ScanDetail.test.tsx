import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { render, screen, waitFor, act } from '@testing-library/react';
import { MemoryRouter, Routes, Route } from 'react-router-dom';
import { ScanDetail } from './ScanDetail';
import { DebugProvider } from '../context/DebugContext';
import { apiClient } from '../api/client';
import { debugLogger } from '../utils/debugLogger';

vi.mock('recharts', async (importOriginal) => {
  const original = await importOriginal<typeof import('recharts')>();
  return {
    ...original,
    ResponsiveContainer: ({ children }: { children: React.ReactNode }) => (
      <div style={{ width: 400, height: 300 }}>{children}</div>
    ),
  };
});

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

  public simulateOpen() {
    this.readyState = 1;
    if (this.onopen) {
      this.onopen(new Event('open'));
    }
  }

  public simulateEvent(type: string, data: any) {
    const dataStr = typeof data === 'string' ? data : JSON.stringify(data);
    const evt = { data: dataStr } as MessageEvent;
    if (this.listeners[type]) {
      this.listeners[type].forEach((cb) => cb(evt));
    }
  }

  static instances: MockEventSource[] = [];
  static reset() {
    MockEventSource.instances = [];
  }
}

describe('ScanDetail Live Real-Time Monitoring & Progressive Findings', () => {
  const originalEventSource = globalThis.EventSource;
  const scanId = '3fa85f64-5717-4562-b3fc-2c963f66afa6';

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

  const renderComponent = (initialPath = `/scans/${scanId}`) => {
    return render(
      <MemoryRouter
        initialEntries={[initialPath]}
        future={{ v7_startTransition: true, v7_relativeSplatPath: true }}
      >
        <DebugProvider>
          <Routes>
            <Route path="/scans/:id" element={<ScanDetail />} />
          </Routes>
        </DebugProvider>
      </MemoryRouter>
    );
  };

  it('renders completed scan with reports and static findings', async () => {
    vi.spyOn(apiClient, 'get').mockImplementation(async (url: string) => {
      if (url.includes(`/scans/${scanId}/findings`)) {
        return {
          data: {
            success: true,
            data: {
              items: [
                {
                  id: 'f-1',
                  scan_id: scanId,
                  tenant_id: 't-1',
                  host_ip: '192.168.1.1',
                  port: 80,
                  protocol: 'tcp',
                  severity: 'LOW',
                  created_at: new Date().toISOString(),
                },
              ],
              total: 1,
              limit: 100,
              offset: 0,
            },
          },
        } as any;
      }
      if (url.includes(`/scans/${scanId}`)) {
        return {
          data: {
            success: true,
            data: {
              id: scanId,
              tenant_id: 't-1',
              target_id: 'target-1',
              profile_id: 'prof-1',
              trigger_type: 'MANUAL',
              status: 'COMPLETED',
              summary_stats: {
                total_hosts: 1,
                up_hosts: 1,
                down_hosts: 0,
                total_open_ports: 1,
                critical_count: 0,
                high_count: 0,
                medium_count: 0,
                low_count: 1,
                info_count: 0,
                duration_seconds: 12.5,
              },
              created_at: new Date().toISOString(),
              updated_at: new Date().toISOString(),
            },
          },
        } as any;
      }
      return { data: { success: true } } as any;
    });

    renderComponent();

    await waitFor(() => {
      expect(screen.getByText('Completata')).toBeDefined();
      expect(screen.getByText('Scarica Report PDF (White-Label)')).toBeDefined();
      expect(screen.getByText('Esporta CSV')).toBeDefined();
      expect(screen.queryByText(/Scansione in tempo reale attiva/i)).toBeNull();
      expect(screen.getByText('192.168.1.1')).toBeDefined();
    });

    // Per una scansione già completata, nessuna connessione SSE aperta
    expect(MockEventSource.instances.length).toBe(0);
  });

  it('connects to SSE stream for RUNNING scan, updates live phases, and streams findings incrementally', async () => {
    vi.spyOn(apiClient, 'get').mockImplementation(async (url: string) => {
      if (url.includes(`/scans/${scanId}/findings`)) {
        return {
          data: {
            success: true,
            data: {
              items: [],
              total: 0,
              limit: 100,
              offset: 0,
            },
          },
        } as any;
      }
      if (url.includes(`/scans/${scanId}`)) {
        return {
          data: {
            success: true,
            data: {
              id: scanId,
              tenant_id: 't-1',
              target_id: 'target-1',
              profile_id: 'prof-1',
              trigger_type: 'MANUAL',
              status: 'RUNNING',
              summary_stats: {
                total_hosts: 1,
                up_hosts: 1,
                down_hosts: 0,
                total_open_ports: 0,
                critical_count: 0,
                high_count: 0,
                medium_count: 0,
                low_count: 0,
                info_count: 0,
                duration_seconds: 0,
              },
              created_at: new Date().toISOString(),
              updated_at: new Date().toISOString(),
            },
          },
        } as any;
      }
      return { data: { success: true } } as any;
    });

    renderComponent();

    // 1. Verifica caricamento iniziale scansione in RUNNING
    await waitFor(() => {
      expect(screen.getByText(/Scansione in tempo reale attiva/i)).toBeDefined();
      expect(screen.getByText('LIVE STREAM')).toBeDefined();
    });

    // EventSource deve essere connesso al topic di scansione
    expect(MockEventSource.instances.length).toBe(1);
    const es = MockEventSource.instances[0];
    expect(es.url).toContain(`/api/v1/scans/${scanId}/stream`);

    // 2. Simula open SSE
    act(() => {
      es.simulateOpen();
    });

    // 3. Simula evento di cambio fase PORT_SCAN (35%)
    act(() => {
      es.simulateEvent('phase', {
        scan_id: scanId,
        phase: 'PORT_SCAN',
        message: 'Scansione porte TCP SYN in corso...',
        percentage: 35,
      });
    });

    await waitFor(() => {
      expect(screen.getByText('Scansione porte TCP SYN in corso...')).toBeDefined();
      expect(screen.getByText('35%')).toBeDefined();
      expect(screen.getByText('PORT_SCAN (35%)')).toBeDefined();
    });

    // 4. Simula evento finding_discovered in tempo reale
    const liveFinding = {
      id: 'f-live-99',
      scan_id: scanId,
      tenant_id: 't-1',
      host_ip: '10.0.0.50',
      port: 443,
      protocol: 'tcp',
      service_name: 'https',
      service_version: 'Apache 2.4.41',
      cve_id: 'CVE-2023-9999',
      cvss_score: 9.8,
      severity: 'CRITICAL',
      created_at: new Date().toISOString(),
    };

    act(() => {
      es.simulateEvent('finding_discovered', {
        scan_id: scanId,
        finding: liveFinding,
      });
    });

    // Il finding deve apparire nella tabella senza ricaricamento
    await waitFor(() => {
      expect(screen.getByText('10.0.0.50')).toBeDefined();
      expect(screen.getByText(/443\/TCP/i)).toBeDefined();
      expect(screen.getByText('CRITICAL')).toBeDefined();
    });

    // 5. Simula completamento scansione via status_change
    act(() => {
      es.simulateEvent('status_change', {
        scan_id: scanId,
        status: 'COMPLETED',
        summary_stats: {
          total_hosts: 1,
          up_hosts: 1,
          down_hosts: 0,
          total_open_ports: 1,
          critical_count: 1,
          high_count: 0,
          medium_count: 0,
          low_count: 0,
          info_count: 0,
          duration_seconds: 18.2,
        },
      });
    });

    await waitFor(() => {
      expect(screen.getByText('Completata')).toBeDefined();
      expect(screen.queryByText(/Scansione in tempo reale attiva/i)).toBeNull();
      expect(screen.getByText('Scarica Report PDF (White-Label)')).toBeDefined();
    });
  });
});
