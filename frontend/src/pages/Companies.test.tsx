import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { Companies } from './Companies';
import * as CompanyContextModule from '../context/CompanyContext';

vi.mock('../context/CompanyContext', async () => {
  const actual = await vi.importActual<any>('../context/CompanyContext');
  return {
    ...actual,
    useCompany: vi.fn(),
  };
});

const mockCompanies = [
  {
    id: 'c-1',
    name: 'Acme Security Corp',
    tenant_id: 't-1',
    parent_id: null,
    color_tag: '#3b82f6',
    description: 'Azienda principale perimetro cloud',
    direct_target_count: 2,
    target_count: 5,
    aggregate_risk: 'CRITICAL',
    severity_counts: { CRITICAL: 1 },
    created_at: '2026-10-10T10:00:00Z',
    updated_at: '2026-10-10T10:00:00Z',
    children: [
      {
        id: 'd-1',
        name: 'Reparto IT',
        tenant_id: 't-1',
        parent_id: 'c-1',
        color_tag: '#10b981',
        children: [],
      },
    ],
  },
];

describe('Companies Page', () => {
  const mockSetSelected = vi.fn();
  const mockCreateCompany = vi.fn();
  const mockUpdateCompany = vi.fn();
  const mockDeleteCompany = vi.fn();
  const mockRefresh = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    (CompanyContextModule.useCompany as any).mockReturnValue({
      companies: mockCompanies,
      selectedCompanyId: null,
      setSelectedCompanyId: mockSetSelected,
      createCompany: mockCreateCompany,
      updateCompany: mockUpdateCompany,
      deleteCompany: mockDeleteCompany,
      refreshCompanies: mockRefresh,
      loading: false,
      error: null,
    });
  });

  it('renders companies dashboard and stats', () => {
    render(
      <BrowserRouter>
        <Companies />
      </BrowserRouter>
    );

    expect(screen.getByText('Gestione Aziende & Scope')).toBeTruthy();
    expect(screen.getByText('Acme Security Corp')).toBeTruthy();
    expect(screen.getByText('Azienda principale perimetro cloud')).toBeTruthy();
    expect(screen.getByText('CRITICAL')).toBeTruthy();
  });

  it('selects company when clicking activate button', () => {
    render(
      <BrowserRouter>
        <Companies />
      </BrowserRouter>
    );

    const activateBtn = screen.getByText('Seleziona come Attiva');
    fireEvent.click(activateBtn);

    expect(mockSetSelected).toHaveBeenCalledWith('c-1');
  });

  it('filters companies with search input', () => {
    render(
      <BrowserRouter>
        <Companies />
      </BrowserRouter>
    );

    const searchInput = screen.getByPlaceholderText(/Cerca azienda per nome o descrizione/i);
    fireEvent.change(searchInput, { target: { value: 'Inesistente' } });

    expect(screen.getByText('Nessuna Azienda Trovata')).toBeTruthy();
  });
});
