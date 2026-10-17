import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { GroupModal } from './GroupModal';
import type { TargetGroupNode } from '../types/target';

describe('Company -> department editor', () => {
  it('only allows the selected company as parent, never a department', () => {
    const groups = [
      { id: 'company', name: 'Company', parent_id: null },
      { id: 'department', name: 'IT', parent_id: 'company' },
      { id: 'other', name: 'Other Company', parent_id: null },
    ] as TargetGroupNode[];
    render(<GroupModal isOpen onClose={() => {}} onGroupSaved={() => {}} existingGroups={groups} companyId="company" defaultParentId="company" />);
    expect((screen.getByLabelText('Azienda del reparto') as HTMLSelectElement).disabled).toBe(true);
    expect(screen.getAllByRole('option')).toHaveLength(1);
    expect(screen.queryByRole('option', { name: /IT/ })).toBeNull();
    expect(screen.queryByRole('option', { name: /Other Company/ })).toBeNull();
  });
});
