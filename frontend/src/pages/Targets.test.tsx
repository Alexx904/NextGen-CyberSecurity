import { fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { expect, it, vi } from 'vitest';
import { Targets } from './Targets';
import { useCompany } from '../context/CompanyContext';
import type { CompanyContextType } from '../context/CompanyContext';

vi.mock('../context/CompanyContext', () => ({ useCompany: vi.fn() }));
vi.mock('../api/client', () => ({ apiClient: { get: vi.fn(async () => ({ data: { success: true, data: { items: [{ id: 'target', value: '192.0.2.7', type: 'IP', group_id: 'department', is_blacklisted: false, created_at: '' }] } } })) } }));
vi.mock('../components/TargetModal', () => ({ TargetModal: () => null }));
vi.mock('../components/TreeExplorer', () => ({
  getGroupAndDescendantIds: (node: { id: string }) => [node.id],
  TreeExplorer: ({ onSelectGroup }: { onSelectGroup: (id: string, node: unknown) => void }) => <button onClick={() => onSelectGroup('department', { id: 'department', name: 'Department A', children: [] })}>Select department fixture</button>,
}));

it('clears the old department and bulk selection when switching company', async () => {
  const company = { selectedCompanyId: 'company-a', selectedCompany: { id: 'company-a', name: 'Company A' }, companyGroupIds: new Set(['company-a', 'department']), refreshCompanies: vi.fn(), setSelectedCompanyId: vi.fn() } as unknown as CompanyContextType;
  vi.mocked(useCompany).mockReturnValue(company);
  const view = render(<MemoryRouter><Targets /></MemoryRouter>);
  await screen.findByText('192.0.2.7');
  fireEvent.click(screen.getByRole('button', { name: 'Select department fixture' }));
  fireEvent.click(screen.getByTitle('Seleziona tutti i target filtrati'));
  expect(screen.getByText(/bersaglio\/i selezionato\/i/)).toBeDefined();
  vi.mocked(useCompany).mockReturnValue({ ...company, selectedCompanyId: 'company-b', selectedCompany: { ...company.selectedCompany!, id: 'company-b', name: 'Company B' }, companyGroupIds: new Set(['company-b']) });
  view.rerender(<MemoryRouter><Targets /></MemoryRouter>);
  expect(screen.queryByText('192.0.2.7')).toBeNull();
  expect(screen.queryByText('Department A')).toBeNull();
  expect(screen.queryByText(/bersaglio\/i selezionato\/i/)).toBeNull();
});
