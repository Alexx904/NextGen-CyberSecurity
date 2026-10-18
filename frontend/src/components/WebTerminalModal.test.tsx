import { act, render, waitFor } from '@testing-library/react';
import { expect, it, vi } from 'vitest';
import { WebTerminalModal } from './WebTerminalModal';
import { getStreamTicket } from '../api/client';
import { debugLogger } from '../utils/debugLogger';

const terminalFixture = vi.hoisted(() => ({ input: null as null | ((data: string) => void) }));

vi.mock('../api/client', () => ({ getStreamTicket: vi.fn(), getStoredAccessToken: vi.fn() }));
vi.mock('@xterm/xterm', () => ({ Terminal: class {
  open() {} write() {} dispose() {} focus() {} onData(callback: (data: string) => void) { terminalFixture.input = callback; } onResize() {} loadAddon() {}
} }));
vi.mock('@xterm/addon-fit', () => ({ FitAddon: class { fit() {} } }));

it('does not open an SSH WebSocket if the terminal is closed while its ticket is pending', async () => {
  let resolveTicket!: (ticket: string) => void;
  vi.mocked(getStreamTicket).mockReturnValue(new Promise(resolve => { resolveTicket = resolve; }));
  const socket = vi.fn();
  vi.stubGlobal('WebSocket', socket);
  const view = render(<WebTerminalModal docked server={{ id: 'server', tenant_id: 'tenant', name: 'Fixture', host: '192.0.2.1', port: 22, username: 'test', auth_type: 'PASSWORD', status: 'ONLINE', created_at: '', updated_at: '' }} onClose={vi.fn()} />);
  await waitFor(() => expect(getStreamTicket).toHaveBeenCalled());
  view.unmount();
  await act(async () => { resolveTicket('fixture-ticket'); });
  expect(socket).not.toHaveBeenCalled();
  vi.unstubAllGlobals();
});

it('logs WS state and frame lengths without terminal payloads or ticket secrets', async () => {
  vi.mocked(getStreamTicket).mockResolvedValue('fixture-private-ticket');
  const log = vi.spyOn(debugLogger, 'logWSEvent');
  let ws!: { onopen: () => void; onmessage: (event: { data: string }) => void; readyState: number };
  vi.stubGlobal('WebSocket', class {
    static OPEN = 1;
    readyState = 1;
    onopen = () => {};
    onmessage = (_event: { data: string }) => {};
    constructor() { ws = this; }
    send() {} close() {}
  });
  vi.stubGlobal('ResizeObserver', class { observe() {} disconnect() {} });
  const view = render(<WebTerminalModal docked server={{ id: 'server', tenant_id: 'tenant', name: 'Fixture', host: '192.0.2.1', port: 22, username: 'test', auth_type: 'PASSWORD', status: 'ONLINE', created_at: '', updated_at: '' }} onClose={vi.fn()} />);
  await waitFor(() => expect(terminalFixture.input).not.toBeNull());
  act(() => { ws.onopen(); terminalFixture.input!('terminal-sensitive-input'); ws.onmessage({ data: 'terminal-sensitive-output' }); });
  expect(log).toHaveBeenCalledWith(expect.objectContaining({ eventType: 'open' }));
  expect(log).toHaveBeenCalledWith(expect.objectContaining({ direction: 'out', eventType: 'frame', data: { characters: 24 } }));
  expect(log).toHaveBeenCalledWith(expect.objectContaining({ direction: 'in', eventType: 'frame', data: { characters: 25 } }));
  const logged = JSON.stringify(log.mock.calls);
  expect(logged).not.toContain('terminal-sensitive');
  expect(logged).not.toContain('fixture-private-ticket');
  view.unmount();
  log.mockRestore();
  vi.unstubAllGlobals();
});
