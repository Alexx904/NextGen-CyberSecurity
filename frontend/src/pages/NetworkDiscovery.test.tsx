import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { render, screen, fireEvent, waitFor, act, within } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { NetworkDiscovery } from './NetworkDiscovery';
import { DebugProvider } from '../context/DebugContext';
import { discoveryApi } from '../api/discovery';
import { apiClient } from '../api/client';
import { debugLogger } from '../utils/debugLogger';

vi.mock('../context/CompanyContext', () => ({
  useCompany: () => ({ selectedCompanyId: null, refreshCompanies: vi.fn() }),
  getAllDescendantGroupIds: (node: { id: string }) => [node.id],
}));

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

  it.each([false, true])('configures only the selected device, already inventoried=%s', async (alreadyImported) => {
    let inventoryReady = alreadyImported;
    vi.mocked(discoveryApi.getLastResults).mockResolvedValue({ id: 'session', tenant_id: 'tenant', subnet: '192.0.2.0/30', total_discovered: 2, duration_seconds: 1, started_at: '', completed_at: '2026-10-18T00:00:00Z', devices: [1, 2].map(n => ({ ip: `192.0.2.${n}`, status: 'up', latency_ms: 1, scanned_at: '' })) });
    vi.mocked(apiClient.get).mockImplementation((async (url: string) => ({ data: { success: true, data: url === '/target-groups' ? [{ id: 'qa', name: 'QA', children: [] }] : url === '/scan-profiles' ? { items: [] } : { items: inventoryReady ? [1, 2].map(n => ({ id: `target-${n}`, group_id: 'qa', value: `192.0.2.${n}`, is_blacklisted: false })) : [], total: inventoryReady ? 2 : 0 } } })) as never);
    vi.spyOn(discoveryApi, 'importTargets').mockImplementation(async () => {
      inventoryReady = true;
      return { group_id: 'qa', total_imported: 1, total_skipped: 0, imported: [{ id: 'target-1', value: '192.0.2.1', type: 'IP', group_id: 'qa', created_at: '' }] };
    });
    vi.spyOn(apiClient, 'post').mockImplementation((async (url: string, body: object) => ({ data: { success: true, data: url === '/scan-profiles' ? { ...body, id: 'profile' } : { id: 'scan' } } })) as never);
    renderComponent();
    fireEvent.click(within(await screen.findByRole('row', { name: /192\.0\.2\.1/ })).getByRole('checkbox'));
    fireEvent.click(screen.getByRole('button', { name: 'Scansiona selezionati (1)' }));
    fireEvent.change(screen.getByLabelText('Cartella Scope (Gerarchia ad Albero)'), { target: { value: 'qa' } });
    fireEvent.click(screen.getByRole('button', { name: 'Importa e configura scansione' }));
    const launch = await screen.findByRole('button', { name: 'Avvia scansione' });
    await waitFor(() => expect(screen.queryByRole('status')).toBeNull());
    fireEvent.click(screen.getByLabelText(/Confermo di possedere/));
    expect((launch as HTMLButtonElement).disabled).toBe(false);
    fireEvent.click(launch);
    await waitFor(() => expect(apiClient.post).toHaveBeenCalledWith('/scans', expect.objectContaining({ target_id: 'target-1', profile_id: 'profile', authorization: expect.objectContaining({ accepted: true, targets: [{ id: 'target-1', value: '192.0.2.1' }] }) })));
    expect(vi.mocked(apiClient.post).mock.calls.filter(([url]) => url === '/scans')).toHaveLength(1);
    expect(discoveryApi.importTargets).toHaveBeenCalledTimes(alreadyImported ? 0 : 1);
  });

  it('renders initial state with default subnet and empty discovery message', async () => {
    renderComponent();

    expect(screen.getByText('Network Discovery')).toBeDefined();
    const input = screen.getByLabelText(/Subnet o Range IP Bersaglio/i) as HTMLInputElement;
    expect(input.value).toBe('192.168.1.0/24');

    await waitFor(() => {
      expect(screen.getByText('Nessuna scansione discovery eseguita')).toBeDefined();
    });
  });

  it('restores a partial persisted snapshot and reloads its completed version', async () => {
    const snapshot = {
      id: 'recovered-session', tenant_id: 'tenant', subnet: '192.168.2.0/24',
      devices: [{ ip: '192.168.2.10', hostname: 'saved-device', latency_ms: 1, status: 'up', scanned_at: '2026-10-18T10:00:00Z' }],
      total_discovered: 1, duration_seconds: 0, started_at: '2026-10-18T10:00:00Z', completed_at: '0001-01-01T00:00:00Z',
    };
    vi.mocked(discoveryApi.getLastResults).mockResolvedValueOnce(snapshot)
      .mockResolvedValueOnce({ ...snapshot, completed_at: '2026-10-18T10:00:10Z', duration_seconds: 10 });
    renderComponent();
    expect(await screen.findByText('saved-device')).toBeDefined();
    expect(screen.getByRole('status').textContent).toContain('Risultati parziali salvati');
    expect(MockEventSource.instances).toHaveLength(0);
    fireEvent.click(screen.getByRole('button', { name: 'Ricarica risultati' }));
    await waitFor(() => expect(screen.queryByRole('status')).toBeNull());
    expect(screen.getByText('saved-device')).toBeDefined();
  });

  it('shows a recovery error and supports retry', async () => {
    vi.mocked(discoveryApi.getLastResults).mockRejectedValueOnce(new Error('Database unavailable'));
    renderComponent();
    expect((await screen.findByRole('alert')).textContent).toContain('Impossibile recuperare');
    fireEvent.click(screen.getByRole('button', { name: 'Ricarica risultati' }));
    await waitFor(() => expect(screen.queryByRole('alert')).toBeNull());
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
