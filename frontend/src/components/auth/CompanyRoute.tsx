import { Navigate, Outlet } from 'react-router-dom';
import { useCompany } from '../../context/CompanyContext';

// Navigation policy only: server-side tenant authorization remains independent.
export function CompanyRoute() {
  const { selectedCompanyId, companies, loading } = useCompany();
  const validCompany = selectedCompanyId && companies.some(c => c.id === selectedCompanyId);
  // Keep page state during a background inventory refresh. Initial/unresolved
  // selections still wait; server-side tenant authorization is unchanged.
  if (loading && !validCompany) return <p role="status" className="p-6 text-slate-400">Caricamento azienda...</p>;
  if (!validCompany) return <Navigate to="/" replace />;
  return <Outlet />;
}
