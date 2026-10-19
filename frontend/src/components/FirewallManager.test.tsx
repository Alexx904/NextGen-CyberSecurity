import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, expect, it, vi } from 'vitest';
import { FirewallManager } from './FirewallManager';
import { sshApi } from '../api/ssh';
import type { FirewallStatus, SSHServerItem } from '../types/ssh';

vi.mock('../api/ssh', () => ({ sshApi: {
  getFirewallStatus: vi.fn(), setFirewallPolicy: vi.fn(), moveFirewallRule: vi.fn(), replaceFirewallRule: vi.fn(),
} }));
vi.mock('./WebTerminalModal', () => ({ WebTerminalModal: ({ docked, onClose }: { docked?: boolean; onClose: () => void }) => <section aria-label="Docked terminal fixture">{docked ? 'Docked' : 'Overlay'}<button onClick={onClose}>Close terminal fixture</button></section> }));

afterEach(() => vi.unstubAllGlobals());

it('shows the actionable API error for a firewall HTTP 502', async () => {
  const message = 'Accesso firewall negato: configurare sudo non interattivo.';
  vi.mocked(sshApi.getFirewallStatus).mockRejectedValue({
    message: 'Request failed with status code 502',
    response: { data: { error: { message } } },
  });
  const server: SSHServerItem = { id: 'server', tenant_id: 'tenant', name: 'Debian', host: '192.0.2.1', port: 22, username: 'alex', auth_type: 'PASSWORD', status: 'ONLINE', created_at: '', updated_at: '' };
  render(<FirewallManager server={server} onClose={vi.fn()} />);
  expect(await screen.findByText(message)).toBeDefined();
  expect(screen.queryByText('Request failed with status code 502')).toBeNull();
});

it('keeps anti-lockout protection mandatory in the rule form', async () => {
  vi.mocked(sshApi.getFirewallStatus).mockResolvedValue({ server_id: 'server', revision: 'a'.repeat(64), chains: {
    INPUT: { name: 'INPUT', default_policy: 'ACCEPT', packets: 0, bytes: 0, rules: [] },
    OUTPUT: { name: 'OUTPUT', default_policy: 'ACCEPT', packets: 0, bytes: 0, rules: [] },
    FORWARD: { name: 'FORWARD', default_policy: 'DROP', packets: 0, bytes: 0, rules: [] },
  }, active_safety_timer: false, safety_timeout_seconds: 0 });
  const server: SSHServerItem = { id: 'server', tenant_id: 'tenant', name: 'Lab', host: '192.0.2.1', port: 22, username: 'test', auth_type: 'PASSWORD', status: 'ONLINE', created_at: '', updated_at: '' };
  render(<FirewallManager server={server} onClose={vi.fn()} />);
  const addButton = await screen.findByRole('button', { name: /aggiungi/i });
  await waitFor(() => expect((addButton as HTMLButtonElement).disabled).toBe(false));
  fireEvent.click(addButton);
  expect(screen.getByText('Safety Rollback Timer (Anti-Lockout) obbligatorio')).toBeDefined();
  expect(screen.queryByRole('checkbox')).toBeNull();
});

const labServer: SSHServerItem = { id: 'server', tenant_id: 'tenant', name: 'Lab', host: '192.0.2.1', port: 22, username: 'test', auth_type: 'PASSWORD', status: 'ONLINE', created_at: '', updated_at: '' };
const editableStatus = (): FirewallStatus => ({
  server_id: 'server', revision: 'a'.repeat(64), active_safety_timer: false, safety_timeout_seconds: 0,
  chains: {
    INPUT: { name: 'INPUT', default_policy: 'ACCEPT', packets: 0, bytes: 0, rules: [
      { line_number: 1, packets: 0, bytes: 0, target: 'ACCEPT', protocol: 'tcp', opt: '--', in_interface: 'eth0', out_interface: '*', source: '192.0.2.0/24', destination: '0.0.0.0/0', options: 'tcp dpt:80 /* fixture */' },
      { line_number: 2, packets: 0, bytes: 0, target: 'DROP', protocol: 'tcp', opt: '--', in_interface: '*', out_interface: '*', source: '0.0.0.0/0', destination: '0.0.0.0/0', options: 'tcp dpt:443' },
    ] },
    OUTPUT: { name: 'OUTPUT', default_policy: 'ACCEPT', packets: 0, bytes: 0, rules: [] },
    FORWARD: { name: 'FORWARD', default_policy: 'ACCEPT', packets: 0, bytes: 0, rules: [] },
  },
});
const pendingStatus = (): FirewallStatus => ({ ...editableStatus(), active_safety_timer: true, safety_timeout_seconds: 60, safety_timer_expires_at: new Date(Date.now() + 60000).toISOString() });

it('sends revision-checked reordering and disables mutations while confirmation is pending', async () => {
  vi.mocked(sshApi.getFirewallStatus).mockResolvedValue(editableStatus());
  vi.mocked(sshApi.moveFirewallRule).mockResolvedValue(pendingStatus());
  render(<FirewallManager server={labServer} onClose={vi.fn()} />);
  const down = await screen.findByRole('button', { name: 'Sposta regola 1 giù' });
  fireEvent.click(down);
  await waitFor(() => expect(sshApi.moveFirewallRule).toHaveBeenCalledWith('server', 'INPUT', 1, 2, 'a'.repeat(64)));
  await waitFor(() => expect((screen.getByRole('button', { name: /aggiungi/i }) as HTMLButtonElement).disabled).toBe(true));
});

