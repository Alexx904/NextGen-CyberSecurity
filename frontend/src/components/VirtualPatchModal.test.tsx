import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { VirtualPatchModal } from './VirtualPatchModal';
import { getVirtualPatches, type VirtualPatchBundle } from '../api/virtualPatching';

vi.mock('../api/virtualPatching', () => ({ getVirtualPatches: vi.fn() }));
const bundle: VirtualPatchBundle = {
  finding_id: 'finding-1', cve_id: 'CVE-2021-41773', catalog_revision: 'test', supported: true,
  product: 'Apache HTTP Server', strategy: 'uri_indicator', operational_impact: 'Possible false positives', covered_cves: ['CVE-2021-41773'],
  reason: 'Template per indicatori noti', limitations: ['Review in staging; no guarantee'], sources: ['https://httpd.apache.org/security/vulnerabilities_24.html'],
  files: [{ engine: 'fail2ban', filename: 'filter.conf', content: '[Definition]\nfailregex = known-indicator\n' }, { engine: 'nginx', filename: 'guard.conf', content: 'return 403;\n' }],
};

describe('VirtualPatchModal', () => {
  beforeEach(() => { vi.restoreAllMocks(); vi.mocked(getVirtualPatches).mockReset(); vi.mocked(getVirtualPatches).mockResolvedValue(bundle); });

  it('shows limitations, exports exact content and changes engine', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    Object.defineProperty(navigator, 'clipboard', { configurable: true, value: { writeText } });
    const createObjectURL = vi.fn().mockReturnValue('blob:test');
    const revokeObjectURL = vi.fn();
    Object.defineProperty(URL, 'createObjectURL', { configurable: true, value: createObjectURL });
    Object.defineProperty(URL, 'revokeObjectURL', { configurable: true, value: revokeObjectURL });
    const download = vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => {});
    render(<VirtualPatchModal findingId="finding-1" onClose={vi.fn()} />);
    expect(await screen.findByText(bundle.reason)).not.toBeNull();
    expect(screen.getByText(bundle.limitations[0])).not.toBeNull();
    fireEvent.click(screen.getByRole('button', { name: 'Copia filter.conf' }));
    await waitFor(() => expect(writeText).toHaveBeenCalledWith(bundle.files[0].content));
    expect(await screen.findByText('Copiato: filter.conf')).not.toBeNull();
    fireEvent.click(screen.getByRole('button', { name: 'Scarica filter.conf' }));
    expect(createObjectURL).toHaveBeenCalledWith(expect.any(Blob));
    expect(download).toHaveBeenCalledTimes(1);
    expect(revokeObjectURL).toHaveBeenCalledWith('blob:test');
    fireEvent.change(screen.getByLabelText('Motore di destinazione'), { target: { value: 'nginx' } });
    expect(screen.getByRole('button', { name: 'Copia guard.conf' })).not.toBeNull();
    expect(screen.queryByRole('button', { name: 'Copia filter.conf' })).toBeNull();
  });

  it('renders unsupported finding without export controls', async () => {
    vi.mocked(getVirtualPatches).mockResolvedValue({ ...bundle, supported: false, files: [], reason: 'Nessun template revisionato' });
    render(<VirtualPatchModal findingId="finding-1" onClose={vi.fn()} />);
    expect(await screen.findByText('Nessun template revisionato')).not.toBeNull();
    expect(screen.queryByLabelText('Motore di destinazione')).toBeNull();
    expect(screen.queryByRole('button', { name: /Scarica/ })).toBeNull();
  });

  it('shows load failure and clipboard failure without false success', async () => {
    vi.mocked(getVirtualPatches).mockRejectedValueOnce(new Error('private backend detail'));
    const view = render(<VirtualPatchModal findingId="finding-1" onClose={vi.fn()} />);
    expect((await screen.findByRole('alert')).textContent).toContain('Impossibile caricare');
    expect(screen.queryByText('private backend detail')).toBeNull();
    view.rerender(<VirtualPatchModal findingId="finding-2" onClose={vi.fn()} />);
    await screen.findByText(bundle.reason);
    Object.defineProperty(navigator, 'clipboard', { configurable: true, value: { writeText: vi.fn().mockRejectedValue(new Error('denied')) } });
    fireEvent.click(screen.getByRole('button', { name: 'Copia filter.conf' }));
    expect((await screen.findByRole('alert')).textContent).toContain('Copia non riuscita');
    expect(screen.queryByText('Copiato: filter.conf')).toBeNull();
  });

  it('ignores an old finding response and restores focus on close', async () => {
    let resolve!: (value: VirtualPatchBundle) => void;
    vi.mocked(getVirtualPatches).mockReturnValueOnce(new Promise(done => { resolve = done; }));
    const opener = document.createElement('button'); document.body.append(opener); opener.focus();
    const onClose = vi.fn();
    const view = render(<VirtualPatchModal findingId="finding-1" onClose={onClose} />);
    expect(screen.getByRole('status').textContent).toContain('Caricamento');
    view.rerender(<VirtualPatchModal findingId="finding-2" onClose={onClose} />);
    await screen.findByText(bundle.reason);
    resolve({ ...bundle, reason: 'OLD RESULT' });
    await waitFor(() => expect(screen.queryByText('OLD RESULT')).toBeNull());
    fireEvent.keyDown(screen.getByRole('dialog'), { key: 'Escape' });
    expect(onClose).toHaveBeenCalledTimes(1);
    view.unmount(); expect(document.activeElement).toBe(opener); opener.remove();
  });
});
