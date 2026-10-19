import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { act, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { Compliance } from './Compliance';
import { complianceApi } from '../api/compliance';
import type { ComplianceSnapshot } from '../types/compliance';

const state = vi.hoisted(() => ({ company: 'company-one', role: 'SecOps' }));
vi.mock('../context/CompanyContext', () => ({ useCompany: () => ({ selectedCompanyId: state.company, selectedCompany: { name: 'Test Company' }, departments: [{ id: 'department-one', name: 'Department One' }] }) }));
vi.mock('../context/AuthContext', () => ({ useAuth: () => ({ user: { id: 'actor', role: state.role } }) }));
vi.mock('../api/compliance', () => ({ complianceApi: { snapshot: vi.fn(), asset: vi.fn(), observe: vi.fn(), command: vi.fn(), evidence: vi.fn() } }));
vi.mock('../components/AuditExport', () => ({ AuditExport: () => <div>Audit export</div> }));

function snapshot(): ComplianceSnapshot { return { owners: [], target_counts: {}, cutoff: '2026-10-19T10:00:00Z', rule_version: '1', scope: { company_id: state.company, department_id: '' }, groups: [{ id: state.company, name: 'Test Company', parent_id: null }], assets: [], findings: [], scans: [], issues: [], evidence: [], mean_mttr_hours: null, completed_cycles: 0, limitations: [], scorecards: [{ group_id: state.company, group_name: 'Test Company', framework: 'ALL', score: 0, assessed: 0, applicable: 1, unknown: 1, excluded: 0, controls: [{ id: 'ISO-A.8.8', framework: 'ISO27001', title: 'Vulnerability management', source: 'https://www.iso.org/standard/27001', mapping: 'vulnerability', status: 'unknown', finding_ids: [], issue_ids: [] }] }] }; }
describe('Compliance evidence and scope', () => {
  afterEach(() => vi.useRealTimers());
  beforeEach(() => { vi.resetAllMocks(); state.company = 'company-one'; state.role = 'SecOps'; vi.mocked(complianceApi.snapshot).mockResolvedValue(snapshot()); });
  it('shows unknown coverage without claiming absence is compliance', async () => {
    render(<Compliance />); await screen.findByRole('cell', { name: 'Non valutato' });
    expect(screen.getByText('Scorecard ALL: 0.0%')).toBeTruthy();
    expect(screen.getByText(/Nessuna issue correlata/)).toBeTruthy();
    expect(complianceApi.snapshot).toHaveBeenCalledWith('company-one', '', expect.any(AbortSignal));
  });
  it('keeps Auditor read-only', async () => {
    state.role = 'Auditor'; render(<Compliance />); await screen.findByRole('cell', { name: 'Non valutato' });
    expect(screen.queryByText('Registra evidenza di controllo')).toBeNull();
    expect(screen.queryByText('Identità asset e osservazioni')).toBeNull();
    expect(screen.getByText('Audit export')).toBeTruthy();
  });
  it('refreshes automatically but pauses while a panel is being edited', async () => {
    vi.useFakeTimers();
    render(<Compliance />); await act(async () => {});
    await act(async () => { vi.advanceTimersByTime(30000); });
    expect(complianceApi.snapshot).toHaveBeenCalledTimes(2);
    fireEvent.click(screen.getByText('Registra evidenza di controllo'));
    fireEvent.change(screen.getByLabelText('Motivazione'), { target: { value: 'Unsaved evidence' } });
    await act(async () => { vi.advanceTimersByTime(60000); });
    expect(complianceApi.snapshot).toHaveBeenCalledTimes(2);
    expect((screen.getByLabelText('Motivazione') as HTMLTextAreaElement).value).toBe('Unsaved evidence');
  });
  it('opens the correct evidence form from the generated activity', async () => {
    render(<Compliance />); fireEvent.click(await screen.findByRole('button', { name: 'Valuta ISO-A.8.8' }));
    expect((screen.getByLabelText('Controllo') as HTMLSelectElement).value).toBe('ISO-A.8.8');
    expect(document.activeElement).toBe(screen.getByLabelText('Valutazione'));
  });
  it('sends source version, rationale and exact evidence scope', async () => {
    render(<Compliance />); await screen.findByRole('cell', { name: 'Non valutato' });
    fireEvent.click(screen.getByText('Registra evidenza di controllo'));
    fireEvent.change(screen.getByLabelText('Motivazione'), { target: { value: 'Reviewed documents' } });
    fireEvent.change(screen.getByLabelText('Riferimento documentale'), { target: { value: 'audit/document' } });
    fireEvent.change(screen.getByLabelText('Versione della fonte'), { target: { value: 'v2' } });
    fireEvent.click(screen.getByRole('button', { name: 'Registra evidenza' }));
    await waitFor(() => expect(complianceApi.evidence).toHaveBeenCalledWith({ group_id: 'company-one', control_id: 'ISO-A.8.8', status: 'unknown', reason: 'Reviewed documents', reference: 'audit/document', source_version: 'v2' }));
  });
  it('reports errors and allows a fresh retry', async () => {
    vi.mocked(complianceApi.snapshot).mockRejectedValueOnce(new Error('Scope unavailable'));
    render(<Compliance />); expect((await screen.findByRole('alert')).textContent).toContain('Scope unavailable');
    fireEvent.click(screen.getByRole('button', { name: 'Aggiorna evidenze' }));
    await screen.findByRole('cell', { name: 'Non valutato' }); expect(screen.queryByRole('alert')).toBeNull();
  });
  it('discards a late response after company change', async () => {
    let resolveOld!: (value: ComplianceSnapshot) => void;
    vi.mocked(complianceApi.snapshot).mockImplementationOnce(() => new Promise(resolve => { resolveOld = resolve; }));
    const old = snapshot(); const { rerender } = render(<Compliance />);
    state.company = 'company-two'; const newer = snapshot(); newer.rule_version = 'new-scope'; vi.mocked(complianceApi.snapshot).mockResolvedValueOnce(newer);
    rerender(<Compliance />); await screen.findByText(/new-scope/); resolveOld(old);
    await waitFor(() => expect(screen.getByText(/new-scope/)).toBeTruthy());
  });
  it('preserves the configured policy and submits the current optimistic version', async () => {
    const data = snapshot(); data.owners = [{ id: 'actor', name: 'Test Operator' }];
    data.issues = [{ id: 'issue-one', asset_id: 'asset-one', asset_name: 'Verified Asset', group_id: state.company, key: 'CVE-2026-1234', port: 443, protocol: 'tcp', version: 3, state: 'open', owner: 'actor', first_seen: data.cutoff, last_seen: data.cutoff, deadline: data.cutoff, breached: true, mttr_hours: null, policy_id: 'existing-policy', policy_version: 'v3', sla_hours: 72, events: [] }];
    vi.mocked(complianceApi.snapshot).mockResolvedValue(data); render(<Compliance />);
    await screen.findByText(/SLA SCADUTO/); fireEvent.click(screen.getByRole('button', { name: 'Aggiorna issue issue-on' }));
    expect((screen.getByLabelText('ID policy') as HTMLInputElement).value).toBe('existing-policy');
    expect((screen.getByLabelText('Ore SLA (0 rimuove la policy)') as HTMLInputElement).value).toBe('72');
    fireEvent.change(within(screen.getByRole('form', { name: 'Aggiorna remediation' })).getByLabelText('Motivazione'), { target: { value: 'Confirm documented policy' } });
    fireEvent.click(screen.getByRole('button', { name: 'Registra aggiornamento' }));
    await waitFor(() => expect(complianceApi.command).toHaveBeenCalledWith('issue-one', { action: 'configure', expected_version: 3, reason: 'Confirm documented policy', owner: 'actor', policy_id: 'existing-policy', policy_version: 'v3', sla_hours: 72 }));
  });
});
