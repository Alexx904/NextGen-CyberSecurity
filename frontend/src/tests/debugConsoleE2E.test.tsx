import React from 'react';
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { render, screen, act } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { DebugProvider } from '../context/DebugContext';
import { DebugFloatingToggle } from '../components/debug/DebugFloatingToggle';
import { DebugConsoleDrawer } from '../components/debug/DebugConsoleDrawer';
import { apiClient } from '../api/client';
import { debugLogger } from '../utils/debugLogger';

// Componente di collaudo che simula una vista dell'applicazione protetta
const MockAppDashboard: React.FC = () => {
  return (
    <div data-testid="mock-dashboard">
      <h1>Dashboard Operativa NextGen</h1>
      <DebugFloatingToggle />
      <DebugConsoleDrawer />
    </div>
  );
};

describe('Debug Console E2E & Integrated Workflow', () => {
  const originalAdapter = apiClient.defaults.adapter;

  beforeEach(() => {
    debugLogger.clear();
    vi.restoreAllMocks();
  });

  afterEach(() => {
    apiClient.defaults.adapter = originalAdapter;
    debugLogger.clear();
  });

  it('executes full end-to-end lifecycle: Axios interception, error badge counting, keyboard shortcut, filtering, and AI Markdown export', async () => {
    // 1. Configurazione Mock Adapter per simulare API Health (200) e API Scan (500)
    apiClient.defaults.adapter = async (config) => {
      if (config.url?.includes('/health')) {
        return {
          data: { status: 'healthy', app: 'nextgen-api', version: '1.0.0' },
          status: 200,
          statusText: 'OK',
          headers: {},
          config,
        };
      }
      if (config.url?.includes('/scans')) {
        const err: any = new Error('Database scan table lock timeout');
        err.config = config;
        err.response = {
          status: 500,
          data: { message: 'Internal Server Error', code: 'DB_LOCK_TIMEOUT' },
        };
        throw err;
      }
      return {
        data: { success: true },
        status: 200,
        statusText: 'OK',
        headers: {},
        config,
      };
    };

    // Mock clipboard per il test di copia
    const writeTextMock = vi.fn().mockResolvedValue(undefined);
    Object.defineProperty(navigator, 'clipboard', {
      value: {
        writeText: writeTextMock,
      },
      configurable: true,
      writable: true,
    });

    // 2. Render dell'applicazione
    render(
      <MemoryRouter>
        <DebugProvider>
          <MockAppDashboard />
        </DebugProvider>
      </MemoryRouter>
    );

    expect(screen.getByTestId('mock-dashboard')).toBeDefined();
    expect(screen.getByRole('button', { name: /Apri Console di Debug/i })).toBeDefined();

    // 3. Esecuzione chiamata API 200 OK (GET /health)
    await act(async () => {
      await apiClient.get('/health');
    });

    // Verifica che debugLogger abbia registrato la richiesta e la risposta
    const initialLogs = debugLogger.getEntries();
    expect(initialLogs.length).toBeGreaterThanOrEqual(2);
    expect(initialLogs.some((e) => e.title.includes('API REQ: GET /health'))).toBe(true);
    expect(initialLogs.some((e) => e.title.includes('API RES: GET /health [200]'))).toBe(true);

    // Non ci sono errori: il badge non deve essere visibile
    expect(screen.queryByTestId('unread-badge')).toBeNull();

    // 4. Simulazione chiamata API 500 (POST /scans)
    await act(async () => {
      try {
        await apiClient.post('/scans', {
          target: '192.168.1.1/24',
          profile: 'full_audit',
          password: 'SensitivePlainPassword123!',
        });
      } catch {
        // Errore 500 atteso
      }
    });

    // Verifica che l'errore sia stato catturato con payload sanificato
    const logsAfterError = debugLogger.getEntries();
    const scanErrorLog = logsAfterError.find((e) => e.title.includes('API ERR: POST /scans'));
    expect(scanErrorLog).toBeDefined();
    expect(scanErrorLog?.status).toBe(500);
    expect((scanErrorLog?.requestData as Record<string, unknown>).password).toBe('[REDACTED]');

    // Il badge contatore errori non letti sul toggle fluttuante deve ora indicare 1
    const badgeAfterApiErr = await screen.findByTestId('unread-badge');
    expect(badgeAfterApiErr).toBeDefined();
    expect(badgeAfterApiErr.textContent).toContain('1');

    // 5. Simulazione errore di runtime JavaScript non gestito
    act(() => {
      window.dispatchEvent(
        new ErrorEvent('error', {
          message: 'Uncaught ReferenceError: nonExistentModule is not defined',
          filename: 'app.js',
          lineno: 120,
          colno: 15,
        })
      );
    });

    // Il badge contatore deve ora indicare 2
    expect(badgeAfterApiErr.textContent).toContain('2');

    // 6. Apertura del drawer tramite scorciatoia globale da tastiera (Ctrl + Shift + D)
    act(() => {
      window.dispatchEvent(
        new KeyboardEvent('keydown', {
          key: 'D',
          ctrlKey: true,
          shiftKey: true,
          bubbles: true,
        })
      );
    });

    // Il drawer è ora aperto: titolo visibile e badge contatore azzerato
    expect(screen.getByText(/Live Debug & Diagnostics Console/i)).toBeDefined();
    expect(screen.queryByTestId('unread-badge')).toBeNull();

    // 7. Test del filtro per livello di severità: solo 'Error'
    const errorFilterBtn = screen.getByTestId('filter-level-error');
    act(() => {
      errorFilterBtn.click();
    });

    // I log info (es. GET /health) non devono comparire, gli errori sì
    expect(screen.queryByText(/GET \/health/i)).toBeNull();
    expect(screen.getByText(/API ERR: POST \/scans/i)).toBeDefined();
    expect(screen.getByText(/RUNTIME ERR: Uncaught ReferenceError/i)).toBeDefined();

    // 8. Test azione "Copia per AI (Markdown)"
    const copyAiBtn = screen.getByTestId('copy-ai-btn');
    await act(async () => {
      copyAiBtn.click();
    });

    expect(writeTextMock).toHaveBeenCalled();
    const copiedText = writeTextMock.mock.calls[0][0];

    // Verifica della struttura conforme a RFC / specs ARCHITECTURE 6.3
    expect(copiedText).toContain('# 🛡️ NextGen CyberSecurity - Diagnostic Session Report');
    expect(copiedText).toContain('## 1. Riepilogo Sessione');
    expect(copiedText).toContain('## 2. Tabella Sinottica degli Errori');
    expect(copiedText).toContain('| `500` |');
    expect(copiedText).toContain('## 3. Schede Dettagliate degli Errori (Deep-Dive)');
    expect(copiedText).toContain('DB_LOCK_TIMEOUT');
    expect(copiedText).toContain('"[REDACTED]"');

    // 9. Test azione "Pulisci Log"
    const clearBtn = screen.getByTestId('clear-logs-btn');
    act(() => {
      clearBtn.click();
    });

    expect(debugLogger.getEntries()).toHaveLength(0);
    expect(screen.getByText(/Nessun log corrisponde ai criteri/i)).toBeDefined();
  });
});
