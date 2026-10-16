import { beforeEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { NewScanModal } from './NewScanModal';
import { apiClient } from '../api/client';
import { SCAN_PRESETS } from './scans/scanConfiguration';

vi.mock('../api/client', () => ({ apiClient: { get: vi.fn(), post: vi.fn() } }));
const company = vi.hoisted(() => ({ selectedCompanyId: null as string | null }));
vi.mock('../context/CompanyContext', () => ({
  useCompany: () => company,
  getAllDescendantGroupIds: (node: { id: string; children?: { id: string }[] }) => [node.id, ...(node.children || []).map((child) => child.id)],
}));

const groups = [{ id: 'company', name: 'ACME', children: [{ id: 'department', name: 'IT', children: [] }] }, { id: 'other', name: 'Other company', children: [] }];
const targets = [
  { id: 'one', group_id: 'company', value: '192.168.1.10', is_blacklisted: false },
  { id: 'two', group_id: 'department', value: '192.168.1.20', is_blacklisted: false },
  { id: 'blocked', group_id: 'company', value: '192.168.1.30', is_blacklisted: true },
];
const success = (data: unknown) => ({ data: { success: true, data } });
const profiles = [{ id: 'misleading', name: 'All-in-One 360° Comprehensive Audit', nmap_args: '-sS -sV -O --script=vuln -T4 --open', timeout_seconds: 3600, is_custom: false }];
const get = (url: string) => {
  if (url === '/target-groups') return Promise.resolve(success(groups));
  if (url === '/scan-profiles') return Promise.resolve(success({ items: profiles }));
  return Promise.resolve(success({ items: targets, total: targets.length }));
};

async function open() {
  const props = { isOpen: true, onClose: vi.fn(), onScanLaunched: vi.fn(), onScheduleCreated: vi.fn() };
  const view = render(<NewScanModal {...props} />);
  await waitFor(() => expect(screen.queryByRole('status')).toBeNull());
  return { ...props, ...view };
}
const launch = () => screen.getByRole('button', { name: 'Avvia scansione' }) as HTMLButtonElement;
const edit = () => fireEvent.click(screen.getByRole('button', { name: 'Comando Nmap' }));
const setCommand = (value: string) => fireEvent.change(screen.getByLabelText('Comando / opzioni Nmap'), { target: { value } });

beforeEach(() => {
  vi.resetAllMocks(); company.selectedCompanyId = null;
  vi.mocked(apiClient.get).mockImplementation(get as never);
  vi.mocked(apiClient.post).mockImplementation(((url: string, body: object) => Promise.resolve(success(url === '/scan-profiles' ? { ...body, id: 'created', is_custom: true } : {}))) as never);
});

describe('Minimal Nmap scan workflow', () => {
  it('contains keyboard focus, closes on Escape and restores the previous focus', async () => {
    const trigger = document.createElement('button'); document.body.appendChild(trigger); trigger.focus();
    const props = await open(); const dialog = screen.getByRole('dialog');
    expect(document.activeElement).toBe(dialog);
    fireEvent.keyDown(dialog, { key: 'Tab' }); expect(document.activeElement).toBe(screen.getByRole('button', { name: 'Chiudi' }));
    fireEvent.keyDown(dialog, { key: 'Tab', shiftKey: true }); expect(document.activeElement).toBe(launch());
    fireEvent.keyDown(dialog, { key: 'Escape' }); expect(props.onClose).toHaveBeenCalledOnce();
    props.unmount(); expect(document.activeElement).toBe(trigger); trigger.remove();
  });
  it('has two modes, explicit port profiles and no duplicated composer or legacy catalog', async () => {
    await open();
    expect(screen.getByRole('dialog', { name: 'Nuova scansione Nmap' })).toBeTruthy();
    expect(screen.getByRole('button', { name: 'Profili' })).toBeTruthy();
    expect(screen.queryByText('Componi Multi-Obiettivo')).toBeNull();
    expect(screen.queryByText('Penetration Testing')).toBeNull();
    expect(screen.getByText('Individua le porte aperte tra le 1000 TCP più frequenti e identifica i servizi esposti e le relative versioni.')).toBeTruthy();
    expect(screen.getByText('Durata indicativa per IP: 1 min – 5 min')).toBeTruthy();
    expect(screen.getByText(/2 target disponibili · 1 esclusi/)).toBeTruthy();
  });

  it('creates the exact preview configuration instead of matching a misleading system name', async () => {
    const props = await open(); fireEvent.click(launch());
    await waitFor(() => expect(props.onScanLaunched).toHaveBeenCalledOnce());
    expect(apiClient.post).toHaveBeenCalledWith('/scan-profiles', expect.objectContaining({ nmap_args: SCAN_PRESETS[0].args, name: SCAN_PRESETS[0].name }));
    expect(apiClient.post).toHaveBeenCalledWith('/scans', { group_id: 'company', profile_id: 'created' });
    expect(props.onClose).toHaveBeenCalledOnce();
  });

  it('reuses only an equivalent stored configuration and timeout', async () => {
    vi.mocked(apiClient.get).mockImplementation(((url: string) => url === '/scan-profiles' ? Promise.resolve(success({ items: [{ ...profiles[0], id: 'exact', nmap_args: SCAN_PRESETS[0].args }] })) : get(url)) as never);
    await open(); fireEvent.click(launch());
    await waitFor(() => expect(apiClient.post).toHaveBeenCalledWith('/scans', { group_id: 'company', profile_id: 'exact' }));
    expect(apiClient.post).not.toHaveBeenCalledWith('/scan-profiles', expect.anything());
  });

  it('preloads the selected preset into the editable command', async () => {
    await open(); fireEvent.change(screen.getByLabelText('Tipo di scansione'), { target: { value: 'quick' } }); edit();
    expect((screen.getByLabelText('Comando / opzioni Nmap') as HTMLTextAreaElement).value).toBe(SCAN_PRESETS[1].args);
  });

  it('accepts a nmap prefix and sends only options; preview includes managed output', async () => {
    await open(); edit(); setCommand('nmap -sS -sV -p 22,443 -T3');
    expect(screen.getByTestId('nmap-preview').textContent).toBe('nmap -oX - -sS -sV -p 22,443 -T3 --stats-every 5s -- <target>');
    fireEvent.click(launch());
    await waitFor(() => expect(apiClient.post).toHaveBeenCalledWith('/scan-profiles', expect.objectContaining({ nmap_args: '-sS -sV -p 22,443 -T3' })));
  });

  it.each(['nmap -sS 192.168.1.1', '-sS; whoami', '-oX file.xml', '-sS -sA', '-p80 -p443', '-p70000', '-sA --open', '-sn -sV'])('blocks invalid command %s before any request', async (command) => {
    await open(); edit(); setCommand(command);
    expect(launch().disabled).toBe(true); expect(screen.getByRole('alert')).toBeTruthy();
    fireEvent.click(launch()); expect(apiClient.post).not.toHaveBeenCalled();
  });

  it('gates NSE controls and resets consent after edits', async () => {
    await open(); edit(); setCommand('-sV --script vuln -T3');
    expect(launch().disabled).toBe(true);
    fireEvent.click(screen.getByLabelText(/Confermo di possedere/)); expect(launch().disabled).toBe(false);
    setCommand('-sV --script=vuln -p443 -T3'); expect(launch().disabled).toBe(true);
  });

  it('does not call the server when the group is empty or all targets are blocked', async () => {
    await open(); fireEvent.change(screen.getByLabelText('Azienda / reparto'), { target: { value: 'other' } });
    expect(launch().disabled).toBe(true); expect(apiClient.post).not.toHaveBeenCalled();
  });

  it('never falls back to other companies if the active company is absent', async () => {
    company.selectedCompanyId = 'missing'; await open();
    expect(screen.queryByRole('option', { name: 'ACME' })).toBeNull(); expect(launch().disabled).toBe(true);
  });

  it('loads targets beyond the first page', async () => {
    vi.mocked(apiClient.get).mockImplementation(((url: string) => url.includes('/targets?') ? Promise.resolve(success({ items: url.includes('offset=0') ? targets.slice(0, 1) : targets.slice(1), total: 3 })) : get(url)) as never);
    await open(); expect(apiClient.get).toHaveBeenCalledWith('/targets?limit=200&offset=1');
    expect(screen.getByText(/2 target disponibili · 1 esclusi/)).toBeTruthy();
  });

  it('blocks submission when initial loading fails', async () => {
    vi.mocked(apiClient.get).mockRejectedValue(new Error('Offline'));
    await open(); expect(screen.getByRole('alert').textContent).toBe('Offline'); expect(launch().disabled).toBe(true);
  });

  it('shows server rejection without launching or closing', async () => {
    vi.mocked(apiClient.post).mockResolvedValue({ data: { success: false, error: { message: 'Opzione rifiutata dal server' } } });
    const props = await open(); fireEvent.click(launch());
    await screen.findByText('Opzione rifiutata dal server'); expect(props.onClose).not.toHaveBeenCalled();
    expect(apiClient.post).toHaveBeenCalledTimes(1);
  });

  it('schedules eligible descendants only and retries only failures in the same configuration', async () => {
    let secondFailed = false;
    vi.mocked(apiClient.post).mockImplementation(((url: string, body: { target_id?: string }) => {
      if (url === '/schedules' && body.target_id === 'two' && !secondFailed) { secondFailed = true; return Promise.resolve({ data: { success: false } }); }
      return Promise.resolve(success(url === '/scan-profiles' ? { ...body, id: 'created' } : {}));
    }) as never);
    const props = await open(); fireEvent.click(screen.getByLabelText('Pianifica scansione ricorrente'));
    fireEvent.click(screen.getByRole('button', { name: 'Crea pianificazione' }));
    await screen.findByText(/1 pianificazioni create; 1 non riuscite/);
    fireEvent.click(screen.getByRole('button', { name: 'Crea pianificazione' }));
    await waitFor(() => expect(props.onClose).toHaveBeenCalledOnce());
    const calls = vi.mocked(apiClient.post).mock.calls.filter(([url]) => url === '/schedules');
    expect(calls.map(([, body]) => (body as { target_id: string }).target_id)).toEqual(['one', 'two', 'two']);
    expect(calls[0][1]).toEqual({ target_id: 'one', profile_id: 'created', cron_expression: '0 2 * * *', is_active: true });
  });

  it('supports a named saved profile and rejects invalid timeout', async () => {
    await open(); edit(); fireEvent.click(screen.getByText('Profili salvati e timeout'));
    fireEvent.change(screen.getByLabelText('Timeout per target (minuti)'), { target: { value: '0' } }); expect(launch().disabled).toBe(true);
    fireEvent.change(screen.getByLabelText('Timeout per target (minuti)'), { target: { value: '30' } });
    fireEvent.change(screen.getByLabelText('Nome profilo (facoltativo)'), { target: { value: 'Internal TCP inventory' } });
    fireEvent.click(launch()); await waitFor(() => expect(apiClient.post).toHaveBeenCalledWith('/scan-profiles', expect.objectContaining({ name: 'Internal TCP inventory', timeout_seconds: 1800 })));
  });
});
