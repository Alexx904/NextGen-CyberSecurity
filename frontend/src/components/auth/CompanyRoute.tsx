import { Navigate, Outlet } from 'react-router-dom';
import { useCompany } from '../../context/CompanyContext';

// Navigation policy only: server-side tenant authorization remains independent.
export function CompanyRoute() {
  const { selectedCompanyId, companies, loading } = useCompany();
  if (loading) return <p role="status" className="p-6 text-slate-400">Caricamento azienda...</p>;
  if (!selectedCompanyId || !companies.some(c => c.id === selectedCompanyId)) return <Navigate to="/" replace />;
  return <Outlet />;
}
