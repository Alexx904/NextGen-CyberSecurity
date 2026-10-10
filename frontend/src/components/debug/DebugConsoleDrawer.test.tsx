import React from 'react';
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { render, screen, act, cleanup } from '@testing-library/react';
import { DebugProvider, useDebugConsole } from '../../context/DebugContext';
import { DebugConsoleDrawer } from './DebugConsoleDrawer';
import { DebugFloatingToggle } from './DebugFloatingToggle';
import { debugLogger } from '../../utils/debugLogger';

// Helper component per forzare l'apertura o loggare entry nei test
const DrawerTestHelper: React.FC = () => {
  const { openDrawer } = useDebugConsole();
  return (
    <div>
      <button data-testid="open-helper" onClick={openDrawer}>
        Open
      </button>
      <DebugFloatingToggle />
      <DebugConsoleDrawer />
    </div>
  );
};

describe('DebugConsoleDrawer & DebugFloatingToggle UI', () => {
  beforeEach(() => {
    debugLogger.clear();
    vi.restoreAllMocks();
  });

  afterEach(() => {
    cleanup();
    debugLogger.clear();
  });

  it('renders floating toggle when closed, and hides toggle when drawer opens', () => {
    render(
      <DebugProvider>
        <DrawerTestHelper />
      </DebugProvider>
    );

    // Con drawer chiuso, il floating toggle è visibile
    expect(screen.getByRole('button', { name: /Apri Console di Debug/i })).toBeDefined();

    // Apri drawer
    act(() => {
      screen.getByTestId('open-helper').click();
    });

    // Con drawer aperto, il toggle si nasconde e appare la testata del drawer
    expect(screen.queryByRole('button', { name: /Apri Console di Debug/i })).toBeNull();
    expect(screen.getByText(/Live Debug & Diagnostics Console/i)).toBeDefined();
  });

  it('displays unread error count badge on floating toggle', () => {
    render(
      <DebugProvider>
        <DrawerTestHelper />
      </DebugProvider>
    );

    expect(screen.queryByTestId('unread-badge')).toBeNull();

    act(() => {
      debugLogger.log({
        level: 'error',
        source: 'api',
        title: 'Network Fail',
      });
    });

    const badge = screen.getByTestId('unread-badge');
    expect(badge).toBeDefined();
    expect(badge.textContent).toContain('1');
  });

  it('renders log list and allows filtering by level and search term', () => {
    render(
      <DebugProvider>
        <DrawerTestHelper />
      </DebugProvider>
    );

    // Popola il buffer con 2 log
    act(() => {
      debugLogger.log({
        level: 'info',
        source: 'api',
        title: 'GET /api/v1/targets',
        url: '/api/v1/targets',
      });
      debugLogger.log({
        level: 'error',
        source: 'api',
        title: 'POST /api/v1/scans failed',
        url: '/api/v1/scans',
        error: 'Critical Scan Error',
      });
    });

    // Apri drawer
    act(() => {
      screen.getByTestId('open-helper').click();
    });

    expect(screen.getByText('GET /api/v1/targets')).toBeDefined();
    expect(screen.getByText('POST /api/v1/scans failed')).toBeDefined();

    // Filtra per Error
    const errorFilterBtn = screen.getByRole('button', { name: /Error/i });
    act(() => {
      errorFilterBtn.click();
    });

    expect(screen.queryByText('GET /api/v1/targets')).toBeNull();
    expect(screen.getByText('POST /api/v1/scans failed')).toBeDefined();

    // Ricerca testuale
    const searchInput = screen.getByPlaceholderText(/Cerca per titolo/i);
    act(() => {
      searchInput.dispatchEvent(
        new Event('input', { bubbles: true })
      );
    });
  });

  it('expands log entry to show details on click', () => {
    render(
      <DebugProvider>
        <DrawerTestHelper />
      </DebugProvider>
    );

    let loggedEntryId = '';
    act(() => {
      const e = debugLogger.log({
        level: 'error',
        source: 'api',
        title: 'Database connection failed',
        url: '/api/v1/health',
        requestData: { timeout: 5000 },
        responseData: { status: 'unhealthy' },
        stack: 'Error at db.go:10',
      });
      loggedEntryId = e.id;
    });

    act(() => {
      screen.getByTestId('open-helper').click();
    });

    const logRow = screen.getByTestId(`log-entry-${loggedEntryId}`);
    expect(logRow).toBeDefined();

    // Clicca per espandere
    act(() => {
      logRow.click();
    });

    expect(screen.getByText(/URL Endpoint:/i)).toBeDefined();
    expect(screen.getByText('/api/v1/health')).toBeDefined();
    expect(screen.getByText(/Request Data \(Sanificato\):/i)).toBeDefined();
    expect(screen.getByText(/Response Data \(Sanificato\):/i)).toBeDefined();
    expect(screen.getByText(/Stack Trace:/i)).toBeDefined();
    expect(screen.getByText(/Error at db.go:10/i)).toBeDefined();
  });

  it('handles Copy for AI action and triggers clipboard write', async () => {
    const writeTextMock = vi.fn().mockResolvedValue(undefined);
    Object.defineProperty(navigator, 'clipboard', {
      value: {
        writeText: writeTextMock,
      },
      configurable: true,
      writable: true,
    });

    render(
      <DebugProvider>
        <DrawerTestHelper />
      </DebugProvider>
    );

    act(() => {
      debugLogger.log({
        level: 'error',
        source: 'runtime',
        title: 'Uncaught TypeError',
      });
    });

    act(() => {
      screen.getByTestId('open-helper').click();
    });

    const copyBtn = screen.getByTestId('copy-ai-btn');
    await act(async () => {
      copyBtn.click();
    });

    expect(writeTextMock).toHaveBeenCalled();
    expect(screen.getByText(/Copiato per AI!/i)).toBeDefined();
  });

  it('clears logs when clear button is clicked', () => {
    render(
      <DebugProvider>
        <DrawerTestHelper />
      </DebugProvider>
    );

    act(() => {
      debugLogger.log({ level: 'info', source: 'app', title: 'Old log' });
    });

    act(() => {
      screen.getByTestId('open-helper').click();
    });

    expect(screen.getByText('Old log')).toBeDefined();

    const clearBtn = screen.getByTestId('clear-logs-btn');
    act(() => {
      clearBtn.click();
    });

    expect(screen.queryByText('Old log')).toBeNull();
    expect(screen.getByText(/Nessun log corrisponde ai criteri/i)).toBeDefined();
  });
});
