import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { CompanyRoute } from './CompanyRoute';

const state = vi.hoisted(() => ({ selectedCompanyId: null as string | null, companies: [{ id: 'company' }], loading: false }));
vi.mock('../../context/CompanyContext', () => ({ useCompany: () => state }));

function mount(path: string) {
  render(<MemoryRouter initialEntries={[path]}><Routes>
    <Route path="/" element={<p>Dashboard unica</p>} />
    <Route path="/ssh-servers" element={<p>Tool SSH indipendente</p>} />
    <Route element={<CompanyRoute />}><Route path="/targets" element={<p>Target aziendali</p>} /></Route>
  </Routes></MemoryRouter>);
}
describe('Company navigation boundary', () => {
  beforeEach(() => { state.selectedCompanyId = null; state.loading = false; });
  it('blocks a direct operational URL without a company', () => { mount('/targets'); expect(screen.getByText('Dashboard unica')).toBeDefined(); expect(screen.queryByText('Target aziendali')).toBeNull(); });
  it('blocks stale company IDs', () => { state.selectedCompanyId = 'removed'; mount('/targets'); expect(screen.getByText('Dashboard unica')).toBeDefined(); });
  it('permits the selected company', () => { state.selectedCompanyId = 'company'; mount('/targets'); expect(screen.getByText('Target aziendali')).toBeDefined(); });
  it('waits for company resolution', () => { state.loading = true; mount('/targets'); expect(screen.getByRole('status')).toBeDefined(); expect(screen.queryByText('Target aziendali')).toBeNull(); });
  it('always permits independent SSH tools', () => { mount('/ssh-servers'); expect(screen.getByText('Tool SSH indipendente')).toBeDefined(); });
});
