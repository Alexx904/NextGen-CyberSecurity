import React from 'react';
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { render, screen, act, renderHook } from '@testing-library/react';
import { DebugProvider, useDebugConsole } from './DebugContext';
import { debugLogger } from '../utils/debugLogger';

// Componente di test per accedere ai valori del context
const TestConsumer: React.FC = () => {
  const {
    isOpen,
    unreadErrorsCount,
    entries,
    openDrawer,
    closeDrawer,
    toggleDrawer,
    clearLogs,
    setFilterOptions,
  } = useDebugConsole();

  return (
    <div>
      <span data-testid="is-open">{isOpen ? 'OPEN' : 'CLOSED'}</span>
      <span data-testid="unread-count">{unreadErrorsCount}</span>
      <span data-testid="entries-count">{entries.length}</span>
      <button data-testid="btn-open" onClick={openDrawer}>
        Open
      </button>
      <button data-testid="btn-close" onClick={closeDrawer}>
        Close
      </button>
      <button data-testid="btn-toggle" onClick={toggleDrawer}>
        Toggle
      </button>
      <button data-testid="btn-clear" onClick={clearLogs}>
        Clear
      </button>
      <button
        data-testid="btn-filter-errors"
        onClick={() => setFilterOptions({ level: 'error' })}
      >
        Filter Errors
      </button>
    </div>
  );
};

describe('DebugContext & DebugProvider', () => {
  beforeEach(() => {
    debugLogger.clear();
  });

  afterEach(() => {
    debugLogger.clear();
    vi.restoreAllMocks();
  });

  it('throws an error if useDebugConsole is used outside of DebugProvider', () => {
    // Silenzia console.error per il test dell'errore atteso
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

    expect(() => {
      renderHook(() => useDebugConsole());
    }).toThrow("useDebugConsole deve essere utilizzato all'interno di un <DebugProvider>");

    consoleSpy.mockRestore();
  });

  it('initializes with drawer closed, 0 unread errors, and syncs entries with debugLogger', () => {
    render(
      <DebugProvider>
        <TestConsumer />
      </DebugProvider>
    );

    expect(screen.getByTestId('is-open').textContent).toBe('CLOSED');
    expect(screen.getByTestId('unread-count').textContent).toBe('0');
    expect(screen.getByTestId('entries-count').textContent).toBe('0');
  });

  it('handles openDrawer, closeDrawer, and toggleDrawer', () => {
    render(
      <DebugProvider>
        <TestConsumer />
      </DebugProvider>
    );

    act(() => {
      screen.getByTestId('btn-open').click();
    });
    expect(screen.getByTestId('is-open').textContent).toBe('OPEN');

    act(() => {
      screen.getByTestId('btn-close').click();
    });
    expect(screen.getByTestId('is-open').textContent).toBe('CLOSED');

    act(() => {
      screen.getByTestId('btn-toggle').click();
    });
    expect(screen.getByTestId('is-open').textContent).toBe('OPEN');
  });

  it('increments unreadErrorsCount when an error is logged and drawer is closed, and resets on open', () => {
    render(
      <DebugProvider>
        <TestConsumer />
      </DebugProvider>
    );

    // Registra un info log: il contatore non deve incrementare
    act(() => {
      debugLogger.log({
        level: 'info',
        source: 'api',
        title: 'GET /api/v1/targets',
      });
    });
    expect(screen.getByTestId('unread-count').textContent).toBe('0');
    expect(screen.getByTestId('entries-count').textContent).toBe('1');

    // Registra un error log con drawer chiuso
    act(() => {
      debugLogger.log({
        level: 'error',
        source: 'api',
        title: 'POST /api/v1/targets failed',
      });
    });
    expect(screen.getByTestId('unread-count').textContent).toBe('1');
    expect(screen.getByTestId('entries-count').textContent).toBe('2');

    // Registra un secondo error log
    act(() => {
      debugLogger.log({
        level: 'error',
        source: 'runtime',
        title: 'Crash',
      });
    });
    expect(screen.getByTestId('unread-count').textContent).toBe('2');

    // Apertura del drawer: deve azzerare unreadErrorsCount
    act(() => {
      screen.getByTestId('btn-open').click();
    });
    expect(screen.getByTestId('is-open').textContent).toBe('OPEN');
    expect(screen.getByTestId('unread-count').textContent).toBe('0');
  });

  it('clears logs and resets state when clearLogs is called', () => {
    render(
      <DebugProvider>
        <TestConsumer />
      </DebugProvider>
    );

    act(() => {
      debugLogger.log({ level: 'error', source: 'app', title: 'Err 1' });
    });
    expect(screen.getByTestId('unread-count').textContent).toBe('1');

    act(() => {
      screen.getByTestId('btn-clear').click();
    });
    expect(screen.getByTestId('unread-count').textContent).toBe('0');
    expect(screen.getByTestId('entries-count').textContent).toBe('0');
  });

  it('toggles drawer on Ctrl + Shift + D keyboard shortcut', () => {
    render(
      <DebugProvider>
        <TestConsumer />
      </DebugProvider>
    );

    expect(screen.getByTestId('is-open').textContent).toBe('CLOSED');

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

    expect(screen.getByTestId('is-open').textContent).toBe('OPEN');

    act(() => {
      window.dispatchEvent(
        new KeyboardEvent('keydown', {
          key: 'd',
          ctrlKey: true,
          shiftKey: true,
          bubbles: true,
        })
      );
    });

    expect(screen.getByTestId('is-open').textContent).toBe('CLOSED');
  });

  it('captures window error events and logs them to debugLogger', () => {
    render(
      <DebugProvider>
        <TestConsumer />
      </DebugProvider>
    );

    act(() => {
      window.dispatchEvent(
        new ErrorEvent('error', {
          message: 'Uncaught TypeError: window.something is undefined',
          filename: 'main.js',
          lineno: 42,
          colno: 10,
        })
      );
    });

    const entries = debugLogger.getEntries();
    expect(entries.length).toBeGreaterThan(0);
    const last = entries[entries.length - 1];
    expect(last.source).toBe('runtime');
    expect(last.level).toBe('error');
    expect(last.title).toContain('window.something is undefined');
  });

  it('captures window unhandledrejection events and logs them to debugLogger', () => {
    render(
      <DebugProvider>
        <TestConsumer />
      </DebugProvider>
    );

    act(() => {
      const event = new Event('unhandledrejection') as unknown as PromiseRejectionEvent;
      Object.defineProperty(event, 'reason', {
        value: new Error('Async network failure'),
      });
      window.dispatchEvent(event);
    });

    const entries = debugLogger.getEntries();
    expect(entries.length).toBeGreaterThan(0);
    const last = entries[entries.length - 1];
    expect(last.source).toBe('runtime');
    expect(last.level).toBe('error');
    expect(last.error).toBe('Async network failure');
  });
});
