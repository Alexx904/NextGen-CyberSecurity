import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { SSHServers } from './SSHServers';
import { sshApi } from '../api/ssh';
import type { SSHServerItem } from '../types/ssh';

vi.mock('../api/ssh', () => ({ sshApi: { listServers: vi.fn() } }));

describe('Independent SSH customer inventory', () => {
  it('groups and filters clients without CompanyContext', async () => {
    const base = { tenant_id: 't', port: 22, username: 'user', auth_type: 'KEY' as const, status: 'UNCHECKED' as const, created_at: '', updated_at: '' };
    const servers: SSHServerItem[] = [
      { ...base, id: '1', name: 'Alpha server', host: '192.0.2.1', client_name: 'Client A' },
      { ...base, id: '2', name: 'Beta server', host: '192.0.2.2', client_name: 'Client B' },
    ];
    vi.mocked(sshApi.listServers).mockResolvedValue(servers);
    render(<SSHServers />);
    await screen.findByText('Alpha server');
    fireEvent.change(screen.getByLabelText('Filtro cliente SSH'), { target: { value: 'Client B' } });
    expect(screen.queryByText('Alpha server')).toBeNull();
    expect(screen.getByText('Beta server')).toBeDefined();
    expect(screen.getByText('Inventario autonomo, indipendente dall’azienda selezionata')).toBeDefined();
  });
});
