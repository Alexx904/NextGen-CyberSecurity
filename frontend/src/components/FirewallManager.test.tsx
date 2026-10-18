import { fireEvent, render, screen } from '@testing-library/react';
import { expect, it, vi } from 'vitest';
import { FirewallManager } from './FirewallManager';
import { sshApi } from '../api/ssh';
import type { SSHServerItem } from '../types/ssh';

vi.mock('../api/ssh', () => ({ sshApi: { getFirewallStatus: vi.fn() } }));

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
  vi.mocked(sshApi.getFirewallStatus).mockResolvedValue({ server_id: 'server', chains: {
    INPUT: { name: 'INPUT', default_policy: 'ACCEPT', packets: 0, bytes: 0, rules: [] },
    OUTPUT: { name: 'OUTPUT', default_policy: 'ACCEPT', packets: 0, bytes: 0, rules: [] },
    FORWARD: { name: 'FORWARD', default_policy: 'DROP', packets: 0, bytes: 0, rules: [] },
  }, active_safety_timer: false, safety_timeout_seconds: 0 });
  const server: SSHServerItem = { id: 'server', tenant_id: 'tenant', name: 'Lab', host: '192.0.2.1', port: 22, username: 'test', auth_type: 'PASSWORD', status: 'ONLINE', created_at: '', updated_at: '' };
  render(<FirewallManager server={server} onClose={vi.fn()} />);
  fireEvent.click(await screen.findByRole('button', { name: /aggiungi/i }));
  expect(screen.getByText('Safety Rollback Timer (Anti-Lockout) obbligatorio')).toBeDefined();
  expect(screen.queryByRole('checkbox')).toBeNull();
});
