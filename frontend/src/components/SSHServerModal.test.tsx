import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { describe, expect, it, vi, beforeEach } from 'vitest';
import { SSHServerModal } from './SSHServerModal';
import { sshApi } from '../api/ssh';
import type { SSHServerItem } from '../types/ssh';

vi.mock('../api/ssh', () => ({ sshApi: { updateServer: vi.fn(), createServer: vi.fn() } }));
const pin = 'SHA256:' + 'A'.repeat(43);
const server: SSHServerItem = { id: 'server', tenant_id: 'tenant', name: 'Server', host: '192.0.2.1', port: 22, username: 'user', auth_type: 'PASSWORD', status: 'UNCHECKED', created_at: '', updated_at: '' };
beforeEach(() => vi.clearAllMocks());

describe('SSH server host identity', () => {
  it('allows enrollment of a legacy inventory item with a verified pin', async () => {
    render(<SSHServerModal isOpen server={server} onClose={vi.fn()} onSaved={vi.fn()} />);
    expect(screen.getByText(/Senza fingerprint le connessioni sono bloccate/)).toBeDefined();
    fireEvent.change(screen.getByLabelText('Fingerprint chiave host SSH'), { target: { value: pin } });
    const form = screen.getByText('Salva Modifiche').closest('form')!;
    expect(Array.from(form.querySelectorAll('input')).filter(input => !input.checkValidity()).map(input => [input.outerHTML, input.validationMessage])).toEqual([]);
    fireEvent.submit(form);
    await waitFor(() => expect(sshApi.updateServer).toHaveBeenCalledWith('server', expect.objectContaining({ host_key_fingerprint: pin })));
  });
  it('requires re-enrollment when the endpoint changes', async () => {
    render(<SSHServerModal isOpen server={{ ...server, host_key_fingerprint: pin }} onClose={vi.fn()} onSaved={vi.fn()} />);
    fireEvent.change(screen.getByDisplayValue('192.0.2.1'), { target: { value: '192.0.2.2' } });
    expect((screen.getByLabelText('Fingerprint chiave host SSH') as HTMLInputElement).value).toBe('');
    fireEvent.submit(screen.getByText('Salva Modifiche').closest('form')!);
    await waitFor(() => expect(sshApi.updateServer).toHaveBeenCalledWith('server', expect.objectContaining({ host_key_fingerprint: '' })));
  });
});
