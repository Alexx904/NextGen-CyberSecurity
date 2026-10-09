import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { NetworkDiscovery } from './NetworkDiscovery';
import { DebugProvider } from '../context/DebugContext';
import { discoveryApi } from '../api/discovery';
import { apiClient } from '../api/client';
import { debugLogger } from '../utils/debugLogger';

// Mock EventSource
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

describe('NetworkDiscovery Live Real-Time SSE Integration', () => {
  const originalEventSource = globalThis.EventSource;

  beforeEach(() => {
    MockEventSource.reset();
    (globalThis as any).EventSource = MockEventSource;
    localStorage.clear();
    debugLogger.clear();

    // Mock API
    vi.spyOn(discoveryApi, 'getLastResults').mockResolvedValue(null);
    vi.spyOn(apiClient, 'get').mockImplementation(async (url: string) => {
      if (url === '/target-groups') {
        return { data: { success: true, data: [] } } as any;
      }
      return { data: { success: true } } as any;
    });
  });

  afterEach(() => {
    (globalThis as any).EventSource = originalEventSource;
    vi.restoreAllMocks();
  });

  const renderComponent = () =>
    render(
      <MemoryRouter>
        <DebugProvider>
          <NetworkDiscovery />
        </DebugProvider>
      </MemoryRouter>
    );

  it('renders initial state with default subnet and empty discovery message', async () => {
    renderComponent();

    expect(screen.getByText('Network Discovery')).toBeDefined();
    const input = screen.getByLabelText(/Subnet o Range IP Bersaglio/i) as HTMLInputElement;
    expect(input.value).toBe('192.168.1.0/24');

    await waitFor(() => {
      expect(screen.getByText('Nessuna scansione discovery eseguita')).toBeDefined();
    });
  });

  it('streams progressive progress and newly discovered devices into the table in real-time', async () => {
    let resolveScanPromise: (val: any) => void;
    const scanPromise = new Promise((resolve) => {
      resolveScanPromise = resolve;
    });

    vi.spyOn(discoveryApi, 'scan').mockImplementation(() => scanPromise as any);

    renderComponent();

    // Click Avvia Discovery
    const startBtn = screen.getByRole('button', { name: /Avvia Discovery/i });
    fireEvent.click(startBtn);

    // Verifica che isScanning è attivo
    await waitFor(() => {
      expect(screen.getByText(/Ricognizione LAN Live in corso/i)).toBeDefined();
    });

    // EventSource deve essere stato istanziato con session_id
    expect(MockEventSource.instances.length).toBe(1);
    const es = MockEventSource.instances[0];
    expect(es.url).toContain('/api/v1/network-discovery/stream?session_id=');

    // 1. Simula handshake SSE
    act(() => {
      es.simulateOpen();
    });

    // 2. Simula evento di avanzamento progress
    act(() => {
      es.simulateEvent('progress', {
        session_id: 'test-session',
        scanned_hosts: 25,
        total_hosts: 254,
        percentage: 9.8,
        current_ip: '192.168.1.25',
      });
    });

    // Verifica banner di avanzamento con IP corrente e percentuale
    await waitFor(() => {
      expect(screen.getByText('192.168.1.25')).toBeDefined();
      expect(screen.getByText(/25 di 254 host analizzati/i)).toBeDefined();
      expect(screen.getByText('9.8%')).toBeDefined();
    });

    // 3. Simula evento device_discovered per host 192.168.1.10
    const dev1 = {
      ip: '192.168.1.10',
      hostname: 'firewall-core',
      vendor: 'Fortinet',
      mac: '00:09:0F:AA:BB:CC',
      latency_ms: 1.45,
      status: 'up',
      open_ports: [80, 443, 22],
      scanned_at: new Date().toISOString(),
    };

    act(() => {
      es.simulateEvent('device_discovered', {
        session_id: 'test-session',
        device: dev1,
      });
    });

    // Il dispositivo deve comparire immediatamente nella tabella prima della conclusione HTTP
    await waitFor(() => {
      expect(screen.getByText('192.168.1.10')).toBeDefined();
      expect(screen.getByText('firewall-core')).toBeDefined();
      expect(screen.getByText('Fortinet')).toBeDefined();
      expect(screen.getByText('LIVE')).toBeDefined();
    });

    // 4. Selezione dispositivo durante la scansione attiva
    const checkbox = screen.getAllByRole('checkbox')[1]; // primo checkbox riga
    fireEvent.click(checkbox);

    await waitFor(() => {
      expect(screen.getByText(/Importa Selezionati \(1\)/i)).toBeDefined();
    });

    // 5. Risoluzione della chiamata HTTP di scansione
    act(() => {
      resolveScanPromise!({
        id: 'test-session',
        tenant_id: 't-1',
        subnet: '192.168.1.0/24',
        devices: [dev1],
        total_discovered: 1,
        duration_seconds: 4.25,
        started_at: new Date().toISOString(),
        completed_at: new Date().toISOString(),
      });
    });

    await waitFor(() => {
      // Scansione terminata: banner di progresso rimosso
      expect(screen.queryByText(/Ricognizione LAN Live in corso/i)).toBeNull();
      expect(screen.getByText('4.25s')).toBeDefined();
    });
  });
});
