import { beforeEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import App from './App';
import { apiClient } from './api/client';

const session = vi.hoisted(() => ({ authenticated: false }));
vi.mock('./api/client', () => ({ apiClient: { get: vi.fn(), post: vi.fn(), put: vi.fn(), delete: vi.fn() } }));
vi.mock('./context/AuthContext', () => ({
  AuthProvider: ({ children }: { children: ReactNode }) => children,
  useAuth: () => ({ user: null, isAuthenticated: session.authenticated, isLoading: false }),
}));
vi.mock('./pages/Login', async () => {
  const { useNavigate } = await import('react-router-dom');
  return { Login: () => {
    const navigate = useNavigate();
    return <button onClick={() => { session.authenticated = true; navigate('/'); }}>Test sign in</button>;
  } };
});
vi.mock('./pages/Dashboard', () => ({ Dashboard: () => <p>Dashboard fixture</p> }));
vi.mock('./components/layout/AppLayout', async () => {
  const { Outlet } = await import('react-router-dom');
  const { useCompany } = await import('./context/CompanyContext');
  return { AppLayout: () => {
    const { companies, loading } = useCompany();
    return <><p>{loading ? 'Loading companies' : companies.map(company => company.name).join(', ')}</p><Outlet /></>;
  } };
});
vi.mock('./context/DebugContext', () => ({ DebugProvider: ({ children }: { children: ReactNode }) => children }));
vi.mock('./components/debug/DebugFloatingToggle', () => ({ DebugFloatingToggle: () => null }));
vi.mock('./components/debug/DebugConsoleDrawer', () => ({ DebugConsoleDrawer: () => null }));

beforeEach(() => {
  vi.clearAllMocks(); session.authenticated = false; localStorage.clear();
  window.history.replaceState(null, '', '/login');
  vi.mocked(apiClient.get).mockImplementation(async (url: string) => {
    if (!session.authenticated) throw new Error('Unauthorized test session');
    return { data: { success: true, data: url === '/target-groups'
      ? [{ id: 'company', name: 'Recovered Company', parent_id: null, children: [] }]
      : { items: [], total: 0 } } } as Awaited<ReturnType<typeof apiClient.get>>;
  });
});

describe('company context login lifecycle', () => {
  it('loads companies after login without a browser refresh or unauthenticated inventory calls', async () => {
    render(<App />);
    const signIn = await screen.findByRole('button', { name: 'Test sign in' });
    expect(apiClient.get).not.toHaveBeenCalled();
    fireEvent.click(signIn);
    expect(await screen.findByText('Recovered Company')).toBeDefined();
    await waitFor(() => expect(screen.getByText('Dashboard fixture')).toBeDefined());
    expect(apiClient.get).toHaveBeenCalledWith('/target-groups');
  });
});
