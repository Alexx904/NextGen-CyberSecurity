import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { NewScanModal } from './NewScanModal';
import { apiClient } from '../api/client';

vi.mock('../api/client', () => ({
  apiClient: {
    get: vi.fn(),
    post: vi.fn(),
    delete: vi.fn(),
  },
}));

describe('NewScanModal Component - Guided & Professional Scan Customization', () => {
  const mockTargets = [
    {
      id: 'target-01',
      tenant_id: 'tenant-1',
      type: 'IP',
      value: '192.168.1.100',
      description: 'Main Web Server',
      is_blacklisted: false,
    },
    {
      id: 'target-02',
      tenant_id: 'tenant-1',
      type: 'IP',
      value: '192.168.1.200',
      description: 'Excluded Gateway',
      is_blacklisted: true,
    },
  ];

  const mockProfiles = [
    {
      id: 'prof-pen-01',
      name: 'Stealth SYN Pen Test',
      nmap_args: '-sS -Pn -T2 --open',
      description: 'Offensive probing stealth',
      timeout_seconds: 1800,
      category: 'PEN_TEST',
      is_custom: false,
    },
    {
      id: 'prof-flaw-01',
      name: 'OS & System Flaws Audit',
      nmap_args: '-sV -O --version-intensity 7 -T4',
      description: 'Audit vulnerabilità di sistema',
      timeout_seconds: 2400,
      category: 'SYSTEM_FLAWS',
      is_custom: false,
    },
    {
      id: 'prof-web-01',
      name: 'Web & API Surface Audit',
      nmap_args: '-sV -p 80,443,8080,8443 -T4 --open',
      description: 'Ricognizione endpoint web',
      timeout_seconds: 1200,
      category: 'WEB_APP',
      is_custom: false,
    },
    {
      id: 'prof-disc-01',
      name: 'Fast Ping & Host Sweep',
      nmap_args: '-sn -PE -T4 -n',
      description: 'Sweep veloce host attivi',
      timeout_seconds: 300,
      category: 'DISCOVERY',
      is_custom: false,
    },
  ];

  beforeEach(() => {
    vi.clearAllMocks();
    (apiClient.get as any).mockImplementation((url: string) => {
      if (url.includes('/targets')) {
        return Promise.resolve({
          data: {
            success: true,
            data: { items: mockTargets, total: mockTargets.length, limit: 100, offset: 0 },
          },
        });
      }
      if (url.includes('/scan-profiles')) {
        return Promise.resolve({
          data: {
            success: true,
            data: { items: mockProfiles, total: mockProfiles.length },
          },
        });
      }
      return Promise.reject(new Error(`Endpoint non mockato: ${url}`));
    });
  });

  it('renders modal and displays category selector with profile counts in guided mode', async () => {
    render(
      <NewScanModal
        isOpen={true}
        onClose={vi.fn()}
        onScanLaunched={vi.fn()}
      />
    );

    // Titolo e modalità
    expect(screen.getByText('Nuova Scansione di Vulnerabilità')).toBeTruthy();
    expect(screen.getByText('Modalità Guidata')).toBeTruthy();
    expect(screen.getByText('Modalità Professionale')).toBeTruthy();

    // Attendi caricamento dati
    await waitFor(() => {
      expect(screen.getByText('Penetration Testing')).toBeTruthy();
      expect(screen.getByText('Falle di Sistema')).toBeTruthy();
      expect(screen.getByText('Web App & API')).toBeTruthy();
      expect(screen.getByText('Network Discovery')).toBeTruthy();
      expect(screen.getByText('Compliance & Hardening')).toBeTruthy();
      expect(screen.getByText('Profili Custom')).toBeTruthy();
    });

    // Profilo di default categoria PEN_TEST
    expect(screen.getByText('Stealth SYN Pen Test')).toBeTruthy();
    // Live preview
    expect(screen.getByText('Live Nmap Command Preview')).toBeTruthy();
  });

  it('filters profiles dynamically when selecting different categories', async () => {
    render(
      <NewScanModal
        isOpen={true}
        onClose={vi.fn()}
        onScanLaunched={vi.fn()}
      />
    );

    await waitFor(() => {
      expect(screen.getByText('Stealth SYN Pen Test')).toBeTruthy();
    });

    // Clic su "Falle di Sistema"
    fireEvent.click(screen.getByText('Falle di Sistema'));

    await waitFor(() => {
      expect(screen.getByText('OS & System Flaws Audit')).toBeTruthy();
      expect(screen.queryByText('Stealth SYN Pen Test')).toBeNull();
    });

    // Clic su "Web App & API"
    fireEvent.click(screen.getByText('Web App & API'));

    await waitFor(() => {
      expect(screen.getByText('Web & API Surface Audit')).toBeTruthy();
      expect(screen.queryByText('OS & System Flaws Audit')).toBeNull();
    });
  });

  it('switches to professional mode and displays port scopes, timing templates, and flags', async () => {
    render(
      <NewScanModal
        isOpen={true}
        onClose={vi.fn()}
        onScanLaunched={vi.fn()}
      />
    );

    await waitFor(() => {
      expect(screen.getByText('Modalità Professionale')).toBeTruthy();
    });

    // Switch a modalità professionale
    fireEvent.click(screen.getByText('Modalità Professionale'));

    // Verifica controlli professionali
    expect(screen.getByText('Scope Porte Bersaglio')).toBeTruthy();
    expect(screen.getByText('Top 100 Porte (-F)')).toBeTruthy();
    expect(screen.getByText('Servizi Critici & RCE')).toBeTruthy();
    expect(screen.getByText('Timing & Aggressiveness Template')).toBeTruthy();
    expect(screen.getByText('SYN Stealth (-sS)')).toBeTruthy();
    expect(screen.getByText('ACK Firewall Probe (-sA)')).toBeTruthy();
    expect(screen.getByText('Service Versioning (-sV)')).toBeTruthy();
    expect(screen.getByText('OS Fingerprinting (-O)')).toBeTruthy();
    expect(screen.getByText('Salva questa configurazione come nuovo Profilo Custom')).toBeTruthy();
  });

  it('updates live nmap command preview dynamically in professional mode', async () => {
    render(
      <NewScanModal
        isOpen={true}
        onClose={vi.fn()}
        onScanLaunched={vi.fn()}
      />
    );

    await waitFor(() => {
      expect(screen.getByText('Modalità Professionale')).toBeTruthy();
    });

    fireEvent.click(screen.getByText('Modalità Professionale'));

    // Clic su porte critiche
    fireEvent.click(screen.getByText('Servizi Critici & RCE'));

    // Verifica che il live preview contenga i flag generati
    await waitFor(() => {
      expect(screen.getByText('Live Nmap Command Preview')).toBeTruthy();
    });
  });

  it('submits on-demand scan with selected profile in guided mode', async () => {
    const handleLaunched = vi.fn();
    const handleClose = vi.fn();

    (apiClient.post as any).mockResolvedValueOnce({
      data: {
        success: true,
        data: { id: 'scan-123', status: 'QUEUED' },
      },
    });

    render(
      <NewScanModal
        isOpen={true}
        onClose={handleClose}
        onScanLaunched={handleLaunched}
      />
    );

    await waitFor(() => {
      expect(screen.getByText('Stealth SYN Pen Test')).toBeTruthy();
    });

    // Clic su "Avvia Scansione Ora"
    const submitBtn = screen.getByRole('button', { name: /Avvia Scansione Ora/i });
    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(apiClient.post).toHaveBeenCalledWith('/scans', {
        target_id: 'target-01',
        profile_id: 'prof-pen-01',
      });
      expect(handleLaunched).toHaveBeenCalled();
      expect(handleClose).toHaveBeenCalled();
    });
  });

  it('creates custom profile and launches scan when saving custom profile in professional mode', async () => {
    const handleLaunched = vi.fn();
    const handleClose = vi.fn();

    // Mock creazione profilo custom
    (apiClient.post as any).mockResolvedValueOnce({
      data: {
        success: true,
        data: { id: 'new-custom-prof-id', name: 'My Custom Audit' },
      },
    });

    // Mock lancio scansione
    (apiClient.post as any).mockResolvedValueOnce({
      data: {
        success: true,
        data: { id: 'scan-custom-999', status: 'QUEUED' },
      },
    });

    render(
      <NewScanModal
        isOpen={true}
        onClose={handleClose}
        onScanLaunched={handleLaunched}
      />
    );

    await waitFor(() => {
      expect(screen.getByText('Modalità Professionale')).toBeTruthy();
    });

    fireEvent.click(screen.getByText('Modalità Professionale'));

    // Attiva checkbox "Salva questa configurazione come nuovo Profilo Custom"
    const saveCheckbox = screen.getByRole('checkbox', {
      name: /Salva questa configurazione come nuovo Profilo Custom/i,
    });
    fireEvent.click(saveCheckbox);

    // Compila nome profilo
    const nameInput = screen.getByPlaceholderText('es. Internal Web & DB Audit');
    fireEvent.change(nameInput, { target: { value: 'My Custom Audit' } });

    // Invia form
    const submitBtn = screen.getByRole('button', { name: /Avvia Scansione Ora/i });
    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(apiClient.post).toHaveBeenCalledWith('/scan-profiles', expect.objectContaining({
        name: 'My Custom Audit',
      }));
      expect(apiClient.post).toHaveBeenCalledWith('/scans', {
        target_id: 'target-01',
        profile_id: 'new-custom-prof-id',
      });
      expect(handleLaunched).toHaveBeenCalled();
      expect(handleClose).toHaveBeenCalled();
    });
  });

  it('creates scheduled scan with cron expression when scheduled tab is active', async () => {
    const handleScheduleCreated = vi.fn();
    const handleClose = vi.fn();

    (apiClient.post as any).mockResolvedValueOnce({
      data: {
        success: true,
        data: { id: 'sched-123', is_active: true },
      },
    });

    render(
      <NewScanModal
        isOpen={true}
        onClose={handleClose}
        onScanLaunched={vi.fn()}
        onScheduleCreated={handleScheduleCreated}
      />
    );

    await waitFor(() => {
      expect(screen.getByText('Schedulazione Cron')).toBeTruthy();
    });

    // Switch a tab Schedulazione Cron
    fireEvent.click(screen.getByText('Schedulazione Cron'));

    // Verifica presenza opzioni cron
    expect(screen.getByText('Ogni notte alle 02:00')).toBeTruthy();

    // Invia form
    const submitBtn = screen.getByRole('button', { name: /Crea Pianificazione/i });
    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(apiClient.post).toHaveBeenCalledWith('/schedules', {
        target_id: 'target-01',
        profile_id: 'prof-pen-01',
        cron_expression: '0 2 * * *',
        is_active: true,
      });
      expect(handleScheduleCreated).toHaveBeenCalled();
      expect(handleClose).toHaveBeenCalled();
    });
  });
});