it('replaces the selected rule with the visible form and preserves its interface fields', async () => {
  vi.mocked(sshApi.getFirewallStatus).mockResolvedValue(editableStatus());
  vi.mocked(sshApi.replaceFirewallRule).mockResolvedValue(pendingStatus());
  render(<FirewallManager server={labServer} onClose={vi.fn()} />);
  fireEvent.click(await screen.findByRole('button', { name: 'Sostituisci regola 1' }));
  expect(screen.getByRole('note').textContent).toContain('opzioni avanzate');
  expect((screen.getByLabelText('Porta o Range') as HTMLInputElement).value).toBe('80');
  expect((screen.getByLabelText('Interfaccia ingresso') as HTMLInputElement).value).toBe('eth0');
  fireEvent.change(screen.getByLabelText('Azione Target'), { target: { value: 'DROP' } });
  fireEvent.click(screen.getByRole('button', { name: 'Sostituisci Regola' }));
  await waitFor(() => expect(sshApi.replaceFirewallRule).toHaveBeenCalledWith('server', 1, expect.objectContaining({ action: 'DROP', port: '80', in_interface: 'eth0', expected_revision: 'a'.repeat(64), with_safety_timer: true })));
});

it('opens a docked terminal only on request and closes it without closing the firewall', async () => {
  vi.mocked(sshApi.getFirewallStatus).mockResolvedValue(editableStatus());
  render(<FirewallManager server={labServer} onClose={vi.fn()} />);
  expect(screen.queryByRole('region', { name: 'Docked terminal fixture' })).toBeNull();
  fireEvent.click(await screen.findByRole('button', { name: 'Apri terminale' }));
  expect(await screen.findByText('Docked')).toBeDefined();
  fireEvent.click(screen.getByRole('button', { name: 'Close terminal fixture' }));
  expect(screen.queryByText('Docked')).toBeNull();
  expect(screen.getByRole('button', { name: 'Apri terminale' })).toBeDefined();
});

it('requires confirmation and sends the current revision for a policy change', async () => {
  vi.mocked(sshApi.getFirewallStatus).mockResolvedValue(editableStatus());
  vi.mocked(sshApi.setFirewallPolicy).mockResolvedValue(pendingStatus());
  // happy-dom does not implement the browser confirmation dialog.
  const confirm = vi.fn(() => true);
  vi.stubGlobal('confirm', confirm);
  render(<FirewallManager server={labServer} onClose={vi.fn()} />);
  const policy = await screen.findByLabelText('Nuova policy INPUT');
  await waitFor(() => expect((policy as HTMLSelectElement).disabled).toBe(false));
  fireEvent.change(policy, { target: { value: 'DROP' } });
  fireEvent.click(screen.getByRole('button', { name: 'Applica policy' }));
  await waitFor(() => expect(sshApi.setFirewallPolicy).toHaveBeenCalledWith('server', 'INPUT', 'DROP', 'a'.repeat(64)));
  expect(confirm).toHaveBeenCalledWith(expect.stringContaining('DROP'));
});

it('clears the transport port when replacing a TCP rule with ICMP', async () => {
  vi.mocked(sshApi.getFirewallStatus).mockResolvedValue(editableStatus());
  vi.mocked(sshApi.replaceFirewallRule).mockResolvedValue(pendingStatus());
  render(<FirewallManager server={labServer} onClose={vi.fn()} />);
  fireEvent.click(await screen.findByRole('button', { name: 'Sostituisci regola 1' }));
  fireEvent.change(screen.getByLabelText('Protocollo'), { target: { value: 'icmp' } });
  expect((screen.getByLabelText('Porta o Range') as HTMLInputElement).value).toBe('');
  fireEvent.click(screen.getByRole('button', { name: 'Sostituisci Regola' }));
  await waitFor(() => expect(sshApi.replaceFirewallRule).toHaveBeenCalledWith('server', 1, expect.objectContaining({ protocol: 'icmp', port: '' })));
});

it('keeps an expired remote transaction visible without inventing another countdown', async () => {
  vi.mocked(sshApi.getFirewallStatus).mockResolvedValue({ server_id: 'server', chains: {
    INPUT: { name: 'INPUT', default_policy: 'ACCEPT', packets: 0, bytes: 0, rules: [] },
    OUTPUT: { name: 'OUTPUT', default_policy: 'ACCEPT', packets: 0, bytes: 0, rules: [] },
    FORWARD: { name: 'FORWARD', default_policy: 'DROP', packets: 0, bytes: 0, rules: [] },
  }, active_safety_timer: true, safety_timeout_seconds: 60, safety_timer_expires_at: new Date(Date.now() - 60000).toISOString() });
  const server: SSHServerItem = { id: 'server', tenant_id: 'tenant', name: 'Lab', host: '192.0.2.1', port: 22, username: 'test', auth_type: 'PASSWORD', status: 'ONLINE', created_at: '', updated_at: '' };
  render(<FirewallManager server={server} onClose={vi.fn()} />);
  expect(await screen.findByText('Scadenza raggiunta: ripristino da verificare.')).toBeDefined();
  expect(screen.queryByText('60s rimanenti')).toBeNull();
  expect((screen.getByRole('button', { name: 'Conferma Configurazione' }) as HTMLButtonElement).disabled).toBe(true);
});
