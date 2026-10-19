import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { FindingRegulatoryContext } from './FindingRegulatoryContext';
import type { FindingItem } from '../types/finding';

const finding: FindingItem = { id: 'finding', scan_id: 'scan', tenant_id: 'tenant', host_ip: '192.0.2.1', port: 23, protocol: 'tcp', severity: 'HIGH', created_at: '2026-10-19' };
describe('Automatic regulatory context', () => {
  it('shows observed facts, meaning, actions and official references without forms', () => {
    const { container } = render(<FindingRegulatoryContext finding={{ ...finding, regulatory_context: { rule_version: 'v2', confidence: 'PROBABLE', detection_method: 'banner', limit: 'Not certification', implications: [{ kind: 'transport', observation: 'Telnet service', meaning: 'Encryption needs review', action: 'Use SSH', references: [{ id: 'NIS2 Art. 21(2)(h)', title: 'Cryptography', source: 'https://eur-lex.europa.eu/eli/dir/2022/2555/oj/eng' }] }] } }} />);
    expect(screen.getByText(/PROBABLE/)).toBeTruthy(); expect(screen.getByText(/Telnet service/)).toBeTruthy(); expect(screen.getByText(/Use SSH/)).toBeTruthy();
    expect(screen.getByRole('link', { name: /NIS2/ }).getAttribute('href')).toContain('eur-lex.europa.eu');
    expect(container.querySelectorAll('input,select,textarea,form')).toHaveLength(0);
  });
  it('does not treat an unmapped observation as compliance', () => {
    render(<FindingRegulatoryContext finding={{ ...finding, regulatory_context: { rule_version: 'v2', confidence: '', detection_method: '', implications: [], limit: 'No legal assessment' } }} />);
    expect(screen.getByText(/Questo non dimostra sicurezza o conformità/)).toBeTruthy(); expect(screen.queryByRole('link')).toBeNull();
  });
  it('does not fabricate mappings when the server has no context', () => {
    render(<FindingRegulatoryContext finding={finding} />); expect(screen.getByText(/non disponibile da questa versione/)).toBeTruthy(); expect(screen.queryByRole('link')).toBeNull();
  });
});
