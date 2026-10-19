import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { AlertChannelsCard } from './AlertChannelsCard';
import { alertsApi, type AlertChannel } from '../api/alerts';

const state = vi.hoisted(() => ({ role: 'TenantAdmin' }));
vi.mock('../context/AuthContext', () => ({ useAuth: () => ({ user: { role: state.role } }) }));
vi.mock('../context/CompanyContext', () => ({ useCompany: () => ({ companies: [{ id: 'company-1', name: 'QA company' }] }) }));
vi.mock('../api/alerts', () => ({ alertsApi: { list: vi.fn(), save: vi.fn(), remove: vi.fn(), test: vi.fn() } }));
const channel: AlertChannel = { id: 'channel-1', tenant_id: 'tenant-1', name: 'QA channel', kind: 'webhook', enabled: false, events: ['finding.critical'], created_at: '2026-10-18T12:00:00Z', updated_at: '2026-10-18T12:00:00Z' };

describe('AlertChannelsCard', () => {
  beforeEach(() => {
    vi.clearAllMocks(); state.role = 'TenantAdmin';
    vi.mocked(alertsApi.list).mockResolvedValue([]); vi.mocked(alertsApi.save).mockResolvedValue(channel);
    vi.mocked(alertsApi.test).mockResolvedValue({ accepted: true, code: 'ACCEPTED', message: 'Accepted' });
    vi.mocked(alertsApi.remove).mockResolvedValue(undefined);
  });
  async function newWebhook() {
    render(<AlertChannelsCard />); await screen.findByText('Nessun canale configurato.');
    fireEvent.click(screen.getByRole('button', { name: 'Aggiungi canale' }));
    fireEvent.change(screen.getByLabelText('Nome canale'), { target: { value: 'QA channel' } });
    fireEvent.change(screen.getByLabelText('Endpoint HTTPS', { exact: false }), { target: { value: 'https://example.com/test-only' } });
    fireEvent.change(screen.getByLabelText('Segreto HMAC (almeno 32 caratteri)'), { target: { value: 'synthetic-test-key-32-characters-only' } });
  }
  it('keeps SecOps away from destination configuration and API reads', () => {
    state.role = 'SecOps'; render(<AlertChannelsCard />);
    expect(screen.queryByRole('button', { name: 'Aggiungi canale' })).toBeNull(); expect(alertsApi.list).not.toHaveBeenCalled();
  });
  it('requires explicit test acknowledgement, invalidates it on changes and does not save while testing', async () => {
    await newWebhook();
    const test = screen.getByRole('button', { name: 'Invia messaggio di prova' }) as HTMLButtonElement;
    expect(test.disabled).toBe(true);
    fireEvent.click(screen.getByLabelText('Autorizzo l’invio di un messaggio sintetico di prova al destinatario configurato.'));
    fireEvent.click(test);
    await screen.findByText('Il provider ha accettato il messaggio sintetico. Verifica la ricezione nel canale.');
    expect(alertsApi.test).toHaveBeenCalledTimes(1); expect(alertsApi.save).not.toHaveBeenCalled();
    fireEvent.change(screen.getByLabelText('Endpoint HTTPS', { exact: false }), { target: { value: 'https://example.org/changed' } });
    expect(test.disabled).toBe(true); expect(screen.queryByText(/Il provider ha accettato/)).toBeNull();
  });
  it('saves an initially disabled channel without sending and clears draft secrets on cancel', async () => {
    await newWebhook(); fireEvent.click(screen.getByRole('button', { name: 'Salva canale' }));
    await waitFor(() => expect(alertsApi.save).toHaveBeenCalledWith(expect.objectContaining({ name: 'QA channel', enabled: false, config: expect.objectContaining({ endpoint: 'https://example.com/test-only' }) }), undefined));
    expect(alertsApi.test).not.toHaveBeenCalled(); await screen.findByText(/Canale salvato/);
    fireEvent.click(screen.getByRole('button', { name: 'Aggiungi canale' }));
    expect((screen.getByLabelText('Endpoint HTTPS', { exact: false }) as HTMLInputElement).value).toBe('');
  });
  it('preserves saved secrets when editing metadata and sends the concurrency revision', async () => {
    vi.mocked(alertsApi.list).mockResolvedValue([channel]); render(<AlertChannelsCard />);
    fireEvent.click(await screen.findByRole('button', { name: 'Modifica canale QA channel' }));
    expect(screen.queryByLabelText('Endpoint HTTPS', { exact: false })).toBeNull();
    fireEvent.change(screen.getByLabelText('Ambito canale'), { target: { value: 'company-1' } });
    fireEvent.click(screen.getByRole('button', { name: 'Salva canale' }));
    await waitFor(() => expect(alertsApi.save).toHaveBeenCalledWith(expect.objectContaining({ expected_updated_at: channel.updated_at, company_id: 'company-1' }), channel.id));
    expect(vi.mocked(alertsApi.save).mock.calls[0][0].config).toBeUndefined();
  });
  it('does not remove until the visible confirmation', async () => {
    vi.mocked(alertsApi.list).mockResolvedValue([channel]); render(<AlertChannelsCard />);
    fireEvent.click(await screen.findByRole('button', { name: 'Rimuovi canale QA channel' }));
    expect(alertsApi.remove).not.toHaveBeenCalled(); fireEvent.click(screen.getByRole('button', { name: 'Conferma rimozione canale' }));
    await waitFor(() => expect(alertsApi.remove).toHaveBeenCalledWith('channel-1'));
  });
  it('reports a failed provider test without claiming delivery or saving', async () => {
    vi.mocked(alertsApi.test).mockRejectedValueOnce(new Error('private response text'));
    await newWebhook();
    fireEvent.click(screen.getByLabelText('Autorizzo l’invio di un messaggio sintetico di prova al destinatario configurato.'));
    fireEvent.click(screen.getByRole('button', { name: 'Invia messaggio di prova' }));
    expect((await screen.findByRole('alert')).textContent).toContain('Test non accettato');
    expect(screen.queryByText('private response text')).toBeNull();
    expect(screen.queryByText(/Il provider ha accettato/)).toBeNull(); expect(alertsApi.save).not.toHaveBeenCalled();
  });
  it('includes default STARTTLS port in an SMTP configuration', async () => {
    render(<AlertChannelsCard />); await screen.findByText('Nessun canale configurato.');
    fireEvent.click(screen.getByRole('button', { name: 'Aggiungi canale' }));
    fireEvent.change(screen.getByLabelText('Nome canale'), { target: { value: 'QA SMTP' } });
    fireEvent.change(screen.getByLabelText('Tipo canale'), { target: { value: 'smtp' } });
    fireEvent.change(screen.getByLabelText('Host SMTP pubblico'), { target: { value: 'smtp.example.com' } });
    fireEvent.change(screen.getByLabelText('Mittente email'), { target: { value: 'from@example.test' } });
    fireEvent.change(screen.getByLabelText('Destinatari email (separati da virgola)'), { target: { value: 'to@example.test' } });
    fireEvent.click(screen.getByRole('button', { name: 'Salva canale' }));
    await waitFor(() => expect(alertsApi.save).toHaveBeenCalledWith(expect.objectContaining({ kind: 'smtp', config: expect.objectContaining({ smtp_port: 587 }) }), undefined));
  });
});
