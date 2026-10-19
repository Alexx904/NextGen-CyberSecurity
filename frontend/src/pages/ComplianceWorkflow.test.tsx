import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { ComplianceWorkflow, assetCandidates } from './ComplianceWorkflow';
import { complianceApi } from '../api/compliance';
import type { ComplianceFinding, ComplianceIssue, ComplianceSnapshot } from '../types/compliance';

vi.mock('../api/compliance', () => ({ complianceApi: { asset: vi.fn(), observe: vi.fn(), command: vi.fn() } }));
const finding: ComplianceFinding = { id: 'finding', scan_id: 'scan', host_ip: '192.0.2.1', port: 23, protocol: 'tcp', group_id: 'dept', scan_status: 'COMPLETED' };
function issue(id: string, asset: string, group = 'dept'): ComplianceIssue { return { id, asset_id: asset, asset_name: asset, group_id: group, key: 'telnet', port: 23, protocol: 'tcp', version: 4, state: 'open', owner: '', first_seen: '2026-10-19', last_seen: '2026-10-19', deadline: null, breached: false, mttr_hours: null, policy_id: '', policy_version: '', sla_hours: 0, events: [{ action: 'observe', expected_version: 0, actor: 'actor', at: '2026-10-19', reason: 'source', observation: { finding_id: 'old', scan_id: 'oldscan', source_host: finding.host_ip, confidence: 'high', detection_method: 'service' } }] }; }
function snapshot(): ComplianceSnapshot { return { scope: { company_id: 'company', department_id: 'dept' }, groups: [{ id: 'dept', name: 'Department', parent_id: 'company' }], assets: [{ id: 'asset', group_id: 'dept', name: 'asset' }], findings: [finding], issues: [issue('one', 'asset')], owners: [{ id: 'owner', name: 'Operator' }], scorecards: [], scans: [], evidence: [], target_counts: {}, cutoff: '2026-10-19', rule_version: '1', mean_mttr_hours: null, completed_cycles: 0, limitations: [] }; }
function setup(data = snapshot(), write = true) {
  const errors: string[] = []; const pending: Promise<void>[] = [];
  const run = (work: () => Promise<void>) => { pending.push(work().catch(e => { errors.push(e.message); })); };
  const edit = vi.fn(); const evidence = vi.fn();
  render(<ComplianceWorkflow data={data} framework="ALL" canWrite={write} busy={false} run={run} edit={edit} evidence={evidence} />);
  return { errors, pending, edit, evidence };
}
describe('Guided compliance workflow', () => {
  beforeEach(() => vi.resetAllMocks());
  it('never matches history across groups and retains ambiguity', () => {
    const data = snapshot(); data.assets.push({ id: 'other', group_id: 'other-group', name: 'Other' }); data.issues.push(issue('two', 'other', 'other-group'));
    expect(assetCandidates(data, finding).map(a => a.id)).toEqual(['asset']);
    data.assets.push({ id: 'second', group_id: 'dept', name: 'Second' }); data.issues.push(issue('three', 'second'));
    expect(assetCandidates(data, finding)).toHaveLength(2);
    setup(data); expect(screen.getByText(/Identità ambigua/)).toBeTruthy(); expect((screen.getByLabelText('Asset per 192.0.2.1:23') as HTMLSelectElement).value).toBe('');
  });
  it('proposes a historical identity without making writes until confirmation', async () => {
    const { pending } = setup(); expect(complianceApi.observe).not.toHaveBeenCalled();
    fireEvent.click(screen.getByLabelText(/Confermo l’identità/)); fireEvent.submit(screen.getByRole('form', { name: 'Associa finding' })); await Promise.all(pending);
    expect(complianceApi.observe).toHaveBeenCalledWith('asset', 'finding');
  });
  it('preserves a created asset after failed association and retries without duplication', async () => {
    vi.mocked(complianceApi.asset).mockResolvedValue({ id: 'new-asset', name: 'New', group_id: 'dept' });
    vi.mocked(complianceApi.observe).mockRejectedValueOnce(new Error('Conflict')).mockResolvedValueOnce(undefined);
    const { pending, errors } = setup(); fireEvent.change(screen.getByLabelText('Asset per 192.0.2.1:23'), { target: { value: 'new' } });
    fireEvent.submit(screen.getByRole('form', { name: 'Associa finding' })); await Promise.all(pending);
    await screen.findByText(/Asset New registrato/); expect(errors).toEqual(['Conflict']);
    fireEvent.submit(screen.getByRole('form', { name: 'Associa finding' })); await Promise.all(pending);
    expect(complianceApi.asset).toHaveBeenCalledTimes(1); expect(complianceApi.observe).toHaveBeenLastCalledWith('new-asset', 'finding');
  });
  it('reports partial policy success and preserves optimistic versions', async () => {
    const data = snapshot(); data.issues.push(issue('two', 'second'));
    vi.mocked(complianceApi.command).mockResolvedValueOnce(undefined).mockRejectedValueOnce(new Error('409'));
    const { pending, errors } = setup(data);
    fireEvent.click(screen.getByLabelText('Includi asset · telnet nella policy')); fireEvent.click(screen.getByLabelText('Includi second · telnet nella policy'));
    for (const [label, value] of [['Responsabile delle attività', 'owner'], ['Ore della policy', '24'], ['Nome policy', 'team'], ['Versione della policy', 'v1'], ['Motivo dell’assegnazione', 'Approved policy']]) fireEvent.change(screen.getByLabelText(label), { target: { value } });
    fireEvent.submit(screen.getByRole('form', { name: 'Policy attività' })); await Promise.all(pending);
    expect(complianceApi.command).toHaveBeenCalledWith('one', { action: 'configure', expected_version: 4, owner: 'owner', sla_hours: 24, policy_id: 'team', policy_version: 'v1', reason: 'Approved policy' });
    await waitFor(() => expect(screen.getByText(/second · telnet: non salvato/)).toBeTruthy()); expect(errors[0]).toMatch(/solo in parte/);
  });
  it('keeps Auditor actions read-only and excludes negative port observations', () => {
    const data = snapshot(); data.findings.push({ ...finding, id: 'closed', port_state: 'closed', port: 99 }); setup(data, false);
    expect(screen.queryByRole('form')).toBeNull(); expect(screen.queryByText(/Conferma dispositivo.*99/)).toBeNull();
    expect(screen.getByRole('link', { name: /Apri scansioni/ }).getAttribute('href')).toBe('/scans');
    expect(complianceApi.command).not.toHaveBeenCalled();
  });
});
