import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, expect, it, vi } from 'vitest';
import { ChokepointRemediationModal } from './ChokepointRemediationModal';
import { attackPathApi } from '../api/attackPath';
import { sshApi } from '../api/ssh';
import type { AttackNode, ChokepointRemediationResult } from '../types/attackPath';
import type { SSHServerItem } from '../types/ssh';

vi.mock('../api/attackPath', () => ({ attackPathApi: { previewRemediation: vi.fn(), remediate: vi.fn() } }));
vi.mock('../api/ssh', () => ({ sshApi: { listServers: vi.fn() } }));
const node: AttackNode = { id: 'service:fixture', type: 'SERVICE', label: 'QA service', ip: '192.0.2.1', port: 8080, protocol: 'tcp' };
const server: SSHServerItem = { id: 'server', tenant_id: 'tenant', name: 'Fixture', host: '192.0.2.1', port: 22, username: 'qa', auth_type: 'PASSWORD', status: 'ONLINE', host_key_fingerprint: 'SHA256:fixture', created_at: '', updated_at: '' };
const preview: ChokepointRemediationResult = { node_id: node.id, server_id: server.id, command: 'sudo iptables -I INPUT 1 -p tcp --dport 8080 -j DROP', safety_timeout_seconds: 60, rule: { chain: 'INPUT', action: 'DROP', protocol: 'tcp', source: '', destination: node.ip!, port: '8080', position: 1 } };
const pending: ChokepointRemediationResult = { ...preview, firewall: { server_id: server.id, active_safety_timer: true, safety_timeout_seconds: 60, chains: { INPUT: { name: 'INPUT', default_policy: 'ACCEPT', packets: 0, bytes: 0, rules: [] }, OUTPUT: { name: 'OUTPUT', default_policy: 'ACCEPT', packets: 0, bytes: 0, rules: [] }, FORWARD: { name: 'FORWARD', default_policy: 'ACCEPT', packets: 0, bytes: 0, rules: [] } } } };
beforeEach(() => {
  vi.resetAllMocks();
  vi.mocked(sshApi.listServers).mockResolvedValue([server]);
  vi.mocked(attackPathApi.previewRemediation).mockResolvedValue(preview);
  vi.mocked(attackPathApi.remediate).mockResolvedValue(pending);
});
const open = () => {
  const props = { node, onClose: vi.fn(), onCheckFirewall: vi.fn() };
  return { ...props, ...render(<ChokepointRemediationModal {...props} />) };
};
async function selectAndPreview() {
  fireEvent.change(await screen.findByLabelText('Server SSH da gestire'), { target: { value: server.id } });
  fireEvent.click(screen.getByRole('button', { name: 'Anteprima regola' }));
  await screen.findByRole('button', { name: 'Applica blocco porta' });
}

it('requires explicit server, backend preview and acknowledgement before the safety handoff', async () => {
  const props = open();
  await screen.findByLabelText('Server SSH da gestire');
  expect((screen.getByRole('button', { name: 'Anteprima regola' }) as HTMLButtonElement).disabled).toBe(true);
  expect(attackPathApi.previewRemediation).not.toHaveBeenCalled();
  await selectAndPreview();
  const apply = screen.getByRole('button', { name: 'Applica blocco porta' }) as HTMLButtonElement;
  expect(apply.disabled).toBe(true);
  expect(screen.getByText(preview.command)).toBeDefined();
  fireEvent.click(screen.getByRole('checkbox'));
  fireEvent.click(apply);
  await waitFor(() => expect(props.onCheckFirewall).toHaveBeenCalledWith(server));
  expect(attackPathApi.remediate).toHaveBeenCalledOnce();
  expect(attackPathApi.remediate).toHaveBeenCalledWith(node.id, server.id);
});

it('excludes other hosts and a TCP management port from server choices', async () => {
  vi.mocked(sshApi.listServers).mockResolvedValue([{ ...server, host: '192.0.2.2' }, { ...server, id: 'management', port: 8080 }]);
  open();
  expect((await screen.findByRole('status')).textContent).toContain('Nessun server SSH compatibile');
  expect(screen.queryByRole('combobox')).toBeNull();
  expect(attackPathApi.previewRemediation).not.toHaveBeenCalled();
});

it('blocks application without a host pin even after an acknowledged preview', async () => {
  vi.mocked(sshApi.listServers).mockResolvedValue([{ ...server, host_key_fingerprint: '' }]);
  open(); await selectAndPreview();
  fireEvent.click(screen.getByRole('checkbox'));
  expect((screen.getByRole('button', { name: 'Applica blocco porta' }) as HTMLButtonElement).disabled).toBe(true);
  expect(attackPathApi.remediate).not.toHaveBeenCalled();
});

it('invalidates preview and acknowledgement when the selected server changes', async () => {
  vi.mocked(sshApi.listServers).mockResolvedValue([server, { ...server, id: 'second', name: 'Second fixture' }]);
  open(); await selectAndPreview(); fireEvent.click(screen.getByRole('checkbox'));
  fireEvent.change(screen.getByLabelText('Server SSH da gestire'), { target: { value: 'second' } });
  expect(screen.queryByRole('button', { name: 'Applica blocco porta' })).toBeNull();
  expect(attackPathApi.remediate).not.toHaveBeenCalled();
});

it('ignores a late preview after the node changed', async () => {
  let resolve!: (value: ChokepointRemediationResult) => void;
  vi.mocked(attackPathApi.previewRemediation).mockReturnValue(new Promise(value => { resolve = value; }));
  const props = open();
  fireEvent.change(await screen.findByLabelText('Server SSH da gestire'), { target: { value: server.id } });
  fireEvent.click(screen.getByRole('button', { name: 'Anteprima regola' }));
  props.rerender(<ChokepointRemediationModal {...props} node={{ ...node, id: 'new-node' }} />);
  await act(async () => resolve(preview));
  expect(screen.queryByText(preview.command)).toBeNull();
  expect(screen.queryByRole('checkbox')).toBeNull();
});

it('blocks automatic retry after an uncertain apply and offers manual firewall verification', async () => {
  vi.mocked(attackPathApi.remediate).mockRejectedValue(new Error('Connection lost'));
  const props = open(); await selectAndPreview();
  fireEvent.click(screen.getByRole('checkbox'));
  fireEvent.click(screen.getByRole('button', { name: 'Applica blocco porta' }));
  await screen.findByRole('alert');
  expect((screen.getByRole('button', { name: 'Applica blocco porta' }) as HTMLButtonElement).disabled).toBe(true);
  fireEvent.click(screen.getByRole('button', { name: 'Verifica stato firewall' }));
  expect(props.onCheckFirewall).toHaveBeenCalledWith(server);
  expect(attackPathApi.remediate).toHaveBeenCalledOnce();
});

it('contains keyboard focus and closes on Escape without a mutation', async () => {
  const props = open();
  await screen.findByLabelText('Server SSH da gestire');
  const dialog = screen.getByRole('dialog');
  fireEvent.keyDown(dialog, { key: 'Tab' });
  expect(document.activeElement).toBe(screen.getByRole('button', { name: 'Chiudi mitigazione' }));
  fireEvent.keyDown(dialog, { key: 'Escape' });
  expect(props.onClose).toHaveBeenCalledOnce();
  expect(attackPathApi.remediate).not.toHaveBeenCalled();
});
